package query

import (
	"strings"
)

var resolvedPathStore = map[string]ResolvedPath{}

type ResolvedPath struct {
	StartResource TableMetadata
	Steps         []TraversalStep
	EndResource   TableMetadata
}

type TraversalStep struct {
	From         TableMetadata
	Relationship RelationshipMetadata
	To           TableMetadata
}

type metadataBinder struct {
	maximumDepth int
	// schemaMetadata SchemaMetadata
}

func (b metadataBinder) bindMetadata(
	rootResource TableMetadata,
	depth int,
	operations []QueryOperation,
) error {
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
	for _, s := range op.Columns {
		if s.columnData != nil {
			continue
		}

		columnData, err := b.resolveColumn(rootResource, s.columnName)
		if err != nil {
			return err
		}
		s.columnData = columnData
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
			e.Relationship.relationshipName,
		)
		if err != nil {
			return err
		}

		e.Relationship.relationshipData = relationshipData
		e.Link = &TraversalStep{
			From:         rootResource,
			To:           relationshipData.To(),
			Relationship: relationshipData,
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

		columnName := ex.Path.Segments[len(ex.Path.Segments)-1]
		columnData := r.EndResource.GetColumnMetadata(columnName)
		if columnData == nil {
			return bindingErr(
				"%s does not include a column definition for '%s'",
				r.EndResource.Name(),
				columnName,
			)
		}

		ex.ResolvedColumn = &ColumnValue{
			columnName: columnName,
			columnData: columnData,
		}
		return b.validateComparison(columnData, ex.Operator, ex.Value)
	case *CollectionExpression:
		traversalPathLength := len(ex.Path.Segments)
		r, err := b.resolvePath(
			rootResource,
			ex.Path,
			traversalPathLength,
		)
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
) (ResolvedPath, error) {
	if len(path.Segments) == 0 {
		return ResolvedPath{}, syntaxErr("path does not contain any segments")
	}

	if traversalPathLength != len(path.Segments) && traversalPathLength != len(path.Segments)-1 {
		return ResolvedPath{}, internalErr(
			"traversal path length must be equal to the property path length " +
				"or the property path length - 1",
		)
	}

	pathId := rootResource.Name() + "_" + strings.Join(path.Segments, "_")
	if p, exists := resolvedPathStore[pathId]; exists {
		return p, nil
	}

	o := ResolvedPath{
		StartResource: rootResource,
		Steps:         make([]TraversalStep, traversalPathLength),
		EndResource:   rootResource,
	}
	i := 0
	for i < traversalPathLength {
		relationshipName := path.Segments[i]
		relationshipData := o.EndResource.GetRelationshipMetadata(relationshipName)
		if relationshipData == nil {
			return o, bindingErr(
				"%s does not include a relationship definition for '%s'",
				o.EndResource.FullyQualifiedName(),
				relationshipName,
			)
		}

		o.Steps[i] = TraversalStep{
			From:         o.EndResource,
			Relationship: relationshipData,
			To:           relationshipData.To(),
		}
		o.EndResource = relationshipData.To()
		i++
	}

	return o, nil
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
