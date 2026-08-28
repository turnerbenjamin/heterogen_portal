package query

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"

	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

var querySchemaVersion uint32 = 1

type PayloadSigner interface {
	Sign(secret []byte, data []byte) string
	Verify(secret []byte, value string) (data []byte, ok bool)
}

type pagingToken struct {
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

	payloadBytes, err := writeToken(
		queryDataStore,
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
		)
		if err != nil {
			return nil, err
		}

		cursorValues[i] = nextRecordValue
	}
	return cursorValues, nil
}

func (b *pagingTokenBuilder) ParseToken(token string) (*pagingToken, error) {
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

	tokenPayload, err := readToken(bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}

	if tokenPayload.Version != querySchemaVersion {
		return nil, qerr.NextPageTokenErr(
			"the next page token has expired",
		)
	}

	return tokenPayload, nil
}

func writeToken(
	s qstore.QueryDataStore,
	cursorValues []mdl.ValueExpression,
) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write version
	var vb [4]byte
	binary.BigEndian.PutUint32(vb[:], querySchemaVersion)
	_, _ = buf.Write(vb[:])

	// Write cursor value count
	cursorValueCount := len(cursorValues)
	if cursorValueCount > math.MaxUint16 {
		return nil, qerr.InternalErr("max cursor value count exceeded: %d", math.MaxUint16)
	}

	var ccb [2]byte
	binary.BigEndian.PutUint16(ccb[:], uint16(cursorValueCount))
	_, _ = buf.Write(ccb[:])

	// Write cursor values
	if s.OrderByLen() != len(cursorValues) {
		return nil, qerr.InternalErr("mismatch between order by rules and cursor values")
	}

	for _, v := range cursorValues {
		if err := writeValueExpression(buf, v); err != nil {
			return nil, err
		}
	}

	// Write resource name
	writeNullTerminatedString(buf, s.RootResource().Name())

	// Write query string
	writeNullTerminatedString(buf, s.QueryString())
	return buf.Bytes(), nil
}

func readToken(r *bytes.Reader) (*pagingToken, error) {
	o := &pagingToken{}

	// read version
	var versionBuf [4]byte
	if _, err := io.ReadFull(r, versionBuf[:]); err != nil {
		return nil, err
	}
	o.Version = binary.BigEndian.Uint32(versionBuf[:])

	// read cursor value count
	var cursorValueCountBuff [2]byte
	if _, err := io.ReadFull(r, cursorValueCountBuff[:]); err != nil {
		return nil, err
	}
	cursorValueCount := binary.BigEndian.Uint16(cursorValueCountBuff[:])

	// read cursor values
	o.CursorValues = make([]mdl.ValueExpression, cursorValueCount)
	for i := range cursorValueCount {
		v, err := readValueExpression(r)
		if err != nil {
			return nil, err
		}
		o.CursorValues[i] = v
	}

	// read resource name
	rn, err := readNullTerminatedString(r)
	if err != nil {
		return nil, err
	}
	o.ResourceName = rn

	// read query string
	q, err := readNullTerminatedString(r)
	if err != nil {
		return nil, err
	}
	o.QueryString = q

	return o, nil
}

func writeNullTerminatedString(buf *bytes.Buffer, str string) {
	_, _ = buf.WriteString(str)
	_ = buf.WriteByte(0)
}

func readNullTerminatedString(r *bytes.Reader) (string, error) {
	var b []byte

	for {
		c, err := r.ReadByte()
		if err != nil {
			return "", err
		}

		if c == 0 {
			return string(b), nil
		}

		b = append(b, c)
	}
}

