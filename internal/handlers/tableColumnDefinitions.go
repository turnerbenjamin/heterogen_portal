package handlers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/querystack/querymodel"
)

var dateFormat = "02/01/06"

func ColPath(components ...string) string {
	return strings.Join(components, "/")
}

func SingleLevelExpand(localColumn, foreignColumn string) string {
	return fmt.Sprintf("%s(select=%s)", localColumn, foreignColumn)
}

func OrderedColumnData[T querymodel.TableModel](dataMap map[string]ColumnData[T]) []ColumnData[T] {
	keys := make([]string, 0, len(dataMap))
	for k := range dataMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := make([]ColumnData[T], len(keys))
	for i, key := range keys {
		values[i] = dataMap[key]
	}
	return values
}

var userTableColumns = map[string]ColumnData[*model.UsersModel]{
	model.ColUsersId: {
		ColumnName:  model.ColUsersId,
		DisplayName: model.ColDisplayNameUsersId,
		DataType:    model.ColTypeUsersId,
		GetData: func(m *model.UsersModel) string {
			return m.Id
		},
	},
	model.ColUsersGivenName: {
		ColumnName:  model.ColUsersGivenName,
		DisplayName: model.ColDisplayNameUsersGivenName,
		DataType:    model.ColTypeUsersGivenName,
		GetData: func(m *model.UsersModel) string {
			return m.GivenName
		},
	},
	model.ColUsersFamilyName: {
		ColumnName:  model.ColUsersFamilyName,
		DisplayName: model.ColDisplayNameUsersFamilyName,
		DataType:    model.ColTypeUsersFamilyName,
		GetData: func(m *model.UsersModel) string {
			return m.FamilyName
		},
	},
	model.ColUsersUserName: {
		ColumnName:  model.ColUsersUserName,
		DisplayName: model.ColDisplayNameUsersUserName,
		DataType:    model.ColTypeUsersUserName,
		GetData: func(m *model.UsersModel) string {
			return m.UserName
		},
	},
	model.ColUsersEmailAddress: {
		ColumnName:  model.ColUsersEmailAddress,
		DisplayName: model.ColDisplayNameUsersEmailAddress,
		DataType:    model.ColTypeUsersEmailAddress,
		GetData: func(m *model.UsersModel) string {
			return m.EmailAddress
		},
	},
	model.ColUsersCreatedAt: {
		ColumnName:  model.ColUsersCreatedAt,
		DisplayName: model.ColDisplayNameUsersCreatedAt,
		DataType:    model.ColTypeUsersCreatedAt,
		GetData: func(m *model.UsersModel) string {
			return m.CreatedAt.Format(dateFormat)
		},
	},
	model.ColUsersModifiedAt: {
		ColumnName:  model.ColUsersModifiedAt,
		DisplayName: model.ColDisplayNameUsersModifiedAt,
		DataType:    model.ColTypeUsersModifiedAt,
		GetData: func(m *model.UsersModel) string {
			return m.CreatedAt.Format(dateFormat)
		},
	},
}

