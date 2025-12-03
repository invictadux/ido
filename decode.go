package ido

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"sync"
	"time"
)

// ---------------------------------------------------------
// INTERFACES
// ---------------------------------------------------------

// Unmarshaler allows types to customize their IDO decoding.
type Unmarshaler interface {
	UnmarshalIDO([]byte) error
}

// ---------------------------------------------------------
// CACHE
// ---------------------------------------------------------

type decoderFunc func(d []byte, v reflect.Value) error

var decoderCache sync.Map // map[reflect.Type]decoderFunc
var unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

// NOTE: timeType and unsafeString are defined in encode.go and shared.

// ---------------------------------------------------------
// STREAMING API (Decoder)
// ---------------------------------------------------------

// Decoder reads IDO values from an input stream.
type Decoder struct {
	r   *bufio.Reader
	buf []byte
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:   bufio.NewReader(r),
		buf: make([]byte, 0, 1024),
	}
}

func (d *Decoder) Decode(v any) error {
	token, err := d.nextObject()
	if err != nil {
		return err
	}
	return Unmarshal(token, v)
}

func (d *Decoder) nextObject() ([]byte, error) {
	for {
		advance, valid := scanObjectBoundary(d.buf)
		if valid {
			obj := d.buf[:advance]

			// Copy is required because Unmarshal (and custom unmarshalers)
			// might retain the slice data.
			result := make([]byte, advance)
			copy(result, obj)

			copy(d.buf, d.buf[advance:])
			d.buf = d.buf[:len(d.buf)-advance]

			if len(result) > 0 && result[len(result)-1] == '\n' {
				result = result[:len(result)-1]
			}

			return result, nil
		}

		if len(d.buf) == cap(d.buf) {
			newBuf := make([]byte, len(d.buf), cap(d.buf)*2)
			copy(newBuf, d.buf)
			d.buf = newBuf
		}

		n, err := d.r.Read(d.buf[len(d.buf):cap(d.buf)])
		if n > 0 {
			d.buf = d.buf[:len(d.buf)+n]
		}
		if err != nil {
			if err == io.EOF {
				if len(d.buf) > 0 {
					return nil, io.ErrUnexpectedEOF
				}
				return nil, io.EOF
			}
			return nil, err
		}
	}
}

func scanObjectBoundary(data []byte) (int, bool) {
	if len(data) == 0 {
		return 0, false
	}

	start := 0
	for start < len(data) && (data[start] == ' ' || data[start] == '\n' || data[start] == '\r' || data[start] == '\t') {
		start++
	}
	if start == len(data) {
		return start, false
	}

	depth := 0
	arrDepth := 0
	inQuote := false
	isEscaped := false

	for i := start; i < len(data); i++ {
		b := data[i]

		if inQuote {
			if isEscaped {
				isEscaped = false
			} else if b == '\\' {
				isEscaped = true
			} else if b == '"' {
				inQuote = false
			}
			continue
		}

		switch b {
		case '{':
			depth++
		case '}':
			depth--
		case '[':
			arrDepth++
		case ']':
			arrDepth--
		case '"':
			inQuote = true
		}

		if depth == 0 && arrDepth == 0 {
			return i + 1, true
		}
	}

	return 0, false
}

// ---------------------------------------------------------
// STANDARD API (Unmarshal)
// ---------------------------------------------------------

func Unmarshal(data []byte, s any) error {
	rv := reflect.ValueOf(s)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("ido: Unmarshal(non-pointer %v)", reflect.TypeOf(s))
	}

	val := rv.Elem()
	decoder, err := getDecoder(val.Type())
	if err != nil {
		return err
	}

	return decoder(data, val)
}

// ---------------------------------------------------------
// COMPILER (Decoder)
// ---------------------------------------------------------

func getDecoder(t reflect.Type) (decoderFunc, error) {
	if f, ok := decoderCache.Load(t); ok {
		return f.(decoderFunc), nil
	}
	f, err := compileDecoder(t)
	if err != nil {
		return nil, err
	}
	decoderCache.Store(t, f)
	return f, nil
}

func compileDecoder(t reflect.Type) (decoderFunc, error) {
	// 1. Check if type T implements Unmarshaler
	if t.Implements(unmarshalerType) {
		return func(d []byte, v reflect.Value) error {
			if v.Kind() == reflect.Pointer && v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			return v.Interface().(Unmarshaler).UnmarshalIDO(d)
		}, nil
	}

	// 2. Check if *T implements Unmarshaler (when we have T)
	if t.Kind() != reflect.Pointer && reflect.PointerTo(t).Implements(unmarshalerType) {
		return func(d []byte, v reflect.Value) error {
			if !v.CanAddr() {
				return fmt.Errorf("ido: cannot unmarshal into unaddressable value")
			}
			return v.Addr().Interface().(Unmarshaler).UnmarshalIDO(d)
		}, nil
	}

	// 3. Standard types
	switch t.Kind() {
	case reflect.String:
		return decodeString, nil
	case reflect.Bool:
		return decodeBool, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return decodeInt, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return decodeUint, nil
	case reflect.Float32, reflect.Float64:
		return decodeFloat, nil
	case reflect.Slice:
		return compileSliceDecoder(t)
	case reflect.Struct:
		if t == timeType {
			return decodeTime, nil
		}
		return compileStructDecoder(t)
	case reflect.Pointer:
		elemDec, err := getDecoder(t.Elem())
		if err != nil {
			return nil, err
		}
		return func(d []byte, v reflect.Value) error {
			if len(d) == 0 {
				return nil
			}
			if v.IsNil() {
				v.Set(reflect.New(t.Elem()))
			}
			return elemDec(d, v.Elem())
		}, nil
	case reflect.Interface:
		return func(d []byte, v reflect.Value) error { return nil }, nil
	default:
		return nil, fmt.Errorf("unsupported type for decoding: %s", t)
	}
}