func writeValueExpression(buf *bytes.Buffer, expression mdl.ValueExpression) error {
	// Write type
	_ = buf.WriteByte(uint8(expression.GetType()))

	// Reserve 2 bytes for content length
	contentLenPos := buf.Len()
	_ = buf.WriteByte(0)
	_ = buf.WriteByte(0)
	contentStart := buf.Len()

	// Write content
	switch tex := expression.(type) {
	case *mdl.NullLiteral:
		// No content.

	case *mdl.StringLiteral:
		_, _ = buf.WriteString(tex.Value)

	case *mdl.IntLiteral:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(tex.Value))
		_, _ = buf.Write(b[:])

	case *mdl.FloatLiteral:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], math.Float64bits(tex.Value))
		_, _ = buf.Write(b[:])

	case *mdl.StringListLiteral:
		for i, s := range tex.Values {
			if i > 0 {
				_ = buf.WriteByte(0)
			}
			_, _ = buf.WriteString(s)
		}

	case *mdl.IntListLiteral:
		for _, n := range tex.Values {
			var b [8]byte
			binary.BigEndian.PutUint64(b[:], uint64(n))
			_, _ = buf.Write(b[:])
		}

	case *mdl.FloatListLiteral:
		for _, n := range tex.Values {
			var b [8]byte
			binary.BigEndian.PutUint64(b[:], math.Float64bits(n))
			_, _ = buf.Write(b[:])
		}

	default:
		return qerr.InternalErr(
			"unable to write value expression: unfamiliar type encountered",
		)
	}

	contentLenRaw := buf.Len() - contentStart
	if contentLenRaw > math.MaxUint16 {
		return qerr.InternalErr("token content exceeds the maximum length %d", math.MaxUint16)
	}

	// Content length.
	contentLen := uint16(contentLenRaw)
	binary.BigEndian.PutUint16(buf.Bytes()[contentLenPos:], contentLen)

	return nil
}

func readValueExpression(buf *bytes.Reader) (mdl.ValueExpression, error) {
	// Read type
	typeByte, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	ltype := mdl.LiteralType(typeByte)

	// Read content length
	b1, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	b2, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	contentLength := uint16(b1)<<8 | uint16(b2)

	// Read content bytes
	content := make([]byte, contentLength)
	if _, err := io.ReadFull(buf, content); err != nil {
		return nil, qerr.InternalErr(
			"unable to read value expression content: %w",
			err,
		)
	}

	// Parse content to value type
	switch ltype {
	case mdl.LiteralTypeNull:
		if len(content) != 0 {
			return nil, qerr.InternalErr(
				"invalid null literal content length: got %d, expected 0",
				len(content),
			)
		}
		return &mdl.NullLiteral{}, nil

	case mdl.LiteralTypeString:
		return &mdl.StringLiteral{Value: string(content)}, nil

	case mdl.LiteralTypeInt:
		if len(content) != 8 {
			return nil, qerr.InternalErr(
				"invalid int literal content length: got %d, expected 8",
				len(content),
			)
		}

		v := int64(binary.BigEndian.Uint64(content))
		return &mdl.IntLiteral{Value: v}, nil

	case mdl.LiteralTypeFloat:
		if len(content) != 8 {
			return nil, qerr.InternalErr(
				"invalid float literal content length: got %d, expected 8",
				len(content),
			)
		}

		v := math.Float64frombits(binary.BigEndian.Uint64(content))
		return &mdl.FloatLiteral{Value: v}, nil

	case mdl.LiteralTypeStringList:
		if len(content) == 0 {
			return &mdl.StringListLiteral{}, nil
		}

		// NUL is reserved as the string-list separator.
		values := bytes.Split(content, []byte{0})

		result := make([]string, len(values))
		for i, value := range values {
			result[i] = string(value)
		}

		return &mdl.StringListLiteral{Values: result}, nil

	case mdl.LiteralTypeIntList:
		if len(content)%8 != 0 {
			return nil, qerr.InternalErr(
				"invalid int list content length: %d is not divisible by 8",
				len(content),
			)
		}

		values := make([]int64, len(content)/8)

		for i := range values {
			offset := i * 8
			values[i] = int64(binary.BigEndian.Uint64(content[offset : offset+8]))
		}

		return &mdl.IntListLiteral{Values: values}, nil

	case mdl.LiteralTypeFloatList:
		if len(content)%8 != 0 {
			return nil, qerr.InternalErr(
				"invalid float list content length: %d is not divisible by 8",
				len(content),
			)
		}

		values := make([]float64, len(content)/8)

		for i := range values {
			offset := i * 8
			values[i] = math.Float64frombits(
				binary.BigEndian.Uint64(content[offset : offset+8]),
			)
		}

		return &mdl.FloatListLiteral{Values: values}, nil

	default:
		return nil, qerr.InternalErr(
			"unable to read value expression: unfamiliar type encountered",
		)
	}
}
