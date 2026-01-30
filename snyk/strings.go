package snyk

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"
)

var timeType = reflect.TypeFor[time.Time]()

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func Stringify(message any) string {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
	}()

	v := reflect.ValueOf(message)
	stringifyValue(buf, v)
	return buf.String()
}

// stringifyValue is heavily inspired by the go-github.
func stringifyValue(w *bytes.Buffer, val reflect.Value) {
	if val.Kind() == reflect.Pointer && val.IsNil() {
		w.WriteString("<nil>")
		return
	}

	v := reflect.Indirect(val)

	switch v.Kind() {
	case reflect.Bool:
		w.Write(strconv.AppendBool(w.Bytes(), v.Bool())[w.Len():])
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		w.Write(strconv.AppendInt(w.Bytes(), v.Int(), 10)[w.Len():])
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		w.Write(strconv.AppendUint(w.Bytes(), v.Uint(), 10)[w.Len():])
	case reflect.Float32:
		w.Write(strconv.AppendFloat(w.Bytes(), v.Float(), 'g', -1, 32)[w.Len():])
	case reflect.Float64:
		w.Write(strconv.AppendFloat(w.Bytes(), v.Float(), 'g', -1, 64)[w.Len():])
	case reflect.String:
		w.WriteByte('"')
		w.WriteString(v.String())
		w.WriteByte('"')
	case reflect.Slice:
		w.WriteByte('[')
		for i := range v.Len() {
			if i > 0 {
				w.WriteByte(' ')
			}

			stringifyValue(w, v.Index(i))
		}

		w.WriteByte(']')
		return
	case reflect.Struct:
		if v.Type().Name() != "" {
			w.WriteString(v.Type().String())
		}

		// special handling of Time values
		if v.Type() == timeType {
			_, _ = fmt.Fprintf(w, "{%v}", v.Interface())
			return
		}

		w.WriteByte('{')

		var sep bool
		for i := range v.NumField() {
			fv := v.Field(i)
			if fv.Kind() == reflect.Pointer && fv.IsNil() {
				continue
			}
			if fv.Kind() == reflect.Slice && fv.IsNil() {
				continue
			}
			if fv.Kind() == reflect.Map && fv.IsNil() {
				continue
			}

			if sep {
				w.WriteString(", ")
			} else {
				sep = true
			}

			w.WriteString(v.Type().Field(i).Name)
			w.WriteByte(':')
			stringifyValue(w, fv)
		}

		w.WriteByte('}')
	default:
		if v.CanInterface() {
			_, _ = fmt.Fprint(w, v.Interface())
		}
	}
}
