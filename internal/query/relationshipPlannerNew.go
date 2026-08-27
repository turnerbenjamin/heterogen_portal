package query

import (
	"fmt"
	"iter"
)

type ExistsNode struct {
	alias       string
	step        *TraversalStep
	ParentAlias string
	Next        *ExistsNode
}

type Exists struct {
	RootAlias string
	FirstNode *ExistsNode
	LastNode  *ExistsNode
}

type JoinCollection interface {
	Get(string) *Join
	Add(string, *Join)
	Alias() string
	ParentAlias() string
	Joins() iter.Seq[Join]
	JoinsLen() int
}

type Join struct {
	alias       string
	step        TraversalStep
	parentAlias string
	joins       map[string]*Join
}

func (p *Join) Get(id string) *Join {
	plan, ok := p.joins[id]
	if !ok {
		return nil
	}
	return plan
}

func (p *Join) Add(id string, plan *Join) {
	p.joins[id] = plan
}

func (p *Join) Alias() string { return p.alias }

func (p *Join) ParentAlias() string { return p.parentAlias }

func (j *Join) Joins() iter.Seq[Join] {
	return func(yield func(Join) bool) {
		for _, v := range j.joins {
			if !yield(*v) {
				return
			}
		}
	}
}

func (j *Join) JoinsLen() int {
	return len(j.joins)
}

type JoinStore struct {
	rootResource    TableMetadata
	pathIdToJoinMap map[string]*Join // change to value
}

func (s JoinStore) Get(id string) *Join {
	if s.pathIdToJoinMap == nil {
		return nil
	}

	if _, exists := s.pathIdToJoinMap[id]; !exists {
		return nil
	}

	return s.pathIdToJoinMap[id]
}

func (s JoinStore) Add(id string, plan *Join) {
	if s.pathIdToJoinMap == nil {
		s.pathIdToJoinMap = map[string]*Join{}
	}
	s.pathIdToJoinMap[id] = plan
}

func (s JoinStore) Alias() string { return s.rootResource.Name() }

func (s JoinStore) ParentAlias() string { return "" }

func (s JoinStore) Joins() iter.Seq[Join] {
	return func(yield func(Join) bool) {
		for _, v := range s.pathIdToJoinMap {
			if !yield(*v) {
				return
			}
		}
	}
}

func (s JoinStore) JoinsLen() int {
	return len(s.pathIdToJoinMap)
}

type AliasStore struct {
	aliasCount       int
	pathIdToAliasMap map[string]string
}

func (s *AliasStore) GetAlias(pathId string) (string, bool) {
	if s.pathIdToAliasMap == nil {
		return "", false
	}

	if _, exists := s.pathIdToAliasMap[pathId]; !exists {
		return "", false
	}

	return s.pathIdToAliasMap[pathId], true
}

func (s *AliasStore) NextAlias() string {
	s.aliasCount++
	return fmt.Sprintf("t%d", s.aliasCount)
}

func (s *AliasStore) SetAlias(pathId string, alias string) {
	if s.pathIdToAliasMap == nil {
		s.pathIdToAliasMap = map[string]string{}
	}

	s.pathIdToAliasMap[pathId] = alias
}

type RelationshipPlannerNew struct {
	rootAlias string
	JoinStore JoinStore
	Aliases   AliasStore
}

func NewRelationshipPlanner(rootResource TableMetadata) (*RelationshipPlannerNew, error) {
	rootAlias := rootResource.Name()
	planner := &RelationshipPlannerNew{
		rootAlias: rootAlias,
		JoinStore: JoinStore{
			rootResource: rootResource,
		},
		Aliases: AliasStore{},
	}
	planner.Aliases.SetAlias(rootResource.Name(), rootAlias)
	return planner, nil
}

func (p *RelationshipPlannerNew) ProcessExists(path ResolvedPath) (*Exists, error) {
	return p.processExists(
		&Exists{RootAlias: p.rootAlias},
		&path,
		0,
	)
}

func (p *RelationshipPlannerNew) processExists(
	exists *Exists,
	path *ResolvedPath,
	depth int,
) (*Exists, error) {
	if exists == nil || exists.RootAlias == "" {
		return nil, internalErr("exists must be set with a root alias")
	}

	if path == nil {
		return nil, internalErr("unable to process exists as resolved path is nil")
	}

	if depth == len(path.Steps) {
		return exists, nil
	}

	step := path.Steps[depth]

	pathAlias := p.Aliases.NextAlias()

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

func (p *RelationshipPlannerNew) ProcessJoin(path ResolvedPath) error {
	return p.processJoin(
		p.JoinStore,
		path,
		0,
	)
}

func (p *RelationshipPlannerNew) processJoin(
	root JoinCollection,
	path ResolvedPath,
	depth int,
) error {
	if depth >= len(path.Steps) {
		return nil
	}

	step := path.Steps[depth]

	join := root.Get(step.SubPathId)
	if join == nil {
		pathAlias := p.Aliases.NextAlias()
		p.Aliases.SetAlias(step.SubPathId, pathAlias)

		join = &Join{
			alias:       pathAlias,
			step:        *step,
			parentAlias: root.Alias(),
			joins:       map[string]*Join{},
		}
		root.Add(step.SubPathId, join)
	}

	return p.processJoin(join, path, depth+1)
}
