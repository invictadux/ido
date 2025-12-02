package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// !
// "
// #
// $
// %
// &
// '
// (
// )
// *
// + bool -> true
// , none -> data separator symbol
// - bool -> false
// .
// /
// :
// ;
// <
// =
// >
// ?
// @
// _
// `
// { struct -> start of struct
// |
// } struct -> end of struct
// ~
// [ slice -> start of array
// ] slice -> end of array

// data types

// Bool +
// Int +
// Int8 +
// Int16 +
// Int32 +
// Int64 +
// Uint +
// Uint8 +
// Uint16 +
// Uint32 +
// Uint64 +
// Uintptr
// Float32 +
// Float64 +
// Complex64
// Complex128
// Array
// Chan
// Func
// Interface
// Map +
// Pointer
// Slice +
// String +
// Struct +
// UnsafePointer

/*
The data types used are

Number: a signed decimal number that may contain a fractional part.
The format makes distinction between integer and floating-point.

String: the same exact characters, may be excaped

Boolean: - false and + true

Array: can make distinction between string, int and float arrays,
all values in the array have to be the same.

Struct: a collection of values.

null: nothing is used, can be seen as comma and another comma
*/

func Validate(v string, mut reflect.Value, i *int) error {
	if len(v) == 0 {
		return nil
	}

	t := mut.Type().Field(*i)

	switch k := t.Type.Kind(); k {
	case reflect.String:
		mut.FieldByName(t.Name).SetString(v[1 : len(v)-1])
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(v, 10, 64)

		if err != nil {
			return err
		}

		mut.FieldByName(t.Name).SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(v, 10, 32)

		if err != nil {
			return err
		}

		mut.FieldByName(t.Name).SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(v, 64)

		if err != nil {
			return err
		}

		mut.FieldByName(t.Name).SetFloat(n)
	case reflect.Bool:
		switch v {
		case "":
			mut.FieldByName(t.Name).SetBool(false)
		case "+":
			mut.FieldByName(t.Name).SetBool(true)
		default:
			return fmt.Errorf("expecting '' or + found: %v", v)
		}
	case reflect.Slice:
		err := UnmarshalSlice(mut.Field(*i), v)

		if err != nil {
			return err
		}
	default:
		if t.Type == reflect.TypeOf(time.Time{}) {
			n, _ := strconv.ParseInt(v, 10, 64)
			mut.FieldByName(t.Name).Set(reflect.ValueOf(time.UnixMicro(n)))
		}
	}

	return nil
}

func UnmarshalSliceString(f reflect.Value, data string) error {
	i := 0
	start_i := 0
	inString := false
	data_len := len(data)
	arr := []string{}

	for i < data_len {
		switch data[i] {
		case ',', ']':
			if !inString {
				arr = append(arr, data[start_i+1:i-1])
			}
		case '"':
			if !inString {
				inString = true
				start_i = i
			} else {
				inString = false
			}
		case '\\':
			i++
		}

		i++
	}

	f.Set(reflect.ValueOf(arr))
	return nil
}

func UnmarshalSliceStruct(f reflect.Value, data string) error {
	i := 1
	start_i := 1
	inStruct := 0
	inString := false
	data_len := len(data)
	arr := f

	for i < data_len {
		switch data[i] {
		case ',', ']':
			if inStruct == 0 && !inString {
				ev := data[start_i:i]
				c := reflect.New(f.Type().Elem()).Interface()
				err := Unmarshal([]byte(ev), c)

				if err != nil {
					return err
				}

				arr = reflect.Append(arr, reflect.ValueOf(c).Elem())
			}
		case '"':
			if !inString {
				inString = true
			} else {
				inString = false
			}
		case '\\':
			i++
		case '{':
			if !inString {
				if inStruct == 0 {
					start_i = i
				}

				inStruct++
			}
		case '}':
			if !inString {
				inStruct--
			}
		}

		i++
	}

	f.Set(arr)
	return nil
}

func UnmarshalSliceSlice(f reflect.Value, data string) error {
	i := 0
	start_i := 0
	inSlice := 0
	inString := false
	data_len := len(data)
	arr := f

	for i < data_len {
		switch data[i] {
		case ',':
			if inSlice == 0 && !inString {
				ev := data[start_i:i]
				c := reflect.New(f.Type().Elem()).Elem()
				err := UnmarshalSlice(c, ev)

				if err != nil {
					return err
				}

				arr = reflect.Append(arr, c)
			}
		case '"':
			if !inString {
				inString = true
			} else {
				inString = false
			}
		case '\\':
			i++
		case '[':
			if !inString {
				if inSlice == 0 {
					start_i = i
				}

				inSlice++
			}
		case ']':
			if !inString {
				inSlice--
			}
		}
		i++
	}

	if i+1 >= data_len {
		ev := data[start_i:data_len]
		c := reflect.New(f.Type().Elem()).Elem()
		err := UnmarshalSlice(c, ev)

		if err != nil {
			return err
		}

		arr = reflect.Append(arr, c)
	}

	f.Set(arr)
	return nil
}

