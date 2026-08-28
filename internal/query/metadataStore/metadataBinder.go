package metadataStore

import (
	"fmt"
	"strings"

	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

// Paths are persisted to reduce short lived objects in the query pipeline
var pathStore = map[string]mdl.ResolvedPath{}

type MetadataBinder struct { // USE INTERFACE
	accessPolicy         mdl.AccessPolicy
	rootResourceMetadata mdl.TableMetadata
	rootAccessPolicy     mdl.TableAccessPolicy

	pathIdBuilder *strings.Builder
}

func NewMetadataBinder(
	rootMetadata mdl.TableMetadata,
	accessPolicy mdl.AccessPolicy,
) (*MetadataBinder, error) {
	if rootMetadata == nil {
		return nil, qerr.InternalErr("unable to create new MetadataBinder: rootMetadata cannot be nil")
	}

	if accessPolicy == nil {
		return nil, qerr.InternalErr("unable to create new MetadataBinder: accessPolicy cannot be nil")
	}

	rootAccessPolicy, err := getTableAccessPolicy(accessPolicy, rootMetadata)
	if err != nil {
		return nil, qerr.InternalErr("unable to create new MetadataBinder: %w", err)
	}

	return &MetadataBinder{
		accessPolicy:         accessPolicy,
		rootResourceMetadata: rootMetadata,
		rootAccessPolicy:     rootAccessPolicy,
	}, nil
}

func (b *MetadataBinder) ResolveColumn(columnPath string) (mdl.ResolvedColumn, error) {
	pathSegments := strings.Split(columnPath, "/")
	pathLen := len(pathSegments)
	if pathLen == 0 {
		return mdl.ResolvedColumn{}, qerr.SyntaxErr(
			"unable to resolve column: invalid column path: '%s'", columnPath,
		)
	}

	columnName := pathSegments[pathLen-1]
	pathToColumn := pathSegments[0 : pathLen-1]

	resolvedPath, err := b.resolvePath(pathToColumn, b.rootResourceMetadata)
	if err != nil {
		return mdl.ResolvedColumn{}, err
	}

	resource := resolvedPath.EndResource
	columnMetadata := resource.GetColumnMetadata(columnName)
	if columnMetadata == nil {
		return mdl.ResolvedColumn{}, qerr.BindingErr(
			"table %s does not include a column definition for %s",
			resource.Name(),
			columnName,
		)
	}

	accessPolicy, err := getTableAccessPolicy(
		b.accessPolicy,
		resource,
	)
	if err != nil {
		return mdl.ResolvedColumn{}, qerr.InternalErr("unable to resolve column: %w", err)
	}

	if err := validateColumnAccess(
		accessPolicy,
		resolvedPath.EndResource,
		columnMetadata,
	); err != nil {
		return mdl.ResolvedColumn{}, qerr.InternalErr("unable to resolve column: %w", err)
	}

	return mdl.ResolvedColumn{
		ResolvedPath: resolvedPath,
		Metadata:     columnMetadata,
	}, nil
}

func (b *MetadataBinder) ResolvePath(pathString string) (mdl.ResolvedPath, error) {
	pathSegments := strings.Split(pathString, "/")
	pathLen := len(pathSegments)
	if pathLen == 0 {
		return mdl.ResolvedPath{}, qerr.SyntaxErr("unable to resolve path: '%s'", pathString)
	}

	resolvedPath, err := b.resolvePath(pathSegments, b.rootResourceMetadata)
	if err != nil {
		return mdl.ResolvedPath{}, err
	}

	return resolvedPath, nil
}

func (b *MetadataBinder) ResolveRelationship(resource mdl.TableMetadata, relationshipName string) (mdl.TraversalStep, error) {
	if resource == nil {
		return mdl.TraversalStep{}, qerr.InternalErr("unable to get table access policy: resource cannot be nil")
	}

	relationshipData := resource.GetRelationshipMetadata(relationshipName)
	if relationshipData == nil {
		return mdl.TraversalStep{}, qerr.BindingErr(
			"table %s does not include a relationship definition for %s",
			resource.Name(),
			relationshipName,
		)
	}

	step := mdl.TraversalStep{
		SubPathId:    appendToPath(resource.Name(), relationshipData),
		Relationship: relationshipData,
	}

	if err := b.validateTraversalPermissions(step); err != nil {
		return mdl.TraversalStep{}, err
	}

	return step, nil
}

func (b *MetadataBinder) resolvePath(
	pathSegments []string,
	rootResource mdl.TableMetadata,
) (mdl.ResolvedPath, error) {
	// final pathId
	finalPathId := fmt.Sprintf(
		"%s/%s",
		rootResource.Name(),
		strings.Join(pathSegments, "/"),
	)

	// try and retrieve from store
	if path, exists := pathStore[finalPathId]; exists {
		return path, nil
	}

	// prepare path id builder
	if b.pathIdBuilder == nil {
		b.pathIdBuilder = &strings.Builder{}
	} else {
		b.pathIdBuilder.Reset()
	}

	// all path ids start with the resource name
	b.pathIdBuilder.WriteString(rootResource.Name())

	// initialise resolved path
	o := mdl.ResolvedPath{
		StartResource: rootResource,
		EndResource:   rootResource,
	}

	// if path has no length return early
	traversalPathLength := len(pathSegments)
	if traversalPathLength == 0 {
		o.Id = b.pathIdBuilder.String()
		return o, nil
	}

	// Traverse through intermediate steps
	i := 0
	o.Steps = make([]*mdl.TraversalStep, traversalPathLength)
	for i < traversalPathLength {
		// Get relationship data
		relationshipName := pathSegments[i]
		relationshipData := o.EndResource.GetRelationshipMetadata(relationshipName)
		if relationshipData == nil {
			return o, qerr.BindingErr(
				"%s does not include a relationship definition for '%s'",
				o.EndResource.FullyQualifiedName(),
				relationshipName,
			)
		}

		// validate relationship type
		relationshipType := relationshipData.Type()
		isIntermediateStep := i < traversalPathLength-1
		if isIntermediateStep && relationshipType == mdl.RelationshipOneToMany {
			return o, qerr.BindingErr(
				"%s is a 1:N relationship and cannot be used as an "+
					"intermediate path step, please use a collection operator",
				relationshipData.ColumnName(),
			)
		}

		// update the path id
		appendToPathSB(b.pathIdBuilder, relationshipData)

		// append the traversal step to the resolved path
		o.Steps[i] = &mdl.TraversalStep{
			SubPathId:    b.pathIdBuilder.String(),
			Relationship: relationshipData,
		}

		// validate user has permissions to make the traversal
		err := b.validateTraversalPermissions(*o.Steps[i])
		if err != nil {
			return o, err
		}

		// update the final end resource and relationship type
		o.EndResource = relationshipData.To()
		o.Type = relationshipType
		i++
	}

	// Set the final pathid and return
	o.Id = b.pathIdBuilder.String()
	if o.Id != finalPathId {
		return o, qerr.InternalErr(
			"unable to resolve path. There is a disconnect between the " +
				"projected final path id and the built final path id",
		)
	}

	pathStore[o.Id] = o
	return o, nil
}

func (b *MetadataBinder) validateTraversalPermissions(step mdl.TraversalStep) error {
	fromTableAccessPolicy, err := getTableAccessPolicy(b.accessPolicy, step.Relationship.From())
	if err != nil {
		return err
	}

	err = validateColumnAccess(
		fromTableAccessPolicy,
		step.Relationship.From(),
		step.Relationship.FromColumn(),
	)
	if err != nil {
		return err
	}

	toTableAccessPolicy, err := getTableAccessPolicy(b.accessPolicy, step.Relationship.To())
	if err != nil {
		return err
	}
	err = validateColumnAccess(
		toTableAccessPolicy,
		step.Relationship.To(),
		step.Relationship.ToColumn(),
	)
	if err != nil {
		return err
	}

	return nil
}

func validateColumnAccess(
	tableAccessPolicy mdl.TableAccessPolicy,
	tableData mdl.TableMetadata,
	columnData mdl.ColumnMetadata,
) error {
	columnAccessPolicy := tableAccessPolicy.GetColumnAccessPolicy(columnData.Name())
	if columnAccessPolicy == nil {
		return qerr.InternalErr(
			"unable to find access policy for %s.%s",
			tableData.Name(),
			columnData.Name(),
		)
	}

	if !columnAccessPolicy.CanAccess() {
		return qerr.AccessErr(
			"you do not have permission to access the %s column on the %s table",
			columnData.Name(),
			tableData.Name(),
		)
	}
	return nil
}

func getTableAccessPolicy(
	accessPolicy mdl.AccessPolicy,
	tableData mdl.TableMetadata,
) (mdl.TableAccessPolicy, error) {
	if accessPolicy == nil {
		return nil, qerr.InternalErr("access policy cannot be nil")
	}

	tablePolicy := accessPolicy.GetTableAccessPolicy(tableData.Name())
	if tablePolicy == nil {
		return nil, qerr.InternalErr("unable to find access policy for the %s table", tableData.Name())
	}

	if !tablePolicy.CanAccess() {
		return nil, qerr.AccessErr("you do not have permission to access the %s table", tableData.Name())
	}
	return tablePolicy, nil
}

/*
DO WE NEED THESE????
*/

func appendToPath(path string, relationship mdl.RelationshipMetadata) string {
	return fmt.Sprintf("%s/%s", path, relationship.ColumnName())
}

func appendToPathSB(sb *strings.Builder, relationship mdl.RelationshipMetadata) {
	sb.WriteByte('/')
	sb.WriteString(relationship.ColumnName())
}
