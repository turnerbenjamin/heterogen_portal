package accesspolicies

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
)

// Query access policy for business table query access
var BusinessTableQueryAccessPolicy = &model.DatabaseAccessPolicy{

	// Users table permissions
	UsersAccessPolicy: &model.UsersAccessPolicy{
		UserCanAccess: true,

		// Users: Allow
		Id:          true,
		DisplayName: true,

		// Users: Deny
		Oid:          false,
		GivenName:    false,
		FamilyName:   false,
		EmailAddress: false,
		CreatedAt:    false,
		ModifiedAt:   false,
	},

	// Businesses table permissions
	BusinessesAccessPolicy: &model.BusinessesAccessPolicy{
		UserCanAccess: true,

		// Businesses: Allow
		Id:            true,
		Reference:     true,
		TradingName:   true,
		LogoUrl:       true,
		Description:   true,
		BusinessType:  true,
		CphNumber:     true,
		WebsiteUrl:    true,
		CreatedAt:     true,
		CreatedById:   true,
		ModifiedAt:    true,
		ModifiedById:  true,
		EmailAddress:  true,
		ContactNumber: true,
		AddressLine1:  true,
		AddressLine2:  true,
		Town:          true,
		County:        true,
		Country:       true,
		Postcode:      true,

		// Businesses: DENY
		Location: false,
	},

	// Farm Fields table permissions
	FarmFieldsAccessPolicy: &model.FarmFieldsAccessPolicy{
		UserCanAccess: false,
	},
}
