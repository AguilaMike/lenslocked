package utils

import (
	"reflect"
	"strconv"
)

func ConvertBoolCheckbox(value string) bool {
	if value == "on" {
		return reflect.ValueOf(true).Bool()
	} else if v, err := strconv.ParseBool(value); err == nil {
		return reflect.ValueOf(v).Bool()
	}

	return reflect.ValueOf(false).Bool()
}
