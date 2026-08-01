package query

import "fmt"

type JoinCollection interface {
	Get(string) *Join
	Add(string, *Join)
	Alias() string
}

type Join struct {
	alias       string
	step        TraversalStep
	ParentAlias string
	Joins       map[string]*Join
}

func (p *Join) Get(id string) *Join {
	plan, ok := p.Joins[id]
	if !ok {
		return nil
	}
	return plan
}

func (p *Join) Add(id string, plan *Join) {
	p.Joins[id] = plan
}

func (p *Join) Alias() string { return p.alias }

type ExistsNode struct {
	alias       string
	step        TraversalStep
	ParentAlias string
	Next        *ExistsNode
}

type Exists struct {
	Id        string
	RootAlias string
	FirstNode *ExistsNode
	LastNode  *ExistsNode
}

type RelationshipPlanner struct {
	aliasCount   int
	rootAlias    string
	Joins        map[string]*Join
	TableAliases map[string]string
}

func (p *RelationshipPlanner) Get(id string) *Join {
	plan, ok := p.Joins[id]
	if !ok {
		return nil
	}
	return plan
}

func (p *RelationshipPlanner) Add(id string, plan *Join) {
	p.Joins[id] = plan
}

func (p *RelationshipPlanner) Alias() string { return p.rootAlias }

func (p *RelationshipPlanner) getNextAlias() string {
	p.aliasCount++
	return fmt.Sprintf("t%d", p.aliasCount)
}

func NewRelationshipPlan(operations []QueryOperation) (*RelationshipPlanner, error) {
	rootAlias := "root"
	plan := &RelationshipPlanner{
		rootAlias: rootAlias,
		Joins:     map[string]*Join{},
		TableAliases: map[string]string{
			rootAlias: rootAlias,
		},
	}

	for _, op := range operations {
		switch op := op.(type) {
		case *OrderByOperation:
			plan.processOrderByOperation(op)
		case *FilterOperation:
			err := plan.processFilterOperation(plan.rootAlias, op)
			if err != nil {
				return nil, err
			}
		default:
			continue
		}
	}
	return plan, nil
}

func (p *RelationshipPlanner) processOrderByOperation(op *OrderByOperation) {
	for _, r := range op.Rules {
		p.processJoin(p.rootAlias, p, r.ResolvedPath, 0)
	}
}

func (p *RelationshipPlanner) processFilterOperation(
	rootAlias string,
	ex *FilterOperation,
) error {
	return p.processFilterExpression(rootAlias, ex.FilterExpression)
}

func (p *RelationshipPlanner) processFilterExpression(
	rootAlias string,
	ex FilterExpression,
) error {
	switch ex := ex.(type) {
	case *LogicalExpression:
		p.processFilterExpression(rootAlias, ex.Left)
		p.processFilterExpression(rootAlias, ex.Right)

	case *ComparisonExpression:
		exists, err := p.processExists(
			&Exists{RootAlias: rootAlias},
			ex.ResolvedPath,
			0,
		)
		if err != nil {
			return err
		}

		ex.ExistsPlan = exists

	case *CollectionExpression:
		exists, err := p.processExists(
			&Exists{RootAlias: rootAlias},
			ex.ResolvedPath,
			0,
		)
		if err != nil {
			return err
		}

		ex.ExistsPlan = exists

		p.processFilterExpression(exists.LastNode.alias, ex.FilterExpression)
	default:
		return internalErr("unexpected filter expression received")
	}
	return nil
}

func (p *RelationshipPlanner) processExists(
	exists *Exists,
	path *ResolvedPath,
	depth int,
) (*Exists, error) {
	if exists == nil || exists.RootAlias == "" {
		return nil, internalErr("exists must be set with a root alias")
	}

	if depth == len(path.Steps) {
		return exists, nil
	}

	step := path.Steps[depth]

	pathAlias := p.getNextAlias()

	parentAlias := exists.RootAlias
	if exists.LastNode != nil && exists.LastNode.alias != "" {
		parentAlias = exists.LastNode.alias
	}

	node := &ExistsNode{
		alias:       pathAlias,
		ParentAlias: parentAlias,
		step:        step,
		Next:        nil,
	}

	if exists.FirstNode == nil {
		exists.FirstNode = node
	}

	if exists.LastNode != nil {
		exists.LastNode.Next = node
	}
	exists.LastNode = node

	return p.processExists(
		exists,
		path,
		depth+1,
	)
}

func (p *RelationshipPlanner) processJoin(
	rootId string,
	root JoinCollection,
	path *ResolvedPath,
	depth int,
) {
	if path == nil || depth == len(path.Steps) {
		path.Id = rootId
		return
	}

	step := path.Steps[depth]
	pathId := rootId + "_" + step.To.Name()

	pathAlias := p.getNextAlias()
	p.TableAliases[pathId] = pathAlias

	join := root.Get(pathId)
	if join == nil {
		join := &Join{
			alias:       pathAlias,
			step:        step,
			ParentAlias: root.Alias(),
			Joins:       map[string]*Join{},
		}
		root.Add(pathId, join)
	}

	p.processJoin(pathId, join, path, depth+1)
}
