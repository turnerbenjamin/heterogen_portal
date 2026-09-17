package relationships

import (
	"fmt"
	"iter"

	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type existsNode struct {
	step        mdl.TraversalStep
	alias       string
	parentAlias string
	next        *existsNode
}

type aliasType uint8

const (
	aliasTypeExists aliasType = iota
	aliasTypeJoin
)

func (e *existsNode) Step() mdl.TraversalStep {
	return e.step
}

func (e *existsNode) Alias() string {
	return e.alias
}

func (e *existsNode) ParentAlias() string {
	return e.parentAlias
}

func (e *existsNode) Next() mdl.ExistsNode {
	return e.next
}

type Join struct {
	ParentAlias string
	Alias       string
	Step        mdl.TraversalStep
	SubJoins    JoinCollection
}

type JoinCollection struct {
	joins map[string]*Join
}

func (c *JoinCollection) getJoinByPathId(pathId string) (*Join, bool) {
	if c.joins == nil {
		return nil, false
	}

	join, ok := c.joins[pathId]
	return join, ok
}

func (c *JoinCollection) addJoin(pathId string, join *Join) {
	if c.joins == nil {
		c.joins = map[string]*Join{}
	}
	c.joins[pathId] = join
}

func (c *JoinCollection) Joins() iter.Seq2[string, *Join] {
	return func(yield func(string, *Join) bool) {
		for k, v := range c.joins {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (c *JoinCollection) JoinsLen() int {
	return len(c.joins)
}

type AliasStore struct {
	aliasCount             int
	rootAlias              string
	pathIdToExistsAliasMap map[string]string
	pathIdToJoinsAliasMap  map[string]string
}

func (s *AliasStore) GetRootAlias() string {
	return s.rootAlias
}

func (s *AliasStore) GetExistsAlias(path mdl.ResolvedPath) (string, bool) {
	return s.getAlias(aliasTypeExists, path)
}

func (s *AliasStore) GetJoinAlias(path mdl.ResolvedPath) (string, bool) {
	return s.getAlias(aliasTypeJoin, path)
}

func (s *AliasStore) getAlias(t aliasType, path mdl.ResolvedPath) (string, bool) {
	if len(path.Steps) == 0 {
		return s.rootAlias, true
	}
	pathId := path.Id

	mapping := s.getAliasMapping(t)
	if _, exists := mapping[pathId]; !exists {
		return "", false
	}

	return mapping[pathId], true
}

func (s *AliasStore) nextAlias(t aliasType, pathId string) string {
	mapping := s.getAliasMapping(t)
	if _, exists := mapping[pathId]; !exists {
		s.aliasCount++
		mapping[pathId] = fmt.Sprintf("t%d", s.aliasCount)
	}

	return mapping[pathId]
}

func (s *AliasStore) getAliasMapping(t aliasType) map[string]string {
	switch t {
	case aliasTypeExists:
		if s.pathIdToExistsAliasMap == nil {
			s.pathIdToExistsAliasMap = make(map[string]string, 1)
		}
		return s.pathIdToExistsAliasMap
	case aliasTypeJoin:
		if s.pathIdToJoinsAliasMap == nil {
			s.pathIdToJoinsAliasMap = make(map[string]string, 1)
		}
		return s.pathIdToJoinsAliasMap
	default:
		panic("invalid alias type received")
	}
}

type RelationshipPlanner struct {
	Aliases   AliasStore
	JoinStore JoinCollection
}

func NewRelationshipPlanner(rootResource mdl.TableMetadata) (*RelationshipPlanner, error) {
	planner := &RelationshipPlanner{
		Aliases:   AliasStore{rootAlias: rootResource.Name()},
		JoinStore: JoinCollection{},
	}

	return planner, nil
}

func (p *RelationshipPlanner) ProcessExists(path mdl.ResolvedPath) []mdl.ExistsNodeNew {
	existsNodes := make([]mdl.ExistsNodeNew, len(path.Steps))

	parentAlias := p.Aliases.rootAlias
	for i, step := range path.Steps {
		nextAlias := p.Aliases.nextAlias(aliasTypeExists, step.SubPathId)

		existsNodes[i] = mdl.ExistsNodeNew{
			Step:        step,
			Alias:       nextAlias,
			ParentAlias: parentAlias,
		}
		parentAlias = nextAlias
	}
	return existsNodes
}

func (p *RelationshipPlanner) ProcessJoin(path mdl.ResolvedPath) error {
	pathLen := len(path.Steps)
	if pathLen == 0 {
		return nil
	}

	// Handle initial join
	step := path.Steps[0]
	currentJoin, exists := p.JoinStore.getJoinByPathId(step.SubPathId)
	if !exists {
		currentJoin = &Join{
			ParentAlias: p.Aliases.rootAlias,
			Alias:       p.Aliases.nextAlias(aliasTypeJoin, step.SubPathId),
			Step:        *step,
		}
		p.JoinStore.addJoin(step.SubPathId, currentJoin)
	}
	if pathLen < 2 {
		return nil
	}

	for _, step := range path.Steps[1:] {
		nextJoin, exists := currentJoin.SubJoins.getJoinByPathId(step.SubPathId)
		if !exists {
			nextJoin = &Join{
				ParentAlias: currentJoin.Alias,
				Alias:       p.Aliases.nextAlias(aliasTypeJoin, step.SubPathId),
				Step:        *step,
			}
			currentJoin.SubJoins.addJoin(step.SubPathId, nextJoin)
		}
		currentJoin = nextJoin
	}
	return nil
}
