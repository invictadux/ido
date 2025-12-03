package ido

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

// ---------------------------------------------------------
// INTERFACES
// ---------------------------------------------------------

// Marshaler allows types to customize their IDO encoding.
type Marshaler interface {
	MarshalIDO() ([]byte, error)
}

// ---------------------------------------------------------
// SHARED & CACHE
// ---------------------------------------------------------

// bufferPool reuses byte slices to prevent massive GC pressure during Marshal.
var bufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 512)
		return &b
	},
}

// timeType is shared across the package
var timeType = reflect.TypeOf(time.Time{})
var marshalerType = reflect.TypeOf((*Marshaler)(nil)).Elem()

// unsafeString converts []byte to string without allocation.
// Defined here and used by decode.go as well.
func unsafeString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

type encoderFunc func(b *[]byte, v reflect.Value) error

var encoderCache sync.Map // map[reflect.Type]encoderFunc

// ---------------------------------------------------------
// STREAMING API (Encoder)
// ---------------------------------------------------------

// Encoder writes IDO values to an output stream.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the IDO encoding of v to the stream.
func (e *Encoder) Encode(v any) error {
	bufPtr := bufferPool.Get().(*[]byte)
	*bufPtr = (*bufPtr)[:0]

	val := reflect.ValueOf(v)
	encoder, err := getEncoder(val.Type())
	if err != nil {
		bufferPool.Put(bufPtr)
		return err
	}

	if err := encoder(bufPtr, val); err != nil {
		bufferPool.Put(bufPtr)
		return err
	}

	*bufPtr = append(*bufPtr, '\n')

	if _, err := e.w.Write(*bufPtr); err != nil {
		bufferPool.Put(bufPtr)
		return err
	}

	bufferPool.Put(bufPtr)
	return nil
}

// ---------------------------------------------------------
// STANDARD API (Marshal)
// ---------------------------------------------------------

// Marshal encodes any struct into your custom format using pre-computed encoders.
func Marshal(v any) ([]byte, error) {
	bufPtr := bufferPool.Get().(*[]byte)
	*bufPtr = (*bufPtr)[:0]

	val := reflect.ValueOf(v)
	encoder, err := getEncoder(val.Type())
	if err != nil {
		bufferPool.Put(bufPtr)
		return nil, err
	}

	if err := encoder(bufPtr, val); err != nil {
		bufferPool.Put(bufPtr)
		return nil, err
	}

	result := make([]byte, len(*bufPtr))
	copy(result, *bufPtr)
	bufferPool.Put(bufPtr)

	return result, nil
}

// ---------------------------------------------------------
// COMPILER (Encoder)
// ---------------------------------------------------------

func getEncoder(t reflect.Type) (encoderFunc, error) {
	if f, ok := encoderCache.Load(t); ok {
		return f.(encoderFunc), nil
	}
	f, err := compileEncoder(t)
	if err != nil {
		return nil, err
	}
	encoderCache.Store(t, f)
	return f, nil
}

func compileEncoder(t reflect.Type) (encoderFunc, error) {
	// 1. Check for Marshaler interface
	if t.Implements(marshalerType) {
		return func(b *[]byte, v reflect.Value) error {
			if v.Kind() == reflect.Pointer && v.IsNil() {
				return nil
			}
			m, ok := v.Interface().(Marshaler)
			if !ok {
				// Should not happen if Implements check passed
				return fmt.Errorf("ido: value is not a Marshaler")
			}
			data, err := m.MarshalIDO()
			if err != nil {
				return err
			}
			*b = append(*b, data...)
			return nil
		}, nil
	}

	// 2. Standard types
	switch t.Kind() {
	case reflect.String:
		return encodeString, nil
	case reflect.Bool:
		return encodeBool, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return encodeInt, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return encodeUint, nil
	case reflect.Float32:
		return encodeFloat32, nil
	case reflect.Float64:
		return encodeFloat64, nil
	case reflect.Slice:
		return compileSliceEncoder(t)
	case reflect.Struct:
		if t == timeType {
			return encodeTime, nil
		}
		return compileStructEncoder(t)
	case reflect.Pointer:
		elemEnc, err := getEncoder(t.Elem())
		if err != nil {
			return nil, err
		}
		return func(b *[]byte, v reflect.Value) error {
			if v.IsNil() {
				return nil
			}
			return elemEnc(b, v.Elem())
		}, nil
	case reflect.Interface:
		return func(b *[]byte, v reflect.Value) error {
			if v.IsNil() {
				return nil
			}
			enc, err := getEncoder(v.Elem().Type())
			if err != nil {
				return err
			}
			return enc(b, v.Elem())
		}, nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", t)
	}
}

