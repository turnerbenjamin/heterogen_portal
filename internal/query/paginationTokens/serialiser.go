package paginationTokens

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"strings"

	qstore "github.com/turnerbenjamin/heterogen_portal/internal/query/queryDataStore"
	qerr "github.com/turnerbenjamin/heterogen_portal/internal/query/queryError"
	mdl "github.com/turnerbenjamin/heterogen_portal/internal/query/queryModel"
)

type serialiser struct {
	buf *bytes.Buffer
}

type deserialiser struct {
	sb           *strings.Builder
	buf          *bytes.Reader
	valueBuilder mdl.ValueBuilder
}

func serialiseToken(
	str qstore.QueryDataStore,
	version uint32,
	cursorValues []mdl.Value,
) ([]byte, error) {
	s := &serialiser{
		buf: new(bytes.Buffer),
	}

	// Write version
	s.SerialiseUint32(version)

	// Write cursor value count
	cursorValueCount := len(cursorValues)
	if cursorValueCount > math.MaxUint16 {
		return nil, qerr.InternalErr("max cursor value count exceeded: %d", math.MaxUint16)
	}
	s.SerialiseUint16(uint16(cursorValueCount))

	// Write cursor values
	if str.OrderByLen() != len(cursorValues) {
		return nil, qerr.InternalErr("mismatch between order by rules and cursor values")
	}

	for _, v := range cursorValues {
		if err := s.serialiseValueExpression(v); err != nil {
			return nil, err
		}
	}

	// Write resource name
	s.SerialiseString(str.RootResource().Name())

	// Write query string
	s.SerialiseString(str.QueryString())

	return s.buf.Bytes(), nil

}

func deserialiseToken(d []byte, v mdl.ValueBuilder) (PagingToken, error) {
	ds := &deserialiser{
		sb:           new(strings.Builder),
		buf:          bytes.NewReader(d),
		valueBuilder: v,
	}

	o := PagingToken{}

	// read version
	version, err := ds.DeserialiseUint32()
	if err != nil {
		return o, err
	}
	o.Version = version

	// read cursor value count
	cursorValueCount, err := ds.DeserialiseUint16()
	if err != nil {
		return o, err
	}

	// read cursor values
	o.CursorValues = make([]mdl.Value, cursorValueCount)
	for i := range cursorValueCount {
		v, err := ds.DeserialiseValueExpression()
		if err != nil {
			return o, err
		}
		o.CursorValues[i] = v
	}

	// read resource name
	resourceName, err := ds.ReadString()
	if err != nil {
		return o, err
	}
	o.ResourceName = resourceName

	// read query string
	queryString, err := ds.ReadString()
	if err != nil {
		return o, err
	}
	o.QueryString = queryString

	return o, nil
}

func (s *serialiser) serialiseValueExpression(v mdl.Value) error {
	// Write type
	_ = s.buf.WriteByte(uint8(v.Type()))

	// Reserve 2 bytes for content length
	contentLenPos := s.buf.Len()
	_ = s.buf.WriteByte(0)
	_ = s.buf.WriteByte(0)
	contentStart := s.buf.Len()

	// Serialise the content
	v.Serialise(s)

	// Compute content length
	contentLenRaw := s.buf.Len() - contentStart
	if contentLenRaw > math.MaxUint16 {
		return qerr.InternalErr("token content exceeds the maximum length %d", math.MaxUint16)
	}

	// Write content length in the reserved bytes
	contentLen := uint16(contentLenRaw)
	binary.BigEndian.PutUint16(s.buf.Bytes()[contentLenPos:], contentLen)

	return nil
}

func (ds *deserialiser) DeserialiseValueExpression() (mdl.Value, error) {
	// Read type
	typeByte, err := ds.buf.ReadByte()
	if err != nil {
		return nil, err
	}
	ltype := mdl.ValueType(typeByte)

	// Read content length
	b1, err := ds.buf.ReadByte()
	if err != nil {
		return nil, err
	}

	b2, err := ds.buf.ReadByte()
	if err != nil {
		return nil, err
	}
	contentLength := uint16(b1)<<8 | uint16(b2)

	// Read content bytes
	content := make([]byte, contentLength)
	if _, err := io.ReadFull(ds.buf, content); err != nil {
		return nil, qerr.InternalErr(
			"unable to read value expression content: %w",
			err,
		)
	}

	// parse content
	return ds.valueBuilder.ExecuteDeserialisation(ds, ltype, content)
}

