package query

import (
	"strings"
)

var resolvedPathStore = map[string]*ResolvedPath{}

type ResolvedPath struct {
	Id            string
	StartResource TableMetadata
	Steps         []TraversalStep
	EndResource   TableMetadata
	Type          RelationshipType
}

type TraversalStep struct {
	From         TableMetadata
	Relationship RelationshipMetadata
	To           TableMetadata
}

type metadataBinder struct {
	maximumDepth int
	accessPolicy AccessPolicy
}

func (b metadataBinder) bindMetadata(
	rootResource TableMetadata,
	depth int,
	operations []QueryOperation,
) error {
	if b.accessPolicy == nil {
		return internalErr("access policy cannot be nil")
	}

	if depth > b.maximumDepth {
		return syntaxErr("maximum expand depth (%d) exceeded", b.maximumDepth)
	}

	for _, op := range operations {
		switch op := op.(type) {
		case *SelectOperation:
			return b.bindSelectOperation(rootResource, op)
		case *ExpandOperation:
			return b.bindExpandOperation(rootResource, depth, op)
		default:
			return internalErr("unexpected operation received")
		}
	}
	return nil
}

func (b metadataBinder) bindSelectOperation(
	rootResource TableMetadata,
	op *SelectOperation,
) error {
	rootResourceAccessPolicy, err := b.getTableAccessPolicy(rootResource)
	if err != nil {
		return err
	}

	for _, s := range op.Columns {
		if s.ColumnData != nil {
			continue
		}

		columnData, err := b.resolveColumn(rootResource, s.ColumnName)
		if err != nil {
			return err
		}
		s.ColumnData = columnData

		err = validateColumnAccess(
			rootResourceAccessPolicy,
			rootResource,
			columnData,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b metadataBinder) bindExpandOperation(
	rootResource TableMetadata,
	depth int,
	op *ExpandOperation,
) error {
	for _, e := range op.Expands {
		relationshipData, err := b.resolveRelationship(
			rootResource,
			e.RelationshipName,
		)
		if err != nil {
			return err
		}

		e.Link = &TraversalStep{
			From:         rootResource,
			To:           relationshipData.To(),
			Relationship: relationshipData,
		}

		err = b.validateTraversal(e.Link)
		if err != nil {
			return err
		}

		err = b.bindMetadata(e.Link.To, depth+1, e.Operations)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b metadataBinder) bindFilterOperation(
	rootResource TableMetadata,
	depth int,
	ex FilterExpression,
) error {
	switch ex := ex.(type) {
	case *LogicalExpression:
		err := b.bindFilterOperation(rootResource, depth, ex.Left)
		if err != nil {
			return err
		}
		return b.bindFilterOperation(rootResource, depth, ex.Right)
	case *ComparisonExpression:
		traversalPathLength := len(ex.Path.Segments) - 1
		r, err := b.resolvePath(
			rootResource,
			ex.Path,
			traversalPathLength,
		)
		if err != nil {
			return err
		}
		ex.ResolvedPath = r

		if traversalPathLength > 0 && r.Type != RelationshipManyToOne {
			return bindingErr(
				"invalid value - '%s'. Comparison operations are only supported "+
					"for N:1 and 1:1 relationships. Please use a collection "+
					"operator",
				strings.Join(ex.Path.Segments, "/"),
			)
		}

		columnName := ex.Path.Segments[len(ex.Path.Segments)-1]
		columnData := r.EndResource.GetColumnMetadata(columnName)
		if columnData == nil {
			return bindingErr(
				"%s does not include a column definition for '%s'",
				r.EndResource.Name(),
				columnName,
			)
		}

		tableAccessPolicy, err := b.getTableAccessPolicy(r.EndResource)
		if err != nil {
			return err
		}
		err = validateColumnAccess(tableAccessPolicy, r.EndResource, columnData)
		if err != nil {
			return err
		}

		ex.ResolvedColumn = &ColumnValue{
			ColumnName: columnName,
			ColumnData: columnData,
		}
		return b.validateComparison(columnData, ex.Operator, ex.Value)
	case *CollectionExpression:
		traversalPathLength := len(ex.Path.Segments)
		r, err := b.resolvePath(
			rootResource,
			ex.Path,
			traversalPathLength,
		)

		if traversalPathLength > 0 && r.Type != RelationshipOneToMany {
			return bindingErr(
				"invalid value - '%s'. Collection operations are only supported "+
					"for 1:N relationships",
				strings.Join(ex.Path.Segments, "/"),
			)
		}

		if err != nil {
			return err
		}
		ex.ResolvedPath = r

		if len(ex.ResolvedPath.Steps) > 0 {
			depth = depth + 1
		}
		return b.bindFilterOperation(ex.ResolvedPath.EndResource, depth, ex.FilterExpression)
	default:
		return internalErr("unexpected expression type received")
	}
}

func (b metadataBinder) bindOrderByOperation(
	rootResource TableMetadata,
	ex *OrderByOperation,
) error {
	for _, r := range ex.Rules {
		traversalPathLength := len(r.Path.Segments) - 1
		p, err := b.resolvePath(
			rootResource,
			r.Path,
			traversalPathLength,
		)
		if err != nil {
			return err
		}
		r.ResolvedPath = p

		if traversalPathLength > 0 && p.Type != RelationshipManyToOne {
			return bindingErr(
				"invalid value - '%s'. Orderby operations are only supported for"+
					"for N:1 and 1:1 relationships",
				strings.Join(r.Path.Segments, "/"),
			)
		}

		columnName := r.Path.Segments[len(r.Path.Segments)-1]
		columnData := p.EndResource.GetColumnMetadata(columnName)
		if columnData == nil {
			return bindingErr(
				"%s does not include a column definition for '%s'",
				p.EndResource.Name(),
				columnName,
			)
		}

		tableAccessPolicy, err := b.getTableAccessPolicy(p.EndResource)
		if err != nil {
			return err
		}
		err = validateColumnAccess(tableAccessPolicy, p.EndResource, columnData)
		if err != nil {
			return err
		}

		r.ResolvedColumn = &ColumnValue{
			ColumnName: columnName,
			ColumnData: columnData,
		}
	}
	return nil
}

func (b *metadataBinder) resolveColumn(resource TableMetadata, columnName string) (ColumnMetadata, error) {
	metadata := resource.GetColumnMetadata(columnName)
	if metadata == nil {
		return nil, bindingErr(
			"table %s does not include a column definition for %s",
			resource.Name(),
			columnName,
		)
	}

	return metadata, nil
}

func (b *metadataBinder) resolveRelationship(resource TableMetadata, relationshipName string) (RelationshipMetadata, error) {
	metadata := resource.GetRelationshipMetadata(relationshipName)
	if metadata == nil {
		return nil, bindingErr(
			"table %s does not include a relationship definition for %s",
			resource.Name(),
			relationshipName,
		)
	}
	return metadata, nil
}

func (b *metadataBinder) resolvePath(
	rootResource TableMetadata,
	path PropertyPath,
	traversalPathLength int,
) (*ResolvedPath, error) {
	if len(path.Segments) == 0 {
		return nil, syntaxErr("path does not contain any segments")
	}

	if traversalPathLength != len(path.Segments) && traversalPathLength != len(path.Segments)-1 {
		return nil, internalErr(
			"traversal path length must be equal to the property path length " +
				"or the property path length - 1",
		)
	}

	traversalPathSegments := path.Segments[0:traversalPathLength]
	pathId := rootResource.Name() + "_" + strings.Join(traversalPathSegments, "_")

	o := &ResolvedPath{
		StartResource: rootResource,
		Steps:         make([]TraversalStep, traversalPathLength),
		EndResource:   rootResource,
	}

	i := 0
	for i < traversalPathLength {
		relationshipName := path.Segments[i]
		isIntermediateStep := i < traversalPathLength-1

		relationshipData := o.EndResource.GetRelationshipMetadata(relationshipName)
		if relationshipData == nil {
			return o, bindingErr(
				"%s does not include a relationship definition for '%s'",
				o.EndResource.FullyQualifiedName(),
				relationshipName,
			)
		}

		relationshipType := relationshipData.Type()
		if isIntermediateStep && relationshipType != RelationshipManyToOne {
			return o, bindingErr(
				"%s is a 1:N relationship and cannot be used as an "+
					"intermediate path step, please use a collection operator",
				relationshipData.Id(),
			)
		}

		o.Steps[i] = TraversalStep{
			From:         o.EndResource,
			Relationship: relationshipData,
			To:           relationshipData.To(),
		}
		err := b.validateTraversal(&o.Steps[i])
		if err != nil {
			return o, err
		}

		o.EndResource = relationshipData.To()
		o.Type = relationshipType
		i++
	}

	resolvedPathStore[pathId] = o
	return o, nil
}

func (b *metadataBinder) validateTraversal(step *TraversalStep) error {
	fromTableAccessPolicy, err := b.getTableAccessPolicy(step.From)
	if err != nil {
		return err
	}

	err = validateColumnAccess(
		fromTableAccessPolicy,
		step.From,
		step.Relationship.FromColumn(),
	)
	if err != nil {
		return err
	}

	toTableAccessPolicy, err := b.getTableAccessPolicy(step.To)
	if err != nil {
		return err
	}
	err = validateColumnAccess(
		toTableAccessPolicy,
		step.To,
		step.Relationship.ToColumn(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (b *metadataBinder) validateComparison(
	columnData ColumnMetadata,
	operator ComparisonOperator,
	value ValueExpression,
) error {
	isValid := value.IsSupportedByDbType(columnData.Type())
	if !isValid {
		return syntaxErr(
			"%s cannot be compared against type of %s",
			columnData.Name(),
			value.GetTypeName(),
		)
	}

	isValid = value.IsCompatibleWithComparisonOperator(operator)
	if !isValid {
		return syntaxErr(
			"%s cannot be used with the %s operator",
			value.GetTypeName(),
			string(operator),
		)
	}
	return nil
}

func (b *metadataBinder) getTableAccessPolicy(tableData TableMetadata) (TableAccessPolicy, error) {
	ap := b.accessPolicy.GetTableAccessPolicy(tableData.Name())
	if ap == nil {
		return nil, internalErr("unable to find access policy for the %s table", tableData.Name())
	}

	if !ap.CanAccess() {
		return nil, accessErr("you do not have permission to access the %s table", tableData.Name())
	}
	return ap, nil
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