func UnmarshalSlice(f reflect.Value, v string) error {
	switch f.Type().Elem().Kind() {
	case reflect.String:
		err := UnmarshalSliceString(f, v)

		if err != nil {
			return err
		}
	case reflect.Bool:
		slice := []bool{}

		for _, e := range strings.Split(v[1:len(v)-1], ",") {
			switch e {
			case "-":
				slice = append(slice, false)
			case "+":
				slice = append(slice, true)
			default:
				return fmt.Errorf("can't convert %v to bool", e)
			}
		}

		f.Set(reflect.ValueOf(slice))
	case reflect.Int:
		slice := []int{}

		for _, e := range strings.Split(v[1:len(v)-1], ",") {
			n, err := strconv.ParseInt(e, 10, 64)

			if err != nil {
				return err
			}

			slice = append(slice, int(n))
		}

		f.Set(reflect.ValueOf(slice))
	case reflect.Int8:
		slice := []int8{}

		for _, e := range strings.Split(v[1:len(v)-1], ",") {
			n, err := strconv.ParseInt(e, 10, 64)

			if err != nil {
				return err
			}

			slice = append(slice, int8(n))
		}

		f.Set(reflect.ValueOf(slice))
	case reflect.Int16:
		slice := []int16{}

		for _, e := range strings.Split(v[1:len(v)-1], ",") {
			n, err := strconv.ParseInt(e, 10, 64)

			if err != nil {
				return err
			}

			slice = append(slice, int16(n))
		}

		f.Set(reflect.ValueOf(slice))
	case reflect.Int32:
		slice := []int32{}

		for _, e := range strings.Split(v[1:len(v)-1], ",") {
			n, err := strconv.ParseInt(e, 10, 64)

			if err != nil {
				return err
			}

			slice = append(slice, int32(n))
		}

		f.Set(reflect.ValueOf(slice))
	case reflect.Int64:
		slice := []int64{}

		for _, e := range strings.Split(v[1:len(v)-1], ",") {
			n, err := strconv.ParseInt(e, 10, 64)

			if err != nil {
				return err
			}

			slice = append(slice, n)
		}

		f.Set(reflect.ValueOf(slice))
	case reflect.Uint:
		slice := []uint{}

		for _, e := range strings.Split(v[1:len(v)-1], ",") {
			n, err := strconv.ParseUint(e, 10, 32)

			if err != nil {
				return err
			}

			slice = append(slice, uint(n))
		}

		f.Set(reflect.ValueOf(slice))
	case reflect.Uint8:
		slice := []uint8{}

		for _, e := range strings.Split(v[1:len(v)-1], ",") {
			n, err := strconv.ParseUint(e, 10, 32)

			if err != nil {
				return err
			}

			slice = append(slice, uint8(n))
		}

		f.Set(reflect.ValueOf(slice))
	case reflect.Uint16:
		slice := []uint16{}

		for _, e := range strings.Split(v[1:len(v)-1], ",") {
			n, err := strconv.ParseUint(e, 10, 32)

			if err != nil {
				return err
			}

			slice = append(slice, uint16(n))
		}

		f.Set(reflect.ValueOf(slice))
	case reflect.Uint32:
		slice := []uint32{}

		for _, e := range strings.Split(v[1:len(v)-1], ",") {
			n, err := strconv.ParseUint(e, 10, 32)

			if err != nil {
				return err
			}

			slice = append(slice, uint32(n))
		}

		f.Set(reflect.ValueOf(slice))
	case reflect.Uint64:
		slice := []uint64{}

		for _, e := range strings.Split(v[1:len(v)-1], ",") {
			n, err := strconv.ParseUint(e, 10, 32)

			if err != nil {
				return err
			}

			slice = append(slice, n)
		}

		f.Set(reflect.ValueOf(slice))
	case reflect.Float32:
		slice := []float32{}

		for _, e := range strings.Split(v[1:len(v)-1], ",") {
			n, err := strconv.ParseFloat(e, 64)

			if err != nil {
				return err
			}

			slice = append(slice, float32(n))
		}

		f.Set(reflect.ValueOf(slice))
	case reflect.Float64:
		slice := []float64{}

		for _, e := range strings.Split(v[1:len(v)-1], ",") {
			n, err := strconv.ParseFloat(e, 64)

			if err != nil {
				return err
			}

			slice = append(slice, n)
		}

		f.Set(reflect.ValueOf(slice))
	case reflect.Struct:
		err := UnmarshalSliceStruct(f, v)

		if err != nil {
			return err
		}
	case reflect.Slice:
		err := UnmarshalSliceSlice(f, v[1:len(v)-1])

		if err != nil {
			return err
		}
	}

	return nil
}

func Unmarshal(data []byte, s any) error {
	mut := reflect.ValueOf(s).Elem()
	t := reflect.TypeOf(s).Elem()

	i := 1
	elm := 0
	start_i := 1
	inString := false
	isEscaped := false
	inSlice := 0
	inStruct := 0
	data_length := len(data) - 1

	for i < data_length {
		tags := t.Field(elm)
		if tag := tags.Tag.Get("gido"); tag == "-" {
			elm++
			continue
		}

		switch data[i] {
		case ',':
			if inSlice == 0 && inStruct == 0 && !inString {
				ev := data[start_i:i]

				err := Validate(string(ev), mut, &elm)

				if err != nil {
					return err
				}

				start_i = i + 1
				elm++
			}
		case '"':
			if inStruct == 0 && inSlice == 0 {
				if isEscaped {
					isEscaped = false
					i++
					continue
				}

				if !inString {
					start_i = i
					inString = true
				} else {
					inString = false
				}
			}
		case '\\':
			isEscaped = true
		case '{':
			if !inString && inSlice == 0 {
				if inStruct == 0 {
					start_i = i
				}
				inStruct++
			}
		case '}':
			if inStruct > 0 {
				inStruct--
			}
		case '[':
			if inStruct == 0 && !inString {
				inSlice++
			}
		case ']':
			if inStruct == 0 && !inString {
				inSlice--
			}
		}

		if i+1 >= data_length {
			ev := data[start_i:data_length]
			err := Validate(string(ev), mut, &elm)

			if err != nil {
				return err
			}
		}

		i++
	}

	return nil
}
