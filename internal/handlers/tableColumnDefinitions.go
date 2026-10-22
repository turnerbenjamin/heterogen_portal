package handlers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turnerbenjamin/heterogen_portal/internal/model"
	"github.com/turnerbenjamin/heterogen_portal/internal/tables"
	"github.com/turnerbenjamin/querystack/querymodel"
)

var dateFormat = "2 Jan 06 15:04"

func ColPath(components ...string) string {
	return strings.Join(components, "/")
}

func SingleLevelExpand(localColumn, foreignColumn string) string {
	return fmt.Sprintf("%s(select=%s)", localColumn, foreignColumn)
}

func OrderedColumnData[T querymodel.TableModel](dataMap map[string]tables.ColumnData[T]) []tables.ColumnData[T] {
	keys := make([]string, 0, len(dataMap))
	for k := range dataMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := make([]tables.ColumnData[T], len(keys))
	for i, key := range keys {
		values[i] = dataMap[key]
	}
	return values
}

type TableColumnDefinition[T querymodel.TableModel] = map[string]tables.ColumnData[T]

var columnDataUsersUserId = tables.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersId),
	DisplayName: model.ColDisplayNameUsersId,
	GetData: func(m *model.UsersModel) string {
		return m.Id.String()
	},
}

var columnDataUsersGivenName = tables.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersGivenName),
	DisplayName: model.ColDisplayNameUsersGivenName,
	GetData: func(m *model.UsersModel) string {
		return m.GivenName
	},
}

var columnDataUsersFamilyName = tables.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersFamilyName),
	DisplayName: model.ColDisplayNameUsersFamilyName,
	GetData: func(m *model.UsersModel) string {
		return m.FamilyName
	},
}

var columnDataUsersDisplayName = tables.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersDisplayName),
	DisplayName: model.ColDisplayNameUsersDisplayName,
	GetData: func(m *model.UsersModel) string {
		return m.DisplayName
	},
}

var columnDataUsersEmailAddress = tables.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersEmailAddress),
	DisplayName: model.ColDisplayNameUsersEmailAddress,
	GetData: func(m *model.UsersModel) string {
		return m.EmailAddress
	},
}

var columnDataUsersCreatedAt = tables.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersCreatedAt),
	DisplayName: model.ColDisplayNameUsersCreatedAt,
	GetData: func(m *model.UsersModel) string {
		return m.CreatedAt.Format(dateFormat)
	},
}

var columnDataUsersModifiedAt = tables.ColumnData[*model.UsersModel]{
	Path:        ColPath(model.ColUsersModifiedAt),
	DisplayName: model.ColDisplayNameUsersModifiedAt,
	GetData: func(m *model.UsersModel) string {
		return m.CreatedAt.Format(dateFormat)
	},
}

var userTableColumns = map[string]tables.ColumnData[*model.UsersModel]{
	columnDataUsersUserId.Path:       columnDataUsersUserId,
	columnDataUsersGivenName.Path:    columnDataUsersGivenName,
	columnDataUsersFamilyName.Path:   columnDataUsersFamilyName,
	columnDataUsersDisplayName.Path:  columnDataUsersDisplayName,
	columnDataUsersEmailAddress.Path: columnDataUsersEmailAddress,
	columnDataUsersCreatedAt.Path:    columnDataUsersCreatedAt,
	columnDataUsersModifiedAt.Path:   columnDataUsersModifiedAt,
}

var columnDataBusinessesBusinessId = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesId),
	DisplayName: model.ColDisplayNameBusinessesId,
	GetData: func(m *model.BusinessesModel) string {
		return m.Id.String()
	},
}

var columnDataBusinessesReference = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesReference),
	DisplayName: model.ColDisplayNameBusinessesReference,
	GetData: func(m *model.BusinessesModel) string {
		return m.Reference
	},
}

var columnDataBusinessesTradingName = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesTradingName),
	DisplayName: model.ColDisplayNameBusinessesTradingName,
	GetData: func(m *model.BusinessesModel) string {
		return m.TradingName
	},
}

