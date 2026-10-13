package handlers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
	"github.com/turnerbenjamin/querystack/querymodel"
)

var dateFormat = "2 Jan 06 15:04"

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

var columnDataUsersUserId = templates.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersId),
	DisplayName: model.ColDisplayNameUsersId,
	GetData: func(m *model.UsersModel) string {
		return m.Id
	},
}

var columnDataUsersGivenName = templates.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersGivenName),
	DisplayName: model.ColDisplayNameUsersGivenName,
	GetData: func(m *model.UsersModel) string {
		return m.GivenName
	},
}

var columnDataUsersFamilyName = templates.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersFamilyName),
	DisplayName: model.ColDisplayNameUsersFamilyName,
	GetData: func(m *model.UsersModel) string {
		return m.FamilyName
	},
}

var columnDataUsersDisplayName = templates.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersDisplayName),
	DisplayName: model.ColDisplayNameUsersDisplayName,
	GetData: func(m *model.UsersModel) string {
		return m.DisplayName
	},
}

var columnDataUsersEmailAddress = templates.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersEmailAddress),
	DisplayName: model.ColDisplayNameUsersEmailAddress,
	GetData: func(m *model.UsersModel) string {
		return m.EmailAddress
	},
}

var columnDataUsersCreatedAt = templates.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersCreatedAt),
	DisplayName: model.ColDisplayNameUsersCreatedAt,
	GetData: func(m *model.UsersModel) string {
		return m.CreatedAt.Format(dateFormat)
	},
}

var columnDataUsersModifiedAt = templates.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersModifiedAt),
	DisplayName: model.ColDisplayNameUsersModifiedAt,
	GetData: func(m *model.UsersModel) string {
		return m.CreatedAt.Format(dateFormat)
	},
}

var userTableColumns = map[string]templates.ColumnData[*model.UsersModel]{
	columnDataUsersUserId.Path:       columnDataUsersUserId,
	columnDataUsersGivenName.Path:    columnDataUsersGivenName,
	columnDataUsersFamilyName.Path:   columnDataUsersFamilyName,
	columnDataUsersDisplayName.Path:  columnDataUsersDisplayName,
	columnDataUsersEmailAddress.Path: columnDataUsersEmailAddress,
	columnDataUsersCreatedAt.Path:    columnDataUsersCreatedAt,
	columnDataUsersModifiedAt.Path:   columnDataUsersModifiedAt,
}
var columnDataBusinessesBusinessId = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesId),
	DisplayName: model.ColDisplayNameBusinessesId,
	GetData: func(m *model.BusinessesModel) string {
		return m.Id
	},
}

var columnDataBusinessesReference = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesReference),
	DisplayName: model.ColDisplayNameBusinessesReference,
	GetData: func(m *model.BusinessesModel) string {
		return m.Reference
	},
}

var columnDataBusinessesTradingName = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesTradingName),
	DisplayName: model.ColDisplayNameBusinessesTradingName,
	GetData: func(m *model.BusinessesModel) string {
		return m.TradingName
	},
}

var columnDataBusinessesBusinessType = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesBusinessType),
	DisplayName: model.ColDisplayNameBusinessesBusinessType,
	GetData: func(m *model.BusinessesModel) string {
		s, exists := m.BusinessType.String()
		if !exists {
			return "invalid value"
		}
		return s
	},
}

var columnDataBusinessesCphNumber = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesCphNumber),
	DisplayName: model.ColDisplayNameBusinessesCphNumber,
	GetData: func(m *model.BusinessesModel) string {
		return m.CphNumber
	},
}

var columnDataBusinessesEmailAddress = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesEmailAddress),
	DisplayName: model.ColDisplayNameBusinessesEmailAddress,
	GetData: func(m *model.BusinessesModel) string {
		return m.EmailAddress
	},
}

var columnDataBusinessesContactNumber = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesContactNumber),
	DisplayName: model.ColDisplayNameBusinessesContactNumber,
	GetData: func(m *model.BusinessesModel) string {
		return m.ContactNumber
	},
}

var columnDataBusinessesWebsiteUrl = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesWebsiteUrl),
	DisplayName: model.ColDisplayNameBusinessesWebsiteUrl,
	GetData: func(m *model.BusinessesModel) string {
		return m.WebsiteUrl
	},
}

var columnDataBusinessesAddressLine1 = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesAddressLine1),
	DisplayName: model.ColDisplayNameBusinessesAddressLine1,
	GetData: func(m *model.BusinessesModel) string {
		return m.AddressLine1
	},
}

