package types

import (
	"reflect"
)

func ExtractData(data interface{}) (map[string]interface{}, error) {
	output := map[string]interface{}{}
	v := reflect.ValueOf(data)
	typeOfData := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := typeOfData.Field(i)
		jsonTag, ok := field.Tag.Lookup("json")
		if ok {
			output[jsonTag] = v.Field(i).Interface()
		}
	}
	return output, nil
}