var businessTableColumns = map[string]ColumnData[*model.BusinessesModel]{
	model.ColBusinessesId: {
		ColumnName:  model.ColBusinessesId,
		DisplayName: model.ColDisplayNameBusinessesId,
		DataType:    model.ColTypeBusinessesId,
		GetData: func(m *model.BusinessesModel) string {
			return m.Id
		},
	},
	model.ColBusinessesReference: {
		ColumnName:  model.ColBusinessesReference,
		DisplayName: model.ColDisplayNameBusinessesReference,
		DataType:    model.ColTypeBusinessesReference,
		GetData: func(m *model.BusinessesModel) string {
			return m.Reference
		},
	},
	model.ColBusinessesTradingName: {
		ColumnName:  model.ColBusinessesTradingName,
		DisplayName: model.ColDisplayNameBusinessesTradingName,
		DataType:    model.ColTypeBusinessesTradingName,
		GetData: func(m *model.BusinessesModel) string {
			return m.TradingName
		},
	},
	model.ColBusinessesBusinessType: {
		ColumnName:  model.ColBusinessesBusinessType,
		DisplayName: model.ColDisplayNameBusinessesBusinessType,
		DataType:    model.ColTypeBusinessesBusinessType,
		GetData: func(m *model.BusinessesModel) string {
			return model.BusinessType[m.BusinessType]
		},
	},
	model.ColBusinessesCphNumber: {
		ColumnName:  model.ColBusinessesCphNumber,
		DisplayName: model.ColDisplayNameBusinessesCphNumber,
		DataType:    model.ColTypeBusinessesCphNumber,
		GetData: func(m *model.BusinessesModel) string {
			return m.CphNumber
		},
	},
	model.ColBusinessesEmailAddress: {
		ColumnName:  model.ColBusinessesEmailAddress,
		DisplayName: model.ColDisplayNameBusinessesEmailAddress,
		DataType:    model.ColTypeBusinessesEmailAddress,
		GetData: func(m *model.BusinessesModel) string {
			return m.EmailAddress
		},
	},
	model.ColBusinessesContactNumber: {
		ColumnName:  model.ColBusinessesContactNumber,
		DisplayName: model.ColDisplayNameBusinessesContactNumber,
		DataType:    model.ColTypeBusinessesContactNumber,
		GetData: func(m *model.BusinessesModel) string {
			return m.ContactNumber
		},
	},
	model.ColBusinessesWebsiteUrl: {
		ColumnName:  model.ColBusinessesWebsiteUrl,
		DisplayName: model.ColDisplayNameBusinessesWebsiteUrl,
		DataType:    model.ColTypeBusinessesWebsiteUrl,
		GetData: func(m *model.BusinessesModel) string {
			return m.WebsiteUrl
		},
	},
	model.ColBusinessesAddressLine1: {
		ColumnName:  model.ColBusinessesAddressLine1,
		DisplayName: model.ColDisplayNameBusinessesAddressLine1,
		DataType:    model.ColTypeBusinessesAddressLine1,
		GetData: func(m *model.BusinessesModel) string {
			return m.AddressLine1
		},
	},
	model.ColBusinessesAddressLine2: {
		ColumnName:  model.ColBusinessesAddressLine2,
		DisplayName: model.ColDisplayNameBusinessesAddressLine2,
		DataType:    model.ColTypeBusinessesAddressLine2,
		GetData: func(m *model.BusinessesModel) string {
			return m.AddressLine2
		},
	},
	model.ColBusinessesTown: {
		ColumnName:  model.ColBusinessesTown,
		DisplayName: model.ColDisplayNameBusinessesTown,
		DataType:    model.ColTypeBusinessesTown,
		GetData: func(m *model.BusinessesModel) string {
			return m.Town
		},
	},
	model.ColBusinessesCounty: {
		ColumnName:  model.ColBusinessesCounty,
		DisplayName: model.ColDisplayNameBusinessesCounty,
		DataType:    model.ColTypeBusinessesCounty,
		GetData: func(m *model.BusinessesModel) string {
			return m.County
		},
	},
	model.ColBusinessesCountry: {
		ColumnName:  model.ColBusinessesCountry,
		DisplayName: model.ColDisplayNameBusinessesCountry,
		DataType:    model.ColTypeBusinessesCountry,
		GetData: func(m *model.BusinessesModel) string {
			return m.Country
		},
	},
	model.ColBusinessesPostcode: {
		ColumnName:  model.ColBusinessesPostcode,
		DisplayName: model.ColDisplayNameBusinessesPostcode,
		DataType:    model.ColTypeBusinessesPostcode,
		GetData: func(m *model.BusinessesModel) string {
			return m.Postcode
		},
	},
	model.ColBusinessesCreatedAt: {
		ColumnName:  model.ColBusinessesCreatedAt,
		DisplayName: model.ColDisplayNameBusinessesCreatedAt,
		DataType:    model.ColTypeBusinessesCreatedAt,
		GetData: func(m *model.BusinessesModel) string {
			return m.CreatedAt.Format(dateFormat)
		},
	},
	model.RelColBusinessesCreatedById: {
		ColumnName:  model.ColUsersUserName,
		DisplayName: model.RelColDisplayNameBusinessesCreatedById,
		DataType:    model.ColTypeBusinessesCreatedById,
		Path:        ColPath(model.RelColBusinessesCreatedById),
		Expansion:   SingleLevelExpand(model.RelColBusinessesCreatedById, model.ColUsersUserName),
		GetData: func(m *model.BusinessesModel) string {
			return m.CreatedBy.UserName
		},
	},
	model.ColBusinessesModifiedAt: {
		ColumnName:  model.ColBusinessesModifiedAt,
		DisplayName: model.ColDisplayNameBusinessesModifiedAt,
		DataType:    model.ColTypeBusinessesModifiedAt,
		Path:        ColPath(model.ColBusinessesModifiedAt),
		GetData: func(m *model.BusinessesModel) string {
			return m.ModifiedAt.Format(dateFormat)
		},
	},
	model.RelColBusinessesModifiedById: {
		ColumnName:  model.ColUsersUserName,
		DisplayName: model.RelColDisplayNameBusinessesModifiedById,
		DataType:    model.ColTypeBusinessesModifiedById,
		Path:        ColPath(model.RelColBusinessesModifiedById),
		Expansion:   SingleLevelExpand(model.RelColBusinessesModifiedById, model.ColUsersUserName),
		GetData: func(m *model.BusinessesModel) string {
			return m.ModifiedBy.UserName
		},
	},
}
