package handlers

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
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

type UsersTableState struct {
	User      *model.UsersModel
	TableData templates.TableData[*model.UsersModel]
}

func (s *UsersTableState) GetUser() *model.UsersModel {
	return s.User
}

func (s *UsersTableState) SetUser(v *model.UsersModel) {
	s.User = v
}

func UserStateInit() UserState {
	return &userState{}
}

func UsersTableStateInit() *UsersTableState {
	return &UsersTableState{}
}

type BusinessesTableState struct {
	User      *model.UsersModel
	TableData templates.TableData[*model.BusinessesModel]
}

func BusinessesTableStateInit() *BusinessesTableState {
	return &BusinessesTableState{}
}

func (s *BusinessesTableState) GetUser() *model.UsersModel {
	return s.User
}

func (s *BusinessesTableState) SetUser(v *model.UsersModel) {
	s.User = v
}
