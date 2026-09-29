package handlers

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
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
