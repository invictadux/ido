package main

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"
)

var (
	comma          = []byte{','}
	openBrace      = []byte{'{'}
	closeBrace     = []byte{'}'}
	openBracket    = []byte{'['}
	closeBracket   = []byte{']'}
	quote          = []byte{'"'}
	plusComma      = []byte{'+', ','}
	emptyComma     = []byte{','}
	escapedQuote   = []byte(`\"`)
	timeType       = reflect.TypeOf(time.Time{})
	typeFieldCache sync.Map // map[reflect.Type][]int
)

// PrintFields is for debugging tags (unchanged)
func PrintFields(d any) {
	v := reflect.ValueOf(d)
	t := reflect.TypeOf(d)

	for i := 0; i < v.NumField(); i++ {
		tags := t.Field(i)
		if tag := tags.Tag.Get("gido"); tag == "-" {
			fmt.Printf("Name: %v, Tag: %v\n", tags.Name, tag)
		} else if tag == "" {
			fmt.Println(tags.Name)
		} else {
			fmt.Printf("Name: %v, Tag: %v\n", tags.Name, tag)
		}
	}
}

// Marshal encodes any struct into your custom format using []byte appending
func Marshal(s any) ([]byte, error) {
	out := make([]byte, 0, 256)
	out = append(out, '{')

	v := reflect.ValueOf(s)
	t := v.Type()

	fieldIndexesAny, ok := typeFieldCache.Load(t)
	var fieldIndexes []int
	if ok {
		fieldIndexes = fieldIndexesAny.([]int)
	} else {
		fieldIndexes = make([]int, 0, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).Tag.Get("gido") != "-" {
				fieldIndexes = append(fieldIndexes, i)
			} else {
				// still need to preserve position
				fieldIndexes = append(fieldIndexes, -1)
			}
		}
		typeFieldCache.Store(t, fieldIndexes)
	}

	for _, i := range fieldIndexes {
		if i == -1 {
			out = append(out, ',')
			continue
		}
		field := v.Field(i)
		if field.IsZero() {
			out = append(out, ',')
			continue
		}
		var err error
		out, err = marshalValue(out, field)
		if err != nil {
			return nil, err
		}
		out = append(out, ',')
	}

	if len(out) > 1 {
		out[len(out)-1] = '}'
	} else {
		out = append(out, '}')
	}
	return out, nil
}

func marshalValue(out []byte, v reflect.Value) ([]byte, error) {
	switch v.Kind() {
	case reflect.String:
		out = writeEscapedString(out, v.String())

	case reflect.Bool:
		if v.Bool() {
			out = append(out, '+')
		}

	case reflect.Slice:
		return marshalArray(out, v)

	case reflect.Struct:
		if v.Type() == timeType {
			out = strconv.AppendInt(out, v.Interface().(time.Time).UnixMicro(), 10)
		} else {
			b, err := Marshal(v.Interface())
			if err != nil {
				return nil, err
			}
			out = append(out, b...)
		}

	default:
		if v.IsZero() {
			return out, nil
		}
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			out = strconv.AppendInt(out, v.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			out = strconv.AppendUint(out, v.Uint(), 10)
		case reflect.Float32:
			out = strconv.AppendFloat(out, v.Float(), 'f', -1, 32)
		case reflect.Float64:
			out = strconv.AppendFloat(out, v.Float(), 'f', -1, 64)
		default:
			return nil, fmt.Errorf("unsupported kind: %s", v.Kind())
		}
	}
	return out, nil
}

func marshalArray(out []byte, v reflect.Value) ([]byte, error) {
	out = append(out, '[')
	for i := 0; i < v.Len(); i++ {
		var err error
		out, err = marshalValue(out, v.Index(i))
		if err != nil {
			return nil, err
		}
		out = append(out, ',')
	}
	if len(out) > 0 && out[len(out)-1] == ',' {
		out[len(out)-1] = ']'
	} else {
		out = append(out, ']')
	}
	return out, nil
}

func writeEscapedString(dst []byte, s string) []byte {
	dst = append(dst, '"')
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			dst = append(dst, '\\', '"')
		} else {
			dst = append(dst, s[i])
		}
	}
	dst = append(dst, '"')
	return dst
}
