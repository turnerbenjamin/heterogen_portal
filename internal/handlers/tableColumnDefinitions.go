package handlers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
	"github.com/turnerbenjamin/querystack/querymodel"
)

var dateFormat = "02/01/06"

func ColPath(components ...string) string {
	return strings.Join(components, "/")
}

func SingleLevelExpand(localColumn, foreignColumn string) string {
	return fmt.Sprintf("%s(select=%s)", localColumn, foreignColumn)
}

func OrderedColumnData[T querymodel.TableModel](dataMap map[string]templates.ColumnData[T]) []templates.ColumnData[T] {
	keys := make([]string, 0, len(dataMap))
	for k := range dataMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := make([]templates.ColumnData[T], len(keys))
	for i, key := range keys {
		values[i] = dataMap[key]
	}
	return values
}

var userTableColumns = map[string]templates.ColumnData[*model.UsersModel]{
	model.ColUsersId: {
		ColumnName:  model.ColUsersId,
		DisplayName: model.ColDisplayNameUsersId,
		GetData: func(m *model.UsersModel) string {
			return m.Id
		},
	},
	model.ColUsersGivenName: {
		ColumnName:  model.ColUsersGivenName,
		DisplayName: model.ColDisplayNameUsersGivenName,
		GetData: func(m *model.UsersModel) string {
			return m.GivenName
		},
	},
	model.ColUsersFamilyName: {
		ColumnName:  model.ColUsersFamilyName,
		DisplayName: model.ColDisplayNameUsersFamilyName,
		GetData: func(m *model.UsersModel) string {
			return m.FamilyName
		},
	},
	model.ColUsersDisplayName: {
		ColumnName:  model.ColUsersDisplayName,
		DisplayName: model.ColDisplayNameUsersDisplayName,
		GetData: func(m *model.UsersModel) string {
			return m.DisplayName
		},
	},
	model.ColUsersEmailAddress: {
		ColumnName:  model.ColUsersEmailAddress,
		DisplayName: model.ColDisplayNameUsersEmailAddress,
		GetData: func(m *model.UsersModel) string {
			return m.EmailAddress
		},
	},
	model.ColUsersCreatedAt: {
		ColumnName:  model.ColUsersCreatedAt,
		DisplayName: model.ColDisplayNameUsersCreatedAt,
		GetData: func(m *model.UsersModel) string {
			return m.CreatedAt.Format(dateFormat)
		},
	},
	model.ColUsersModifiedAt: {
		ColumnName:  model.ColUsersModifiedAt,
		DisplayName: model.ColDisplayNameUsersModifiedAt,
		GetData: func(m *model.UsersModel) string {
			return m.CreatedAt.Format(dateFormat)
		},
	},
}

