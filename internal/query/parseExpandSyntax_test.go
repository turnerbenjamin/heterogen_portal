package query

import "testing"

func TestParseExpand_Success(t *testing.T) {
	testData := []struct {
		name  string
		query string
	}{
		{
			name:  "single expansion",
			query: "relationship_1",
		},
		{
			name:  "multiple expansion",
			query: "relationship_1,relationship_2",
		},
		{
			name:  "expansion with select",
			query: "relationship_1(select=column_1)",
		},
		{
			name:  "expansion with filter",
			query: "column_1(filter=column_2 eq 1)",
		},
		{
			name:  "expansion with multiple nested operations",
			query: "column_1(select=column_2;filter=column_2 eq 1)",
		},
	}
	_ = testData
}

/*
## 4. parseExpandOperation

This deserves almost as much attention as parseOperations.

### TestParseExpand_Success

t.Run(...)

- Single expansion
- Multiple expansions
- Expansion with select
- Expansion with filter
- Expansion with multiple nested operations
- Expansion with nested expand
- Deep nested expansions
- Empty nested operations (if supported)

### TestParseExpand_InvalidRelationship

t.Run(...)

- EOF
- Comma
- Parenthesis
- Equals
- Semicolon

### TestParseExpand_InvalidNestedSyntax

t.Run(...)

- Missing closing parenthesis
- Empty parentheses
- Unexpected comma
- Unexpected semicolon
- Missing nested operator
- Missing equals in nested operation

### TestParseExpand_AST

t.Run(...)

- Relationship preserved
- Nested operations preserved
- Multiple expansions preserved
- Nested operation order preserved
- Link metadata remains nil

## 5. Nested grammar tests

These are some of the highest-value tests.

### TestNestedQueries

t.Run(...)

- expand(select)
- expand(filter)
- expand(select;filter)
- expand(filter;select)
- expand(expand(select))
- expand(expand(expand(select)))
- Multiple sibling expansions
- Nested sibling expansions
- Multiple nested operators
- Complex realistic query

Example:

expand=orders(
    select=id,total;
    filter=status eq 'Open';
    expand=items(
        select=id,price;
        expand=product(
            select=name
        )
    )
)
*/
