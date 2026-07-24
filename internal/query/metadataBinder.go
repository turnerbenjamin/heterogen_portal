package query

import (
	"fmt"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/internal/model"
)

var resolvedPathStore = map[string]ResolvedPath{}

type ResolvedPath struct {
	StartResource *model.TableMetadata
	Steps         []TraversalStep
	EndResource   *model.TableMetadata
}

type TraversalStep struct {
	From         *model.TableMetadata
	Relationship *model.Relationship
	To           *model.TableMetadata
}

type metadataBinder struct {
	maximumDepth int
}

func (b metadataBinder) bindMetadata(
	rootResource *model.TableMetadata,
	depth int,
	operations []QueryOperation,
) error {
	if depth > b.maximumDepth {
		return fmt.Errorf("maximum expand depth (%d) exceeded", b.maximumDepth)
	}

	for _, op := range operations {
		switch op := op.(type) {
		case *SelectOperation:
			return b.bindSelectOperation(rootResource, op)
		case *ExpandOperation:
			return b.bindExpandOperation(rootResource, depth, op)
		default:
			return fmt.Errorf("unexpected operation received")
		}
	}
	return nil
}

func (b metadataBinder) bindSelectOperation(
	rootResource *model.TableMetadata,
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
	rootResource *model.TableMetadata,
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
			To:           relationshipData.GetTo(),
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
	rootResource *model.TableMetadata,
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
		columnData := r.EndResource.GetColumn(columnName)
		if columnData == nil {
			return fmt.Errorf(
				"%s does not include a column definition for '%s'",
				r.EndResource.GetResourceShortName(),
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
		return fmt.Errorf("unexpected expression type received")
	}
}

func (b *metadataBinder) resolveColumn(resource *model.TableMetadata, columnName string) (*model.ColumnMetadata, error) {
	metadata := resource.GetColumn(columnName)
	if metadata == nil {
		return nil, fmt.Errorf(
			"table %s does not include a column definition for %s",
			resource.GetResourceShortName(),
			columnName,
		)
	}
	return metadata, nil
}

func (b *metadataBinder) resolveRelationship(resource *model.TableMetadata, relationshipName string) (*model.Relationship, error) {
	metadata := resource.GetRelationship(relationshipName)
	if metadata == nil {
		return nil, fmt.Errorf(
			"table %s does not include a relationship definition for %s",
			resource.GetResourceShortName(),
			relationshipName,
		)
	}
	return metadata, nil
}

func (b *metadataBinder) resolvePath(
	rootResource *model.TableMetadata,
	path PropertyPath,
	traversalPathLength int,
) (ResolvedPath, error) {
	if len(path.Segments) == 0 {
		return ResolvedPath{}, fmt.Errorf("path does not contain any segments")
	}

	if traversalPathLength != len(path.Segments) && traversalPathLength != len(path.Segments)-1 {
		return ResolvedPath{}, fmt.Errorf(
			"traversal path length must be equal to the property path length " +
				"or the property path length - 1",
		)
	}

	pathId := rootResource.TableName + "_" + strings.Join(path.Segments, "_")
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
		relationshipData := o.EndResource.GetRelationship(relationshipName)
		if relationshipData == nil {
			return o, fmt.Errorf(
				"%s does not include a relationship definition for '%s'",
				o.EndResource.GetResourceFullname(),
				relationshipName,
			)
		}

		toResource := model.GetTableMetadata(relationshipData.RelatedTable)
		if toResource == nil {
			return o, fmt.Errorf(
				"unable to find table metadata for %s",
				relationshipData.RelatedTable,
			)
		}

		o.Steps[i] = TraversalStep{
			From:         o.EndResource,
			Relationship: relationshipData,
			To:           toResource,
		}
		o.EndResource = toResource
		i++
	}

	return o, nil
}

func (b *metadataBinder) validateComparison(
	columnData *model.ColumnMetadata,
	operator ComparisonOperator,
	value ValueExpression,
) error {
	isValid := value.IsSupportedByDbType(columnData.Type)
	if !isValid {
		return fmt.Errorf(
			"%s cannot be compared against type of %s",
			columnData.Name,
			value.GetTypeName(),
		)
	}

	isValid = value.IsCompatibleWithComparisonOperator(operator)
	if !isValid {
		return fmt.Errorf(
			"%s cannot be used with the %s operator",
			value.GetTypeName(),
			string(operator),
		)
	}
	return nil
}

/*
func newInvalidComparisonTypesError(typeName string, columnName string) *etc.AppError {
	return &etc.AppError{
		Code: http.StatusBadRequest,
		ErrorMessage: }
}

func newInvalidOperatorTypesError(typeName string, operatorName string) *etc.AppError {
	return &etc.AppError{
		Code: http.StatusBadRequest,
		ErrorMessage: fmt.Sprintf(
			"%s cannot be used with the %s operator",
			typeName,
			operatorName,
		)}
*/