var businessTableColumns = map[string]templates.ColumnData[*model.BusinessesModel]{
	model.ColBusinessesId: {
		ColumnName:  model.ColBusinessesId,
		DisplayName: model.ColDisplayNameBusinessesId,
		GetData: func(m *model.BusinessesModel) string {
			return m.Id
		},
	},
	model.ColBusinessesReference: {
		ColumnName:  model.ColBusinessesReference,
		DisplayName: model.ColDisplayNameBusinessesReference,
		GetData: func(m *model.BusinessesModel) string {
			return m.Reference
		},
	},
	model.ColBusinessesTradingName: {
		ColumnName:  model.ColBusinessesTradingName,
		DisplayName: model.ColDisplayNameBusinessesTradingName,
		GetData: func(m *model.BusinessesModel) string {
			return m.TradingName
		},
	},
	model.ColBusinessesBusinessType: {
		ColumnName:  model.ColBusinessesBusinessType,
		DisplayName: model.ColDisplayNameBusinessesBusinessType,
		GetData: func(m *model.BusinessesModel) string {
			s, exists := m.BusinessType.String()
			if !exists {
				return "invalid value"
			}
			return s
		},
	},
	model.ColBusinessesCphNumber: {
		ColumnName:  model.ColBusinessesCphNumber,
		DisplayName: model.ColDisplayNameBusinessesCphNumber,
		GetData: func(m *model.BusinessesModel) string {
			return m.CphNumber
		},
	},
	model.ColBusinessesEmailAddress: {
		ColumnName:  model.ColBusinessesEmailAddress,
		DisplayName: model.ColDisplayNameBusinessesEmailAddress,
		GetData: func(m *model.BusinessesModel) string {
			return m.EmailAddress
		},
	},
	model.ColBusinessesContactNumber: {
		ColumnName:  model.ColBusinessesContactNumber,
		DisplayName: model.ColDisplayNameBusinessesContactNumber,
		GetData: func(m *model.BusinessesModel) string {
			return m.ContactNumber
		},
	},
	model.ColBusinessesWebsiteUrl: {
		ColumnName:  model.ColBusinessesWebsiteUrl,
		DisplayName: model.ColDisplayNameBusinessesWebsiteUrl,
		GetData: func(m *model.BusinessesModel) string {
			return m.WebsiteUrl
		},
	},
	model.ColBusinessesAddressLine1: {
		ColumnName:  model.ColBusinessesAddressLine1,
		DisplayName: model.ColDisplayNameBusinessesAddressLine1,
		GetData: func(m *model.BusinessesModel) string {
			return m.AddressLine1
		},
	},
	model.ColBusinessesAddressLine2: {
		ColumnName:  model.ColBusinessesAddressLine2,
		DisplayName: model.ColDisplayNameBusinessesAddressLine2,
		GetData: func(m *model.BusinessesModel) string {
			return m.AddressLine2
		},
	},
	model.ColBusinessesTown: {
		ColumnName:  model.ColBusinessesTown,
		DisplayName: model.ColDisplayNameBusinessesTown,
		GetData: func(m *model.BusinessesModel) string {
			return m.Town
		},
	},
	model.ColBusinessesCounty: {
		ColumnName:  model.ColBusinessesCounty,
		DisplayName: model.ColDisplayNameBusinessesCounty,
		GetData: func(m *model.BusinessesModel) string {
			return m.County
		},
	},
	model.ColBusinessesCountry: {
		ColumnName:  model.ColBusinessesCountry,
		DisplayName: model.ColDisplayNameBusinessesCountry,
		GetData: func(m *model.BusinessesModel) string {
			return m.Country
		},
	},
	model.ColBusinessesPostcode: {
		ColumnName:  model.ColBusinessesPostcode,
		DisplayName: model.ColDisplayNameBusinessesPostcode,
		GetData: func(m *model.BusinessesModel) string {
			return m.Postcode
		},
	},
	model.ColBusinessesCreatedAt: {
		ColumnName:  model.ColBusinessesCreatedAt,
		DisplayName: model.ColDisplayNameBusinessesCreatedAt,
		GetData: func(m *model.BusinessesModel) string {
			return m.CreatedAt.Format(dateFormat)
		},
	},
	model.RelColBusinessesCreatedById: {
		ColumnName:  model.ColUsersDisplayName,
		DisplayName: model.RelColDisplayNameBusinessesCreatedById,
		Path:        ColPath(model.RelColBusinessesCreatedById),
		Expansion:   SingleLevelExpand(model.RelColBusinessesCreatedById, model.ColUsersDisplayName),
		GetData: func(m *model.BusinessesModel) string {
			return m.CreatedBy.DisplayName
		},
	},
	model.ColBusinessesModifiedAt: {
		ColumnName:  model.ColBusinessesModifiedAt,
		DisplayName: model.ColDisplayNameBusinessesModifiedAt,
		Path:        ColPath(model.ColBusinessesModifiedAt),
		GetData: func(m *model.BusinessesModel) string {
			return m.ModifiedAt.Format(dateFormat)
		},
	},
	model.RelColBusinessesModifiedById: {
		ColumnName:  model.ColUsersDisplayName,
		DisplayName: model.RelColDisplayNameBusinessesModifiedById,
		Path:        ColPath(model.RelColBusinessesModifiedById),
		Expansion:   SingleLevelExpand(model.RelColBusinessesModifiedById, model.ColUsersDisplayName),
		GetData: func(m *model.BusinessesModel) string {
			return m.ModifiedBy.DisplayName
		},
	},
}
