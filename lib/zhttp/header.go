package zhttp

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
)

var (
	_      Binder = (*header)(nil)
	Header        = &header{}
)

type header struct{}

func (h *header) Name() string {
	return "header"
}

func (h *header) Binding(src, dst any) error {
	switch x := src.(type) {
	case http.Header:
		return MapBindStruct(HttpHeaderMap(x), dst, h.Name())
	case map[string]any:
		return MapBindStruct(x, dst, h.Name())
	default:
		return errors.New(`src not  http header`)
	}
}

func HttpHeaderMap(h http.Header) map[string]any {
	m := make(map[string]any, len(h))
	for k := range h {
		m[strings.ToLower(k)] = h.Get(k)
	}
	return m
}

func TypeAndValue(x any) (reflect.Type, reflect.Value) {
	t, v := reflect.TypeOf(x), reflect.ValueOf(x)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	return t, v
}

func MapBindStruct(src map[string]any, dst any, tag string) error {
	yt, yv := TypeAndValue(dst)
	if yv.Kind() != reflect.Struct {
		return errors.New("dst not struct")
	}
	for i := 0; i < yv.NumField(); i++ {
		if !yv.Field(i).CanInterface() {
			continue
		}
		if j, ok := yt.Field(i).Tag.Lookup(tag); ok && j != "-" {
			if field, exists := src[strings.ToLower(j)]; exists {
				yv.FieldByName(yt.Field(i).Name).Set(reflect.ValueOf(field))
			}
		}
	}
	return nil
}
