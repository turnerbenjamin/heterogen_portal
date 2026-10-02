package handlers

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/querystack/querymodel"
)

// NoState is a placeholder type for handlers that do not require
// pipeline state.
type NoState struct{}

// NoStateInit can be used as a state initialiser when using NoState
func NoStateInit() NoState {
	return NoState{}
}

type UserState interface {
	GetUser() *model.UsersModel
	SetUser(v *model.UsersModel)
}
type userState struct {
	user *model.UsersModel
}

func (s *userState) GetUser() *model.UsersModel {
	return s.user
}

func (s *userState) SetUser(v *model.UsersModel) {
	s.user = v
}

func UserStateInit() UserState {
	return &userState{}
}

type HeaderColumn[T querymodel.TableModel] struct {
	Column ColumnData[T]
	Width  string
}

type ColumnData[T querymodel.TableModel] struct {
	ColumnName  string
	DisplayName string
	DataType    model.DataType
	Path        string
	Expansion   string
	GetData     func(T) string
}

type TableState[T querymodel.TableModel] struct {
	TableName         string
	User              *model.UsersModel
	SchemaVersion     uint64
	ColumnDefinitions []ColumnData[T]
	TableHeader       []HeaderColumn[T]
	Rows              []T
}

func UsersTableStateInit() *TableState[*model.UsersModel] {
	return &TableState[*model.UsersModel]{}
}

func BusinessesTableStateInit() *TableState[*model.BusinessesModel] {
	return &TableState[*model.BusinessesModel]{}
}

func (s *TableState[T]) GetUser() *model.UsersModel {
	return s.User
}

func (s *TableState[T]) SetUser(v *model.UsersModel) {
	s.User = v
}
