// Package util (repo) contains repo util code.
package util

import (
	L "backend/internal/logging"
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/lib/pq"
)

// LogPqError logs entire pq.Error and returns code and message of an pq.Error
func LogPqError(ctx context.Context, err interface{}) (code, message string) {
	if pqerr, ok := err.(*pq.Error); ok {
		L.L.WithRequestID(ctx).Error("Database pq error", L.Any("pq.Error", &pqerr))
		code, message = string(pqerr.Code), pqerr.Message
	}
	return
}

func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func GetTag(entity interface{}, field string, tagname string) string {
	f, ok := reflect.TypeOf(entity).FieldByName(field)
	if !ok {
		L.L.Fatal("Field not found!")
	}
	t := f.Tag
	return t.Get(tagname)
}

func GetTagByIndex(entity interface{}, index int, tagname string) string {
	if reflect.ValueOf(entity).Kind() != reflect.Struct {
		//L.L.Fatal("Entity isn't a struct!")
		return ""
	}
	if reflect.ValueOf(entity).NumField() <= index {
		//L.L.Fatal("Index out of range!")
		return ""
	}
	f := reflect.TypeOf(entity).Field(index)
	t := f.Tag
	return t.Get(tagname)
}

func StartLowerCase(s string) bool {
	return s != "" && s[0] >= 'a' && s[0] <= 'z'
}

//Convert struct to json hash map, Name of json fields will be values of tag tagName
func ConvertStructToJSONHash(returnMap map[string]interface{}, structToConvert interface{}, tagName string) error {
	var myStruct interface{}
	if reflect.ValueOf(structToConvert).Kind() == reflect.Ptr {
		myStruct = reflect.ValueOf(structToConvert).Elem().Interface()
	} else {
		if reflect.ValueOf(structToConvert).Kind() == reflect.Struct {
			myStruct = structToConvert
		} else {
			return errors.New("Wrong data type!")
		}

	}
	structValue := reflect.ValueOf(myStruct)
	for i := 0; i < structValue.NumField(); i++ {
		if !structValue.Field(i).IsZero() {
			if structValue.Field(i).Kind() != reflect.Struct {
				key := GetTagByIndex(myStruct, i, tagName)
				if key != "" {
					returnMap[key] = structValue.Field(i).Interface()
				}
			} else {
				var err error
				for j := 0; j < reflect.ValueOf(structValue.Field(i).Interface()).NumField(); j++ {
					if StartLowerCase(reflect.ValueOf(structValue.Field(i).Interface()).Type().Field(j).Name) {
						returnMap[GetTagByIndex(myStruct, i, tagName)] = structValue.Field(i).Interface()
						err = errors.New("Unexported fields, struct isn't mine!")
					}
				}
				if err == nil {
					ConvertStructToJSONHash(returnMap, structValue.Field(i).Interface(), tagName)
				}
			}
		}
	}
	return nil

}

func isEditInfo(s string) bool {
	if len(s) > 8 && s[0:8] == "EditInfo" {
		return true
	}
	return false
}

//Generate objects that represent differences between objects old and new
func GenerateDiferences(old, new interface{}) (interface{}, interface{}, error) {
	if reflect.TypeOf(old) != reflect.TypeOf(new) {
		return nil, nil, errors.New("Data types are not same!")
	}
	tempOld, tempNew := old, new
	if reflect.ValueOf(old).Kind() == reflect.Ptr {
		tempOld, tempNew = reflect.ValueOf(old).Elem().Interface(), reflect.ValueOf(new).Elem().Interface()
	}
	oldDiff := reflect.New(reflect.TypeOf(tempOld))
	newDiff := reflect.New(reflect.TypeOf(tempNew))
	oldVal := reflect.ValueOf(tempOld)
	newVal := reflect.ValueOf(tempNew)

	for i := 0; i < oldVal.NumField(); i++ {
		oldField := oldVal.Field(i)
		newField := newVal.Field(i)
		if isEditInfo(oldVal.Type().Field(i).Name) && oldVal.Field(i).Kind() == reflect.Struct { //&& oldVal.Field(i).Kind() != reflect. {
			oldFieldTemp, newFieldTemp, err := GenerateDiferences(oldField.Interface(), newField.Interface())
			if err != nil {
				return nil, nil, err
			}
			oldDiff.Elem().Field(i).Set(reflect.ValueOf(oldFieldTemp))
			newDiff.Elem().Field(i).Set(reflect.ValueOf(newFieldTemp))
		} else {
			if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
				oldDiff.Elem().Field(i).Set(oldField)
				newDiff.Elem().Field(i).Set(newField)
			}
		}
	}
	return oldDiff.Elem().Interface(), newDiff.Elem().Interface(), nil
}