func (s *serialiser) SerialiseString(str string) {
	// write string to buffer byte by byte
	for _, r := range str {
		b := byte(r)

		// break on null bytes, they are used as sentinals
		if b == 0 {
			break
		}

		s.buf.WriteByte(b)
	}
	s.buf.WriteByte(0)
}

func (ds *deserialiser) DeserialiseString(d []byte) string {
	ds.sb.Reset()
	for _, b := range d {
		if b == 0 {
			break
		}
		ds.sb.WriteByte(b)
	}
	return ds.sb.String()
}

func (ds *deserialiser) ReadString() (string, error) {
	ds.sb.Reset()
	for {
		b, err := ds.buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		if b == 0 {
			break
		}
		ds.sb.WriteByte(b)
	}
	return ds.sb.String(), nil
}

func (ds *deserialiser) DeserialiseListString(d []byte) []string {
	dLen := len(d)

	o := []string{}
	i := 0
	for i < dLen {
		// read string
		s := ds.DeserialiseString(d[i:])
		o = append(o, s)

		// move i past string and separator
		i += len(s) + 1
	}
	return o
}

func (s *serialiser) SerialiseInt(n int64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(n))
	_, _ = s.buf.Write(b[:])
}

func (ds *deserialiser) DeserialiseInt(d []byte) (int64, error) {
	if len(d) != 8 {
		return 0, qerr.InternalErr(
			"invalid int literal content length: got %d, expected 8",
			len(d),
		)
	}

	return int64(binary.BigEndian.Uint64(d)), nil
}

func (ds *deserialiser) DeserialiseListInt(d []byte) ([]int64, error) {
	// Validate data length
	dLen := len(d)
	if dLen%8 != 0 {
		return nil, qerr.InternalErr(
			"invalid int list content length: %d is not divisible by 8",
			dLen,
		)
	}

	// Initialise output slice
	listLen := dLen / 8
	o := make([]int64, listLen)

	// Loop through the data and parse the ints
	li := 0
	bi := 0
	for li < listLen {
		v, err := ds.DeserialiseInt(d[bi:])
		if err != nil {
			return nil, err
		}

		o[li] = v
		li++
		bi += 8
	}
	return o, nil
}

func (s *serialiser) SerialiseNull() {}

func (s *serialiser) SerialiseFloat(f float64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], math.Float64bits(f))
	_, _ = s.buf.Write(b[:])
}

func (ds *deserialiser) DeserialiseFloat(d []byte) (float64, error) {
	if len(d) != 8 {
		return 0, qerr.InternalErr(
			"invalid float literal content length: got %d, expected 8",
			len(d),
		)
	}
	return math.Float64frombits(binary.BigEndian.Uint64(d)), nil
}

func (ds *deserialiser) DeserialiseListFloat(d []byte) ([]float64, error) {
	// Validate data length
	dLen := len(d)
	if dLen%8 != 0 {
		return nil, qerr.InternalErr(
			"invalid float list content length: %d is not divisible by 8",
			dLen,
		)
	}

	// initialise output array
	listLen := dLen / 8
	o := make([]float64, listLen)

	// loop though the data and deserialiser the floats
	li := 0
	bi := 0
	for li < listLen {
		v, err := ds.DeserialiseFloat(d[bi:])
		if err != nil {
			return nil, err
		}

		o[li] = v
		li++
		bi += 8
	}
	return o, nil
}

func (s *serialiser) SerialiseUint32(n uint32) {
	var vb [4]byte
	binary.BigEndian.PutUint32(vb[:], querySchemaVersion)
	_, _ = s.buf.Write(vb[:])
}

func (ds *deserialiser) DeserialiseUint32() (uint32, error) {
	var versionBuf [4]byte
	if _, err := io.ReadFull(ds.buf, versionBuf[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(versionBuf[:]), nil
}

func (s *serialiser) SerialiseUint16(n uint16) {
	var vb [2]byte
	binary.BigEndian.PutUint16(vb[:], n)
	_, _ = s.buf.Write(vb[:])
}

func (ds *deserialiser) DeserialiseUint16() (uint16, error) {
	var versionBuf [2]byte
	if _, err := io.ReadFull(ds.buf, versionBuf[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(versionBuf[:]), nil
}

func (s *serialiser) SerialiseStringList(els []string) {
	for _, el := range els {
		s.SerialiseString(el)
	}
}

func (s *serialiser) SerialiseIntList(els []int64) {
	for _, el := range els {
		s.SerialiseInt(el)
	}
}

func (s *serialiser) SerialiseFloatList(els []float64) {
	for _, el := range els {
		s.SerialiseFloat(el)
	}
}
