package azSqlWriter

import (
	"fmt"

	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type ListLiteral struct {
	listType mdl.LiteralType
	el0      mdl.ValueExpression
	els      []mdl.ValueExpression
}

func newListLiteral(els []mdl.ValueExpression) (*ListLiteral, error) {
	if len(els) == 0 {
		return nil, qerr.InternalErr("a list literal cannot be empty")
	}

	el0 := els[0]
	elType := el0.Type()

	var listType mdl.LiteralType
	switch elType {
	case mdl.LiteralTypeString:
		listType = mdl.LiteralTypeListString
	case mdl.LiteralTypeInt:
		listType = mdl.LiteralTypeListInt
	case mdl.LiteralTypeFloat:
		listType = mdl.LiteralTypeListFloat
	default:
		return nil, qerr.InternalErr("lists of type %s", mdl.LiteralTypeName(elType))
	}

	o := &ListLiteral{
		listType: listType,
		el0:      el0,
		els:      make([]mdl.ValueExpression, len(els)),
	}

	for i, el := range els {
		if el.Type() != elType {
			return nil, qerr.InternalErr("mixed-type lists are not supported")
		}

		o.els[i] = el
	}
	return o, nil
}

func (l *ListLiteral) Type() mdl.LiteralType {
	return l.listType
}

func (l *ListLiteral) Value() any {
	return l.els
}

func (l *ListLiteral) Serialise(s mdl.Serialiser) {
	s.SerialiseList(l.els)
}

func (l *ListLiteral) IsCompatibleWithComparisonOperator(op mdl.ComparisonOperator) bool {
	return op == mdl.ComparisonIn
}

func (l *ListLiteral) IsSupportedByDbType(dbType mdl.DbDataTypeName) bool {
	return l.el0.IsSupportedByDbType(dbType)
}

func (l *ListLiteral) WriteFilterExpression(
	w mdl.QueryWriter,
	fieldName string,
	op mdl.ComparisonOperator,
) error {
	switch op {
	case mdl.ComparisonIn, mdl.ComparisonNotIn:
		modifier := ""
		if op == mdl.ComparisonNotIn {
			modifier = "NOT "
		}

		w.Write("%s %sIN (", fieldName, modifier)
		for i, el := range l.els {
			if i != 0 {
				w.Write(",")
			}
			w.Write(w.Placeholder(el.Value()))
		}
		w.Write(")")
	default:
		return fmt.Errorf("unsupported list operation: %s", string(op))
	}
	return nil
}

// func WriteListFilterExpression[T any](
// 	values []T,
// 	w mdl.QueryWriter,
// 	fieldName string,
// 	op mdl.ComparisonOperator,
// ) error {
// 	switch op {
// 	case mdl.ComparisonIn, mdl.ComparisonNotIn:
// 		modifier := ""
// 		if op == mdl.ComparisonNotIn {
// 			modifier = "NOT "
// 		}

// 		w.Write("%s %sIN (", fieldName, modifier)
// 		for i, v := range values {
// 			if i != 0 {
// 				w.Write(",")
// 			}
// 			w.Write(w.Placeholder(v))
// 		}
// 		w.Write(")")
// 	default:
// 		return fmt.Errorf("unsupported list operation: %s", string(op))
// 	}
// 	return nil
// }

/*
LITERALS ARE TIED TO SYNTAX - WE NEED A LITERAL BUILDER INTERFACE TO BE RETURNED
FROM THE WRITER - THE BELOW WILL USED FOR THE AZURE SQL IMPLEMENTATION OF THE
INTERFACE - WE WILL HAVE FROMSTRING,INT METHODS

LITERALS NEED TO BE RESPONSIBLE FOR WRITING SQL, SERIALISING FOR TOKEN STRINGS
AND COMPATIBILITY TO ISOLATE THEM FROM THE REST OF THE QUERY PACKAGE
*/

// type StringListLiteral struct {
// 	Values []string
// }

// func (l *StringListLiteral) GetTypeName() string { return "string list" }

// func (l *StringListLiteral) GetType() mdl.LiteralType { return mdl.LiteralTypeStringList }

// func (l *StringListLiteral) IsCompatibleWithComparisonOperator(op mdl.ComparisonOperator) bool {
// 	return op == mdl.ComparisonIn
// }

// func (l *StringListLiteral) IsSupportedByDbType(dbType mdl.DbDataTypeName) bool {
// 	sl := StringLiteral{}
// 	return (&sl).IsSupportedByDbType(dbType)
// }

// func (l *StringListLiteral) WriteFilterExpression(
// 	w mdl.QueryWriter,
// 	fieldName string,
// 	op mdl.ComparisonOperator,
// ) error {
// 	return WriteListFilterExpression(l.Values, w, fieldName, op)
// }

// type IntListLiteral struct {
// 	Values []int64
// }

// func (l *IntListLiteral) GetTypeName() string      { return "int" }
// func (l *IntListLiteral) GetType() mdl.LiteralType { return mdl.LiteralTypeIntList }

// func (l *IntListLiteral) IsCompatibleWithComparisonOperator(op mdl.ComparisonOperator) bool {
// 	return op == mdl.ComparisonIn
// }

// func (l *IntListLiteral) IsSupportedByDbType(dbType mdl.DbDataTypeName) bool {
// 	il := IntListLiteral{}
// 	return (&il).IsSupportedByDbType(dbType)
// }

// func (l *IntListLiteral) WriteFilterExpression(
// 	w mdl.QueryWriter,
// 	fieldName string,
// 	op mdl.ComparisonOperator,
// ) error {
// 	return WriteListFilterExpression(l.Values, w, fieldName, op)
// }

// type FloatListLiteral struct {
// 	Values []float64
// }

// func (l *FloatListLiteral) GetTypeName() string      { return "float list" }
// func (l *FloatListLiteral) GetType() mdl.LiteralType { return mdl.LiteralTypeFloatList }

// func (l *FloatListLiteral) IsCompatibleWithComparisonOperator(op mdl.ComparisonOperator) bool {
// 	return op == mdl.ComparisonIn
// }

// func (l *FloatListLiteral) IsSupportedByDbType(dbType mdl.DbDataTypeName) bool {
// 	fl := FloatLiteral{}
// 	return (&fl).IsSupportedByDbType(dbType)
// }

// func (l *FloatListLiteral) WriteFilterExpression(
// 	w mdl.QueryWriter,
// 	fieldName string,
// 	op mdl.ComparisonOperator,
// ) error {
// 	return WriteListFilterExpression(l.Values, w, fieldName, op)
// }

// func WriteListFilterExpression[T any](
// 	values []T,
// 	w mdl.QueryWriter,
// 	fieldName string,
// 	op mdl.ComparisonOperator,
// ) error {
// 	switch op {
// 	case mdl.ComparisonIn, mdl.ComparisonNotIn:
// 		modifier := ""
// 		if op == mdl.ComparisonNotIn {
// 			modifier = "NOT "
// 		}

// 		w.Write("%s %sIN (", fieldName, modifier)
// 		for i, v := range values {
// 			if i != 0 {
// 				w.Write(",")
// 			}
// 			w.Write(w.Placeholder(v))
// 		}
// 		w.Write(")")
// 	default:
// 		return fmt.Errorf("unsupported list operation: %s", string(op))
// 	}
// 	return nil
// }
