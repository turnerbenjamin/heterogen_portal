package handlers

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/tables"
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

type TableState[T querymodel.TableModel] interface {
	GetTableData() *tables.TableData[T]
}

type userState struct {
	user *model.UsersModel
}

func UserStateInit() UserState {
	return &userState{}
}

func (s *userState) GetUser() *model.UsersModel {
	return s.user
}

func (s *userState) SetUser(v *model.UsersModel) {
	s.user = v
}

type UsersTableState struct {
	User      *model.UsersModel
	TableData *tables.TableData[*model.UsersModel]
}

func UsersTableStateInit() *UsersTableState {
	return &UsersTableState{
		TableData: &tables.TableData[*model.UsersModel]{
			TableDefinition: model.UsersTable,
		},
	}
}

func (s *UsersTableState) GetUser() *model.UsersModel {
	return s.User
}

func (s *UsersTableState) SetUser(v *model.UsersModel) {
	s.User = v
}

func (s *UsersTableState) GetTableData() *tables.TableData[*model.UsersModel] {
	return s.TableData
}

type BusinessesTableState struct {
	User      *model.UsersModel
	TableData *tables.TableData[*model.BusinessesModel]
}

func BusinessesTableStateInit() *BusinessesTableState {
	return &BusinessesTableState{
		TableData: &tables.TableData[*model.BusinessesModel]{
			TableDefinition: model.BusinessesTable,
		},
	}
}

func (s *BusinessesTableState) GetUser() *model.UsersModel {
	return s.User
}

func (s *BusinessesTableState) SetUser(v *model.UsersModel) {
	s.User = v
}

func (s *BusinessesTableState) GetTableData() *tables.TableData[*model.BusinessesModel] {
	return s.TableData
}
