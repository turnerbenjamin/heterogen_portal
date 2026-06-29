package handlers

import "github.com/turnerbenjamin/heterogen_portal/internal/db"

// NoState is a placeholder type for handlers that do not require
// pipeline state.
type NoState struct{}

// NoStateInit can be used as a state initialiser when using NoState
func NoStateInit() NoState {
	return NoState{}
}

type UserState interface {
	GetUser() *db.User
	SetUser(v *db.User)
}
type userState struct {
	user *db.User
}

func (s *userState) GetUser() *db.User {
	return s.user
}

func (s *userState) SetUser(v *db.User) {
	s.user = v
}

func UserStateInit() UserState {
	return &userState{}
}
