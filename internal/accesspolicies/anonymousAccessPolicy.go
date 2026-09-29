package accesspolicies

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
)

// Query access policy for anonymous access
var AnonymousAccessPolicy = &model.DatabaseAccessPolicy{

	// Users table permissions
	UsersAccessPolicy: &model.UsersAccessPolicy{
		UserCanAccess: true,

		// Users: Allow
		Id:       true,
		UserName: true,

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
		Id:           true,
		Reference:    true,
		TradingName:  true,
		LogoUrl:      true,
		Description:  true,
		BusinessType: true,
		CphNumber:    true,
		WebsiteUrl:   true,

		// Businesses: Deny
		CreatedAt:     false,
		CreatedById:   false,
		ModifiedAt:    false,
		ModifiedById:  false,
		Location:      false,
		EmailAddress:  false,
		ContactNumber: false,
		AddressLine1:  false,
		AddressLine2:  false,
		Town:          false,
		County:        false,
		Country:       false,
		Postcode:      false,
	},

	// Farm Fields table permissions
	FarmFieldsAccessPolicy: &model.FarmFieldsAccessPolicy{
		UserCanAccess: true,

		// Farm Fields: Allow
		Id:         true,
		Reference:  true,
		BusinessId: true,
		Location:   true,

		// Farm Fields: Deny
		CreatedAt:    false,
		CreatedById:  false,
		ModifiedAt:   false,
		ModifiedById: false,
	},
}