var columnDataBusinessesBusinessType = tables.ColumnData[*model.BusinessesModel]{
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

var columnDataBusinessesCphNumber = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesCphNumber),
	DisplayName: model.ColDisplayNameBusinessesCphNumber,
	GetData: func(m *model.BusinessesModel) string {
		return m.CphNumber
	},
}

var columnDataBusinessesEmailAddress = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesEmailAddress),
	DisplayName: model.ColDisplayNameBusinessesEmailAddress,
	GetData: func(m *model.BusinessesModel) string {
		return m.EmailAddress
	},
}

var columnDataBusinessesContactNumber = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesContactNumber),
	DisplayName: model.ColDisplayNameBusinessesContactNumber,
	GetData: func(m *model.BusinessesModel) string {
		return m.ContactNumber
	},
}

var columnDataBusinessesWebsiteUrl = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesWebsiteUrl),
	DisplayName: model.ColDisplayNameBusinessesWebsiteUrl,
	GetData: func(m *model.BusinessesModel) string {
		return m.WebsiteUrl
	},
}

var columnDataBusinessesAddressLine1 = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesAddressLine1),
	DisplayName: model.ColDisplayNameBusinessesAddressLine1,
	GetData: func(m *model.BusinessesModel) string {
		return m.AddressLine1
	},
}

var columnDataBusinessesAddressLine2 = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesAddressLine2),
	DisplayName: model.ColDisplayNameBusinessesAddressLine2,
	GetData: func(m *model.BusinessesModel) string {
		return m.AddressLine2
	},
}

var columnDataBusinessesTown = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesTown),
	DisplayName: model.ColDisplayNameBusinessesTown,
	GetData: func(m *model.BusinessesModel) string {
		return m.Town
	},
}

var columnDataBusinessesCounty = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesCounty),
	DisplayName: model.ColDisplayNameBusinessesCounty,
	GetData: func(m *model.BusinessesModel) string {
		return m.County
	},
}

var columnDataBusinessesCountry = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesCountry),
	DisplayName: model.ColDisplayNameBusinessesCountry,
	GetData: func(m *model.BusinessesModel) string {
		return m.Country
	},
}

var columnDataBusinessesPostcode = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesPostcode),
	DisplayName: model.ColDisplayNameBusinessesPostcode,
	GetData: func(m *model.BusinessesModel) string {
		return m.Postcode
	},
}

var columnDataBusinessesCreatedAt = tables.ColumnData[*model.BusinessesModel]{
	DisplayName: model.ColDisplayNameBusinessesCreatedAt,
	GetData: func(m *model.BusinessesModel) string {
		return m.CreatedAt.Format(dateFormat)
	},
	Path: ColPath(model.ColBusinessesCreatedAt),
}

var columnDataBusinessesCreatedByIdUsersDisplayName = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.RelColBusinessesCreatedById, model.ColUsersDisplayName),
	DisplayName: model.RelColDisplayNameBusinessesCreatedById,
	Expansion:   SingleLevelExpand(model.RelColBusinessesCreatedById, model.ColUsersDisplayName),
	GetData: func(m *model.BusinessesModel) string {
		return m.CreatedBy.DisplayName
	},
}

var columnDataBusinessesModifiedAt = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.ColBusinessesModifiedAt),
	DisplayName: model.ColDisplayNameBusinessesModifiedAt,
	GetData: func(m *model.BusinessesModel) string {
		return m.ModifiedAt.Format(dateFormat)
	},
}

var columnDataBusinessesModifiedByIdUsersDisplayName = tables.ColumnData[*model.BusinessesModel]{
	Path:        ColPath(model.RelColBusinessesModifiedById, model.ColUsersDisplayName),
	DisplayName: model.RelColDisplayNameBusinessesModifiedById,
	Expansion:   SingleLevelExpand(model.RelColBusinessesModifiedById, model.ColUsersDisplayName),
	GetData: func(m *model.BusinessesModel) string {
		return m.ModifiedBy.DisplayName
	},
}

var businessTableColumns = map[string]tables.ColumnData[*model.BusinessesModel]{
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
