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

type TableColumnDefinition[T querymodel.TableModel] = map[string]templates.ColumnData[T]

var userTableColumns = map[string]templates.ColumnData[*model.UsersModel]{
	model.ColUsersId: {
		Path:        ColPath(model.ColUsersId),
		DisplayName: model.ColDisplayNameUsersId,
		GetData: func(m *model.UsersModel) string {
			return m.Id
		},
	},
	model.ColUsersGivenName: {
		Path:        ColPath(model.ColUsersGivenName),
		DisplayName: model.ColDisplayNameUsersGivenName,
		GetData: func(m *model.UsersModel) string {
			return m.GivenName
		},
	},
	model.ColUsersFamilyName: {
		Path:        ColPath(model.ColUsersFamilyName),
		DisplayName: model.ColDisplayNameUsersFamilyName,
		GetData: func(m *model.UsersModel) string {
			return m.FamilyName
		},
	},
	model.ColUsersDisplayName: {
		Path:        ColPath(model.ColUsersDisplayName),
		DisplayName: model.ColDisplayNameUsersDisplayName,
		GetData: func(m *model.UsersModel) string {
			return m.DisplayName
		},
	},
	model.ColUsersEmailAddress: {
		Path:        ColPath(model.ColUsersEmailAddress),
		DisplayName: model.ColDisplayNameUsersEmailAddress,
		GetData: func(m *model.UsersModel) string {
			return m.EmailAddress
		},
	},
	model.ColUsersCreatedAt: {
		Path:        ColPath(model.ColUsersCreatedAt),
		DisplayName: model.ColDisplayNameUsersCreatedAt,
		GetData: func(m *model.UsersModel) string {
			return m.CreatedAt.Format(dateFormat)
		},
	},
	model.ColUsersModifiedAt: {
		Path:        ColPath(model.ColUsersModifiedAt),
		DisplayName: model.ColDisplayNameUsersModifiedAt,
		GetData: func(m *model.UsersModel) string {
			return m.CreatedAt.Format(dateFormat)
		},
	},
}

var businessTableColumns = map[string]templates.ColumnData[*model.BusinessesModel]{
	model.ColBusinessesId: {
		Path:        ColPath(model.ColBusinessesId),
		DisplayName: model.ColDisplayNameBusinessesId,
		GetData: func(m *model.BusinessesModel) string {
			return m.Id
		},
	},
	model.ColBusinessesReference: {
		Path:        ColPath(model.ColBusinessesReference),
		DisplayName: model.ColDisplayNameBusinessesReference,
		GetData: func(m *model.BusinessesModel) string {
			return m.Reference
		},
	},
	model.ColBusinessesTradingName: {
		Path:        ColPath(model.ColBusinessesTradingName),
		DisplayName: model.ColDisplayNameBusinessesTradingName,
		GetData: func(m *model.BusinessesModel) string {
			return m.TradingName
		},
	},
	model.ColBusinessesBusinessType: {
		Path:        ColPath(model.ColBusinessesBusinessType),
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
		Path:        ColPath(model.ColBusinessesCphNumber),
		DisplayName: model.ColDisplayNameBusinessesCphNumber,
		GetData: func(m *model.BusinessesModel) string {
			return m.CphNumber
		},
	},
	model.ColBusinessesEmailAddress: {
		Path:        ColPath(model.ColBusinessesEmailAddress),
		DisplayName: model.ColDisplayNameBusinessesEmailAddress,
		GetData: func(m *model.BusinessesModel) string {
			return m.EmailAddress
		},
	},
	model.ColBusinessesContactNumber: {
		Path:        ColPath(model.ColBusinessesContactNumber),
		DisplayName: model.ColDisplayNameBusinessesContactNumber,
		GetData: func(m *model.BusinessesModel) string {
			return m.ContactNumber
		},
	},
	model.ColBusinessesWebsiteUrl: {
		Path:        ColPath(model.ColBusinessesWebsiteUrl),
		DisplayName: model.ColDisplayNameBusinessesWebsiteUrl,
		GetData: func(m *model.BusinessesModel) string {
			return m.WebsiteUrl
		},
	},
	model.ColBusinessesAddressLine1: {
		Path:        ColPath(model.ColBusinessesAddressLine1),
		DisplayName: model.ColDisplayNameBusinessesAddressLine1,
		GetData: func(m *model.BusinessesModel) string {
			return m.AddressLine1
		},
	},
	model.ColBusinessesAddressLine2: {
		Path:        ColPath(model.ColBusinessesAddressLine2),
		DisplayName: model.ColDisplayNameBusinessesAddressLine2,
		GetData: func(m *model.BusinessesModel) string {
			return m.AddressLine2
		},
	},
	model.ColBusinessesTown: {
		Path:        ColPath(model.ColBusinessesTown),
		DisplayName: model.ColDisplayNameBusinessesTown,
		GetData: func(m *model.BusinessesModel) string {
			return m.Town
		},
	},
	model.ColBusinessesCounty: {
		Path:        ColPath(model.ColBusinessesCounty),
		DisplayName: model.ColDisplayNameBusinessesCounty,
		GetData: func(m *model.BusinessesModel) string {
			return m.County
		},
	},
	model.ColBusinessesCountry: {
		Path:        ColPath(model.ColBusinessesCountry),
		DisplayName: model.ColDisplayNameBusinessesCountry,
		GetData: func(m *model.BusinessesModel) string {
			return m.Country
		},
	},
	model.ColBusinessesPostcode: {
		Path:        ColPath(model.ColBusinessesPostcode),
		DisplayName: model.ColDisplayNameBusinessesPostcode,
		GetData: func(m *model.BusinessesModel) string {
			return m.Postcode
		},
	},
	model.ColBusinessesCreatedAt: {
		DisplayName: model.ColDisplayNameBusinessesCreatedAt,
		GetData: func(m *model.BusinessesModel) string {
			return m.CreatedAt.Format(dateFormat)
		},
		Path: ColPath(model.ColBusinessesCreatedAt),
	},
	model.RelColBusinessesCreatedById: {
		Path:        ColPath(model.RelColBusinessesCreatedById, model.ColUsersDisplayName),
		DisplayName: model.RelColDisplayNameBusinessesCreatedById,
		Expansion:   SingleLevelExpand(model.RelColBusinessesCreatedById, model.ColUsersDisplayName),
		GetData: func(m *model.BusinessesModel) string {
			return m.CreatedBy.DisplayName
		},
	},
	model.ColBusinessesModifiedAt: {
		Path:        ColPath(model.ColBusinessesModifiedAt),
		DisplayName: model.ColDisplayNameBusinessesModifiedAt,
		GetData: func(m *model.BusinessesModel) string {
			return m.ModifiedAt.Format(dateFormat)
		},
	},
	model.RelColBusinessesModifiedById: {
		Path:        ColPath(model.RelColBusinessesModifiedById, model.ColUsersDisplayName),
		DisplayName: model.RelColDisplayNameBusinessesModifiedById,
		Expansion:   SingleLevelExpand(model.RelColBusinessesModifiedById, model.ColUsersDisplayName),
		GetData: func(m *model.BusinessesModel) string {
			return m.ModifiedBy.DisplayName
		},
	},
}