func compileStructDecoder(t reflect.Type) (decoderFunc, error) {
	type fieldInfo struct {
		idx     int
		decoder decoderFunc
	}
	var fields []fieldInfo

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get("ido") == "-" {
			continue
		}
		dec, err := compileDecoder(f.Type)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fieldInfo{idx: i, decoder: dec})
	}

	return func(data []byte, v reflect.Value) error {
		if len(data) < 2 {
			return nil
		}
		content := data[1 : len(data)-1]
		if len(content) == 0 {
			return nil
		}

		for _, field := range fields {
			if len(content) == 0 {
				break
			}

			token, advance := nextToken(content)

			if len(token) > 0 {
				if err := field.decoder(token, v.Field(field.idx)); err != nil {
					return err
				}
			}

			if advance >= len(content) {
				content = nil
			} else {
				content = content[advance:]
			}
		}
		return nil
	}, nil
}

func compileSliceDecoder(t reflect.Type) (decoderFunc, error) {
	elemDec, err := compileDecoder(t.Elem())
	if err != nil {
		return nil, err
	}

	return func(data []byte, v reflect.Value) error {
		if len(data) < 2 {
			return nil
		}
		content := data[1 : len(data)-1]
		if len(content) == 0 {
			return nil
		}

		v.SetLen(0)

		for len(content) > 0 {
			token, advance := nextToken(content)

			newElem := reflect.New(t.Elem()).Elem()
			if len(token) > 0 {
				if err := elemDec(token, newElem); err != nil {
					return err
				}
			}
			v.Set(reflect.Append(v, newElem))

			if advance >= len(content) {
				break
			}
			content = content[advance:]
		}
		return nil
	}, nil
}

func nextToken(data []byte) (token []byte, advance int) {
	depth := 0
	arrDepth := 0
	inQuote := false
	isEscaped := false

	for i, b := range data {
		if inQuote {
			if isEscaped {
				isEscaped = false
			} else if b == '\\' {
				isEscaped = true
			} else if b == '"' {
				inQuote = false
			}
			continue
		}

		switch b {
		case '{':
			depth++
		case '}':
			depth--
		case '[':
			arrDepth++
		case ']':
			arrDepth--
		case '"':
			inQuote = true
		case ',':
			if depth == 0 && arrDepth == 0 {
				return data[:i], i + 1
			}
		}
	}
	return data, len(data)
}

// ---------------------------------------------------------
// PRIMITIVES (Decoder)
// ---------------------------------------------------------

func decodeString(d []byte, v reflect.Value) error {
	if len(d) >= 2 && d[0] == '"' && d[len(d)-1] == '"' {
		v.SetString(unescape(d[1 : len(d)-1]))
	} else {
		v.SetString(unsafeString(d))
	}
	return nil
}

func decodeBool(d []byte, v reflect.Value) error {
	if len(d) == 1 && d[0] == '+' {
		v.SetBool(true)
	} else {
		v.SetBool(false)
	}
	return nil
}

func decodeInt(d []byte, v reflect.Value) error {
	if len(d) == 0 {
		return nil
	}
	n, err := strconv.ParseInt(unsafeString(d), 10, 64)
	if err != nil {
		return err
	}
	v.SetInt(n)
	return nil
}

func decodeUint(d []byte, v reflect.Value) error {
	if len(d) == 0 {
		return nil
	}
	n, err := strconv.ParseUint(unsafeString(d), 10, 64)
	if err != nil {
		return err
	}
	v.SetUint(n)
	return nil
}

func decodeFloat(d []byte, v reflect.Value) error {
	if len(d) == 0 {
		return nil
	}
	n, err := strconv.ParseFloat(unsafeString(d), 64)
	if err != nil {
		return err
	}
	v.SetFloat(n)
	return nil
}

func decodeTime(d []byte, v reflect.Value) error {
	if len(d) == 0 {
		return nil
	}
	n, err := strconv.ParseInt(unsafeString(d), 10, 64)
	if err != nil {
		return nil
	}
	v.Set(reflect.ValueOf(time.UnixMicro(n).UTC()))
	return nil
}

func unescape(b []byte) string {
	hasEscape := false
	for _, c := range b {
		if c == '\\' {
			hasEscape = true
			break
		}
	}
	if !hasEscape {
		return string(b)
	}
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		if b[i] == '\\' && i+1 < len(b) && b[i+1] == '"' {
			out = append(out, '"')
			i++
		} else {
			out = append(out, b[i])
		}
	}
	return string(out)
}
