package utils

import (
	"fmt"
	"reflect"
)

func ToPointer[T any](v T) *T {
	if reflect.TypeOf(v).Kind() == reflect.String && fmt.Sprintf("%v", v) == "" {
		return nil
	}
	return &v
}