func compileStructEncoder(t reflect.Type) (encoderFunc, error) {
	type fieldInfo struct {
		idx     int
		encoder encoderFunc
	}
	var fields []fieldInfo

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get("ido") == "-" {
			continue
		}
		enc, err := compileEncoder(f.Type)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fieldInfo{idx: i, encoder: enc})
	}

	return func(b *[]byte, v reflect.Value) error {
		*b = append(*b, '{')
		for _, field := range fields {
			fv := v.Field(field.idx)
			if fv.IsZero() {
				*b = append(*b, ',')
				continue
			}
			if err := field.encoder(b, fv); err != nil {
				return err
			}
			*b = append(*b, ',')
		}

		slice := *b
		if len(slice) > 1 && slice[len(slice)-1] == ',' {
			slice[len(slice)-1] = '}'
		} else {
			*b = append(slice, '}')
		}
		return nil
	}, nil
}

func compileSliceEncoder(t reflect.Type) (encoderFunc, error) {
	elemEnc, err := compileEncoder(t.Elem())
	if err != nil {
		return nil, err
	}
	return func(b *[]byte, v reflect.Value) error {
		*b = append(*b, '[')
		l := v.Len()
		for i := 0; i < l; i++ {
			if err := elemEnc(b, v.Index(i)); err != nil {
				return err
			}
			*b = append(*b, ',')
		}

		slice := *b
		if len(slice) > 1 && slice[len(slice)-1] == ',' {
			slice[len(slice)-1] = ']'
		} else {
			*b = append(slice, ']')
		}
		return nil
	}, nil
}

// ---------------------------------------------------------
// PRIMITIVES (Encoder)
// ---------------------------------------------------------

func encodeString(b *[]byte, v reflect.Value) error {
	val := v.String()
	*b = append(*b, '"')
	for i := 0; i < len(val); i++ {
		c := val[i]
		if c == '"' {
			*b = append(*b, '\\', '"')
		} else if c == '\\' {
			*b = append(*b, '\\', '\\')
		} else {
			*b = append(*b, c)
		}
	}
	*b = append(*b, '"')
	return nil
}

func encodeBool(b *[]byte, v reflect.Value) error {
	if v.Bool() {
		*b = append(*b, '+')
	}
	return nil
}

func encodeInt(b *[]byte, v reflect.Value) error {
	*b = strconv.AppendInt(*b, v.Int(), 10)
	return nil
}

func encodeUint(b *[]byte, v reflect.Value) error {
	*b = strconv.AppendUint(*b, v.Uint(), 10)
	return nil
}

func encodeFloat32(b *[]byte, v reflect.Value) error {
	*b = strconv.AppendFloat(*b, v.Float(), 'f', -1, 32)
	return nil
}

func encodeFloat64(b *[]byte, v reflect.Value) error {
	*b = strconv.AppendFloat(*b, v.Float(), 'f', -1, 64)
	return nil
}

func encodeTime(b *[]byte, v reflect.Value) error {
	t := v.Interface().(time.Time)
	*b = strconv.AppendInt(*b, t.UnixMicro(), 10)
	return nil
}

func PrintFields(d any) {
	v := reflect.ValueOf(d)
	t := reflect.TypeOf(d)
	for i := 0; i < v.NumField(); i++ {
		tags := t.Field(i)
		if tag := tags.Tag.Get("ido"); tag == "-" {
			fmt.Printf("Name: %v, Tag: %v (IGNORED)\n", tags.Name, tag)
		} else if tag == "" {
			fmt.Println(tags.Name)
		} else {
			fmt.Printf("Name: %v, Tag: %v\n", tags.Name, tag)
		}
	}
}
