package query

import (
	"fmt"
	"strings"
)

type ResolvedPath struct {
	Id            string
	StartResource TableMetadata
	Steps         []*TraversalStep
	EndResource   TableMetadata
	Type          RelationshipType
}

type TraversalStep struct {
	SubPathId    string
	Relationship RelationshipMetadata
}

// Paths are persisted to reduce short lived objects in the query pipeline
var pathStore = map[string]ResolvedPath{}

type metadataBinderNew struct {
	accessPolicy         AccessPolicy
	rootResourceMetadata TableMetadata
	rootAccessPolicy     TableAccessPolicy

	pathIdBuilder *strings.Builder
}

// All columns, for select and order by will be represented by this type
type ResolvedColumn struct {
	ResolvedPath ResolvedPath
	Metadata     ColumnMetadata
}

func NewMetadataBinder(
	rootMetadata TableMetadata,
	accessPolicy AccessPolicy,
) (*metadataBinderNew, error) {
	if rootMetadata == nil {
		return nil, internalErr("unable to create new MetadataBinder: rootMetadata cannot be nil")
	}

	if accessPolicy == nil {
		return nil, internalErr("unable to create new MetadataBinder: accessPolicy cannot be nil")
	}

	rootAccessPolicy, err := getTableAccessPolicy(accessPolicy, rootMetadata)
	if err != nil {
		return nil, internalErr("unable to create new MetadataBinder: %w", err)
	}

	return &metadataBinderNew{
		accessPolicy:         accessPolicy,
		rootResourceMetadata: rootMetadata,
		rootAccessPolicy:     rootAccessPolicy,
	}, nil
}

func (b *metadataBinderNew) ResolveColumn(columnPath string) (ResolvedColumn, error) {
	pathSegments := strings.Split(columnPath, "/")
	pathLen := len(pathSegments)
	if pathLen == 0 {
		return ResolvedColumn{}, syntaxErr(
			"unable to resolve column: invalid column path: '%s'", columnPath,
		)
	}

	columnName := pathSegments[pathLen-1]
	pathToColumn := pathSegments[0 : pathLen-1]

	resolvedPath, err := b.resolvePath(pathToColumn, b.rootResourceMetadata)
	if err != nil {
		return ResolvedColumn{}, err
	}

	resource := resolvedPath.EndResource
	columnMetadata := resource.GetColumnMetadata(columnName)
	if columnMetadata == nil {
		return ResolvedColumn{}, bindingErr(
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
		return ResolvedColumn{}, internalErr("unable to resolve column: %w", err)
	}

	if err := validateColumnAccess(
		accessPolicy,
		resolvedPath.EndResource,
		columnMetadata,
	); err != nil {
		return ResolvedColumn{}, internalErr("unable to resolve column: %w", err)
	}

	return ResolvedColumn{
		ResolvedPath: resolvedPath,
		Metadata:     columnMetadata,
	}, nil
}

func (b *metadataBinderNew) ResolvePath(pathString string) (ResolvedPath, error) {
	pathSegments := strings.Split(pathString, "/")
	pathLen := len(pathSegments)
	if pathLen == 0 {
		return ResolvedPath{}, syntaxErr("unable to resolve path: '%s'", pathString)
	}

	resolvedPath, err := b.resolvePath(pathSegments, b.rootResourceMetadata)
	if err != nil {
		return ResolvedPath{}, err
	}

	return resolvedPath, nil
}

func (b *metadataBinderNew) resolveRelationship(resource TableMetadata, relationshipName string) (TraversalStep, error) {
	if resource == nil {
		return TraversalStep{}, internalErr("unable to get table access policy: resource cannot be nil")
	}

	relationshipData := resource.GetRelationshipMetadata(relationshipName)
	if relationshipData == nil {
		return TraversalStep{}, bindingErr(
			"table %s does not include a relationship definition for %s",
			resource.Name(),
			relationshipName,
		)
	}

	step := TraversalStep{
		SubPathId:    appendToPath(resource.Name(), relationshipData),
		Relationship: relationshipData,
	}

	if err := b.validateTraversalPermissions(step); err != nil {
		return TraversalStep{}, err
	}

	return step, nil
}

func (b *metadataBinderNew) resolvePath(
	pathSegments []string,
	rootResource TableMetadata,
) (ResolvedPath, error) {
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
	o := ResolvedPath{
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
	o.Steps = make([]*TraversalStep, traversalPathLength)
	for i < traversalPathLength {
		// Get relationship data
		relationshipName := pathSegments[i]
		relationshipData := o.EndResource.GetRelationshipMetadata(relationshipName)
		if relationshipData == nil {
			return o, bindingErr(
				"%s does not include a relationship definition for '%s'",
				o.EndResource.FullyQualifiedName(),
				relationshipName,
			)
		}

		// validate relationship type
		relationshipType := relationshipData.Type()
		isIntermediateStep := i < traversalPathLength-1
		if isIntermediateStep && relationshipType == RelationshipOneToMany {
			return o, bindingErr(
				"%s is a 1:N relationship and cannot be used as an "+
					"intermediate path step, please use a collection operator",
				relationshipData.Id(),
			)
		}

		// append the traversal step to the resolved path
		o.Steps[i] = &TraversalStep{
			SubPathId:    b.pathIdBuilder.String(),
			Relationship: relationshipData,
		}

		// validate user has permissions to make the traversal
		err := b.validateTraversalPermissions(*o.Steps[i])
		if err != nil {
			return o, err
		}

		// update the path id
		appendToPathSB(b.pathIdBuilder, relationshipData)

		// update the final end resource and relationship type
		o.EndResource = relationshipData.To()
		o.Type = relationshipType
		i++
	}

	// Set the final pathid and return
	o.Id = b.pathIdBuilder.String()
	if o.Id != finalPathId {
		return o, internalErr(
			"unable to resolve path. There is a disconnect between the " +
				"projected final path id and the built final path id",
		)
	}

	pathStore[o.Id] = o
	return o, nil
}

func (b *metadataBinderNew) validateTraversalPermissions(step TraversalStep) error {
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
	tableAccessPolicy TableAccessPolicy,
	tableData TableMetadata,
	columnData ColumnMetadata,
) error {
	columnAccessPolicy := tableAccessPolicy.GetColumnAccessPolicy(columnData.Name())
	if columnAccessPolicy == nil {
		return internalErr(
			"unable to find access policy for %s.%s",
			tableData.Name(),
			columnData.Name(),
		)
	}

	if !columnAccessPolicy.CanAccess() {
		return accessErr(
			"you do not have permission to access the %s column on the %s table",
			columnData.Name(),
			tableData.Name(),
		)
	}
	return nil
}

func getTableAccessPolicy(accessPolicy AccessPolicy, tableData TableMetadata) (TableAccessPolicy, error) {
	if accessPolicy == nil {
		return nil, internalErr("access policy cannot be nil")
	}

	tablePolicy := accessPolicy.GetTableAccessPolicy(tableData.Name())
	if tablePolicy == nil {
		return nil, internalErr("unable to find access policy for the %s table", tableData.Name())
	}

	if !tablePolicy.CanAccess() {
		return nil, accessErr("you do not have permission to access the %s table", tableData.Name())
	}
	return tablePolicy, nil
}

/*
DO WE NEED THESE????
*/

func appendToPath(path string, relationship RelationshipMetadata) string {
	return fmt.Sprintf("%s/%s", path, relationship.RelationshipColumn())
}

func appendToPathSB(sb *strings.Builder, relationship RelationshipMetadata) {
	sb.WriteByte('/')
	sb.WriteString(relationship.RelationshipColumn())
}