var columnDataBusinessesAddressLine2 = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesAddressLine2),
	DisplayName: model.ColDisplayNameBusinessesAddressLine2,
	GetData: func(m *model.BusinessesModel) string {
		return m.AddressLine2
	},
}

var columnDataBusinessesTown = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesTown),
	DisplayName: model.ColDisplayNameBusinessesTown,
	GetData: func(m *model.BusinessesModel) string {
		return m.Town
	},
}

var columnDataBusinessesCounty = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesCounty),
	DisplayName: model.ColDisplayNameBusinessesCounty,
	GetData: func(m *model.BusinessesModel) string {
		return m.County
	},
}

var columnDataBusinessesCountry = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesCountry),
	DisplayName: model.ColDisplayNameBusinessesCountry,
	GetData: func(m *model.BusinessesModel) string {
		return m.Country
	},
}

var columnDataBusinessesPostcode = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesPostcode),
	DisplayName: model.ColDisplayNameBusinessesPostcode,
	GetData: func(m *model.BusinessesModel) string {
		return m.Postcode
	},
}

var columnDataBusinessesCreatedAt = templates.ColumnData[*model.BusinessesModel]{
	DisplayName: model.ColDisplayNameBusinessesCreatedAt,
	GetData: func(m *model.BusinessesModel) string {
		return m.CreatedAt.Format(dateFormat)
	},
	Path: ColPath(model.ColBusinessesCreatedAt),
}

var columnDataBusinessesCreatedByIdUsersDisplayName = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.RelColBusinessesCreatedById, model.ColUsersDisplayName),
	DisplayName: model.RelColDisplayNameBusinessesCreatedById,
	Expansion:   SingleLevelExpand(model.RelColBusinessesCreatedById, model.ColUsersDisplayName),
	GetData: func(m *model.BusinessesModel) string {
		return m.CreatedBy.DisplayName
	},
}

var columnDataBusinessesModifiedAt = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesModifiedAt),
	DisplayName: model.ColDisplayNameBusinessesModifiedAt,
	GetData: func(m *model.BusinessesModel) string {
		return m.ModifiedAt.Format(dateFormat)
	},
}

var columnDataBusinessesModifiedByIdUsersDisplayName = templates.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.RelColBusinessesModifiedById, model.ColUsersDisplayName),
	DisplayName: model.RelColDisplayNameBusinessesModifiedById,
	Expansion:   SingleLevelExpand(model.RelColBusinessesModifiedById, model.ColUsersDisplayName),
	GetData: func(m *model.BusinessesModel) string {
		return m.ModifiedBy.DisplayName
	},
}

var businessTableColumns = map[string]templates.ColumnData[*model.BusinessesModel]{
	columnDataBusinessesBusinessId.Path:                   columnDataBusinessesBusinessId,
	columnDataBusinessesReference.Path:                    columnDataBusinessesReference,
	columnDataBusinessesTradingName.Path:                  columnDataBusinessesTradingName,
	columnDataBusinessesBusinessType.Path:                 columnDataBusinessesBusinessType,
	columnDataBusinessesCphNumber.Path:                    columnDataBusinessesCphNumber,
	columnDataBusinessesEmailAddress.Path:                 columnDataBusinessesEmailAddress,
	columnDataBusinessesContactNumber.Path:                columnDataBusinessesContactNumber,
	columnDataBusinessesWebsiteUrl.Path:                   columnDataBusinessesWebsiteUrl,
	columnDataBusinessesAddressLine1.Path:                 columnDataBusinessesAddressLine1,
	columnDataBusinessesAddressLine2.Path:                 columnDataBusinessesAddressLine2,
	columnDataBusinessesTown.Path:                         columnDataBusinessesTown,
	columnDataBusinessesCounty.Path:                       columnDataBusinessesCounty,
	columnDataBusinessesCountry.Path:                      columnDataBusinessesCountry,
	columnDataBusinessesPostcode.Path:                     columnDataBusinessesPostcode,
	columnDataBusinessesCreatedAt.Path:                    columnDataBusinessesCreatedAt,
	columnDataBusinessesCreatedByIdUsersDisplayName.Path:  columnDataBusinessesCreatedByIdUsersDisplayName,
	columnDataBusinessesModifiedAt.Path:                   columnDataBusinessesModifiedAt,
	columnDataBusinessesModifiedByIdUsersDisplayName.Path: columnDataBusinessesModifiedByIdUsersDisplayName,
	columnDataBusinessesModifiedByIdUsersDisplayName.Path: columnDataBusinessesModifiedByIdUsersDisplayName,
}
