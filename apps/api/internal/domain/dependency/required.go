package dependency

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var ErrRequired = errors.New("domain dependency is required")

func Must(name string, value any) {
	if !IsNil(value) {
		return
	}

	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		normalizedName = "unnamed dependency"
	}

	panic(fmt.Errorf("%w: %s", ErrRequired, normalizedName))
}

func IsNil(value any) bool {
	if value == nil {
		return true
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer,
		reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
