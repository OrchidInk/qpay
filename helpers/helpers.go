package helpers

import (
	"os"
	"reflect"
	"strings"
	"time"
)

type ContextKey string

func TrimStringsInStruct(s interface{}) {
	val := reflect.ValueOf(s).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		trimTag := fieldType.Tag.Get("trim")
		if field.Kind() == reflect.Struct {
			TrimStringsInStruct(field.Addr().Interface())
		}

		if trimTag == "true" && field.Kind() == reflect.String {
			trimmed := strings.TrimSpace(field.String())
			field.SetString(trimmed)
		}
	}
}

func ConvertDatetimeToTimezone(datetime time.Time) (time.Time, error) {
	tz, err := time.LoadLocation(os.Getenv("TIMEZONE"))
	if err != nil {
		return time.Time{}, err
	}

	return datetime.In(tz), nil
}
