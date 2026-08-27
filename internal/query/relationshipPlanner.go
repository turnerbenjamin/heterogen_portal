package query

// import "fmt"

// type JoinCollection interface {
// 	Get(string) *Join
// 	Add(string, *Join)
// 	Alias() string
// }

// type Join struct {
// 	alias       string
// 	step        TraversalStep
// 	ParentAlias string
// 	Joins       map[string]Join
// }

// func (p *Join) Get(id string) *Join {
// 	plan, ok := p.Joins[id]
// 	if !ok {
// 		return nil
// 	}
// 	return plan
// }

// func (p *Join) Add(id string, plan *Join) {
// 	p.Joins[id] = plan
// }

// func (p *Join) Alias() string { return p.alias }

// type ExistsNode struct {
// 	alias       string
// 	step        *TraversalStep
// 	ParentAlias string
// 	Next        *ExistsNode
// }

// type Exists struct {
// 	RootAlias string
// 	FirstNode *ExistsNode
// 	LastNode  *ExistsNode
// }

// type RelationshipPlanner struct {
// 	aliasCount   int
// 	rootAlias    string
// 	Joins        map[string]*Join
// 	TableAliases map[string]string
// }

// func (p *RelationshipPlanner) Get(id string) *Join {
// 	plan, ok := p.Joins[id]
// 	if !ok {
// 		return nil
// 	}
// 	return plan
// }

// func (p *RelationshipPlanner) Add(id string, plan *Join) {
// 	p.Joins[id] = plan
// }

// func (p *RelationshipPlanner) Alias() string { return p.rootAlias }

// func (p *RelationshipPlanner) getNextAlias() string {
// 	p.aliasCount++
// 	return fmt.Sprintf("t%d", p.aliasCount)
// }

// func NewRelationshipPlan(operations *Operations) (*RelationshipPlanner, error) {
// 	rootAlias := operations.rootResource.Name()
// 	plan := &RelationshipPlanner{
// 		rootAlias: rootAlias,
// 		Joins:     map[string]*Join{},
// 		TableAliases: map[string]string{
// 			rootAlias: rootAlias,
// 		},
// 	}

// 	if operations.OrderByOperation != nil {
// 		err := plan.processOrderByOperation(operations.OrderByOperation)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}

// 	if operations.FilterOperation != nil {
// 		err := plan.processFilterExpression(
// 			plan.rootAlias,
// 			operations.FilterOperation.FilterExpression,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}

// 	return plan, nil
// }

// func (p *RelationshipPlanner) processOrderByOperation(op *OrderByOperation) error {
// 	for _, r := range op.Rules {
// 		err := p.processJoin(p, r.ResolvedPath, 0)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (p *RelationshipPlanner) processFilterExpression(
// 	rootAlias string,
// 	ex FilterExpression,
// ) error {
// 	switch ex := ex.(type) {
// 	case *LogicalExpression:
// 		p.processFilterExpression(rootAlias, ex.Left)
// 		p.processFilterExpression(rootAlias, ex.Right)

// 	case *ComparisonExpression:
// 		exists, err := p.processExists(
// 			&Exists{RootAlias: rootAlias},
// 			ex.ResolvedPath,
// 			0,
// 		)
// 		if err != nil {
// 			return err
// 		}

// 		ex.ExistsPlan = exists

// 	case *CollectionExpression:
// 		exists, err := p.processExists(
// 			&Exists{RootAlias: rootAlias},
// 			ex.ResolvedPath,
// 			0,
// 		)
// 		if err != nil {
// 			return err
// 		}

// 		ex.ExistsPlan = exists

// 		p.processFilterExpression(exists.LastNode.alias, ex.FilterExpression)
// 	default:
// 		return internalErr("unexpected filter expression received")
// 	}
// 	return nil
// }

// func (p *RelationshipPlanner) processExists(
// 	exists *Exists,
// 	path *ResolvedPath,
// 	depth int,
// ) (*Exists, error) {
// 	if exists == nil || exists.RootAlias == "" {
// 		return nil, internalErr("exists must be set with a root alias")
// 	}

// 	if path == nil {
// 		return nil, internalErr("unable to process exists as resolved path is nil")
// 	}

// 	if depth == len(path.Steps) {
// 		return exists, nil
// 	}

// 	step := path.Steps[depth]

// 	pathAlias := p.getNextAlias()

// 	parentAlias := exists.RootAlias
// 	if exists.LastNode != nil && exists.LastNode.alias != "" {
// 		parentAlias = exists.LastNode.alias
// 	}

// 	node := &ExistsNode{
// 		alias:       pathAlias,
// 		ParentAlias: parentAlias,
// 		step:        step,
// 		Next:        nil,
// 	}

// 	if exists.FirstNode == nil {
// 		exists.FirstNode = node
// 	}

// 	if exists.LastNode != nil {
// 		exists.LastNode.Next = node
// 	}
// 	exists.LastNode = node

// 	return p.processExists(
// 		exists,
// 		path,
// 		depth+1,
// 	)
// }

// func (p *RelationshipPlanner) processJoin(
// 	root JoinCollection,
// 	path *ResolvedPath,
// 	depth int,
// ) error {
// 	if path == nil {
// 		return internalErr("unable to process exists as resolved path is nil")
// 	}

// 	if depth >= len(path.Steps) {
// 		return nil
// 	}

// 	step := path.Steps[depth]

// 	pathAlias := p.getNextAlias()
// 	p.TableAliases[step.SubPathId] = pathAlias

// 	join := root.Get(step.SubPathId)
// 	if join == nil {
// 		join = &Join{
// 			alias:       pathAlias,
// 			step:        step,
// 			ParentAlias: root.Alias(),
// 			Joins:       map[string]*Join{},
// 		}
// 		root.Add(step.SubPathId, join)
// 	}

// 	return p.processJoin(join, path, depth+1)
// }
