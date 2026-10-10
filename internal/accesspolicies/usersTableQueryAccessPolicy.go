package accesspolicies

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
)

// Query access policy for business table query access
var UsersTableQueryAccessPolicy = &model.DatabaseAccessPolicy{

	// Users table permissions
	UsersAccessPolicy: &model.UsersAccessPolicy{
		UserCanAccess: true,

		// Users: Allow
		Id:           true,
		DisplayName:  true,
		GivenName:    true,
		FamilyName:   true,
		EmailAddress: true,
		CreatedAt:    true,
		ModifiedAt:   true,

		// Users: Deny
		Oid: false,
	},

	// Businesses table permissions
	BusinessesAccessPolicy: &model.BusinessesAccessPolicy{
		UserCanAccess: false,
	},

	// Farm Fields table permissions
	FarmFieldsAccessPolicy: &model.FarmFieldsAccessPolicy{
		UserCanAccess: false,
	},
}
