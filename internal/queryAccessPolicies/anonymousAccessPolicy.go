package queryAccessPolicies

import (
	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

var allowAccess = model.ColumnAccessPolicy{UserCanAccess: true}
var denyAccess = model.ColumnAccessPolicy{UserCanAccess: false}

// Query access policy for anonymous access
var anonymousAccessPolicy = &model.DatabaseAccessPolicy{

	// Users table permissions
	UsersAccessPolicy: &model.UsersAccessPolicy{
		UserCanAccess: true,

		// Users: Allow
		Id:       allowAccess,
		UserName: allowAccess,

		// Users: Deny
		Oid:          denyAccess,
		GivenName:    denyAccess,
		FamilyName:   denyAccess,
		EmailAddress: denyAccess,
		CreatedAt:    denyAccess,
		ModifiedAt:   denyAccess,
	},

	// Businesses table permissions
	BusinessesAccessPolicy: &model.BusinessesAccessPolicy{
		UserCanAccess: true,

		// Businesses: Allow
		Id:           allowAccess,
		Reference:    allowAccess,
		TradingName:  allowAccess,
		LogoUrl:      allowAccess,
		Description:  allowAccess,
		BusinessType: allowAccess,
		CphNumber:    allowAccess,
		WebsiteUrl:   allowAccess,
		CreatedAt:    allowAccess,
		CreatedById:  allowAccess,
		ModifiedAt:   allowAccess,
		ModifiedById: allowAccess,
		Location:     allowAccess,

		// Businesses: Deny
		EmailAddress:  denyAccess,
		ContactNumber: denyAccess,
		AddressLine1:  denyAccess,
		AddressLine2:  denyAccess,
		Town:          denyAccess,
		County:        denyAccess,
		Country:       denyAccess,
		Postcode:      denyAccess,
	},

	// Farm Fields table permissions
	FarmFieldsAccessPolicy: &model.FarmFieldsAccessPolicy{
		UserCanAccess: true,

		// Farm Fields: Allow
		Id:           allowAccess,
		Reference:    allowAccess,
		BusinessId:   allowAccess,
		CreatedAt:    allowAccess,
		CreatedById:  allowAccess,
		ModifiedAt:   allowAccess,
		ModifiedById: allowAccess,
		Location:     allowAccess,

		// Farm Fields: Deny
	},
}

func GetAnonymousAccessPolicy() queryModel.AccessPolicy {
	return anonymousAccessPolicy
}
