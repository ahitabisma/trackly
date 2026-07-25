package validatorx

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func getJSONFieldName(s interface{}, structField string) string {
	t := reflect.TypeOf(s)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	field, ok := t.FieldByName(structField)
	if !ok {
		return strings.ToLower(structField)
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return strings.ToLower(structField)
	}

	name := strings.Split(jsonTag, ",")[0]
	return name
}

func ValidateStruct(s interface{}) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)

	for _, e := range err.(validator.ValidationErrors) {
		field := getJSONFieldName(s, e.Field())

		switch e.Tag() {
		case "required":
			errors[field] = field + " is required"

		case "email":
			errors[field] = field + " must be a valid email"

		case "min":
			errors[field] = field + " minimum is " + e.Param() + " length"

		default:
			errors[field] = field + " is invalid"
		}
	}

	return errors
}
