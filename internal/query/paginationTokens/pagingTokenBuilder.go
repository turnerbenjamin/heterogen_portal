package paginationTokens

import (
	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

var querySchemaVersion uint32 = 1

type PayloadSigner interface {
	Sign(secret []byte, data []byte) string
	Verify(secret []byte, value string) (data []byte, ok bool)
}

type PagingToken struct {
	Version      uint32
	CursorValues []mdl.ValueExpression
	ResourceName string
	QueryString  string
}

type pagingTokenBuilder struct {
	payloadSigner PayloadSigner
	payloadSecret []byte
}

func NewNextPageTokenBuilder(payloadSigner PayloadSigner, payloadSecret []byte) (*pagingTokenBuilder, error) {
	if payloadSigner == nil {
		return nil, qerr.InternalErr("unable to build next page token. payload signer cannot be nil")
	}

	if payloadSecret == nil {
		return nil, qerr.InternalErr("unable to build next page token. payload secret cannot be nil")
	}

	return &pagingTokenBuilder{
		payloadSigner: payloadSigner,
		payloadSecret: payloadSecret,
	}, nil
}

func (b *pagingTokenBuilder) BuildToken(
	queryDataStore qstore.QueryDataStore,
	lastRecord mdl.TableModel,
) (string, error) {
	if b.payloadSigner == nil {
		return "", qerr.InternalErr("unable to build next page token. payload signer cannot be nil")
	}

	if b.payloadSecret == nil {
		return "", qerr.InternalErr("unable to build next page token. payload secret cannot be nil")
	}

	cursorValues, err := getCursorValues(queryDataStore, lastRecord)
	if err != nil {
		return "", err
	}

	payloadBytes, err := serialiseToken(
		queryDataStore,
		querySchemaVersion,
		cursorValues,
	)
	if err != nil {
		return "", err
	}
	return b.payloadSigner.Sign(b.payloadSecret, payloadBytes), nil
}

func getCursorValues(
	s qstore.QueryDataStore,
	lastRecord mdl.TableModel,
) ([]mdl.ValueExpression, error) {
	cursorValues := make([]mdl.ValueExpression, s.OrderByLen())

	i := 0
	for rule := range s.OrderBy() {
		// THIS IS GROSS - CHANGE SIGNATURE OF GetValueExpression !!!!!!!!!!!!!!
		nextRecordValue, err := lastRecord.GetValueExpression(
			rule.ResolvedColumn.ResolvedPath.Steps,
			rule.ResolvedColumn.Metadata.Name(),
			s.FilterExpressionBuilder().ValueBuilder(),
		)
		if err != nil {
			return nil, err
		}

		cursorValues[i] = nextRecordValue
	}
	return cursorValues, nil
}

func (b *pagingTokenBuilder) ParseToken(
	token string,
	valueBuilder mdl.ValueBuilder,
) (*PagingToken, error) {
	if b.payloadSigner == nil {
		return nil, qerr.InternalErr(
			"unable to build next page token. payload signer cannot be nil",
		)
	}

	if b.payloadSecret == nil {
		return nil, qerr.InternalErr(
			"unable to build next page token. payload secret cannot be nil",
		)
	}

	payloadBytes, ok := b.payloadSigner.Verify(b.payloadSecret, token)
	if !ok {
		return nil, qerr.NextPageTokenErr(
			"the next page token is invalid",
		)
	}

	tokenPayload, err := deserialiseToken(
		payloadBytes,
		valueBuilder,
	)
	if err != nil {
		return nil, err
	}

	if tokenPayload.Version != querySchemaVersion {
		return nil, qerr.NextPageTokenErr(
			"the next page token has expired",
		)
	}

	return &tokenPayload, nil
}
