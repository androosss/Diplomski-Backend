// SPORTOS is a service that provides payment routing of depositing or withdrawing actions.
// It provides communication with different payment providers and with PAM (Player Account Manager) service.
package sportos

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type ContextKey string

const (
	CONTEXT_USER_ID_KEY           = ContextKey("UserID")
	CONTEXT_PAM_TOKEN_KEY         = ContextKey("PamToken")
	CONTEXT_API_JOURNAL_ID_KEY    = ContextKey("ApiJournalId")
	CONTEXT_SOURCE_IP_KEY         = ContextKey("SourceIp")
	CONTEXT_SCHEDULE_ID_KEY       = ContextKey("ScheduleId")
	CONTEXT_SCHEDULE_INTERVAL_KEY = ContextKey("ScheduleInterval")
)

type HeaderName string

const (
	HEADER_CONTENT_TYPE  HeaderName = "Content-Type"
	HEADER_AUTHORIZATION HeaderName = "Authorization"
	HEADER_X_REAL_IP     HeaderName = "X-Real-Ip"
	HEADER_RANGE         HeaderName = "Range"
)

// Parses the parameter path and fetches the string value from iface
func GetFieldByPath(iface interface{}, path string) (ret string, err error) {
	r := iface
	var pathElements []string
	if !strings.Contains(path, ".") {
		pathElements = []string{path}
	} else {
		pathElements = strings.Split(path, ".")
	}
	for _, s := range pathElements {
		r, err = GetSubInterface(r, s)
	}
	if r == nil {
		ret = ""
	} else {
		if reflect.ValueOf(r).Kind() != reflect.String {
			return "", fmt.Errorf("GetFieldByPath: path '%s' must point to a string value", path)
		}
		ret = r.(string)
	}
	return
}

func ParseValuesFromMap(config map[string]interface{}, path string) (string, error) {
	re := regexp.MustCompile(`{.*?}`)
	matches := re.FindAllString(path, -1)
	for _, match := range matches {
		value, err := parseValuesFromMapSingle(config, match)
		if err != nil {
			return "", err
		}
		path = strings.Replace(path, match, value, -1)
	}
	return path, nil
}

func AutogenerateByRuleForPaymentInstrument(ctx context.Context, Config interface{}, rule string) (string, error) {
	ret := ""
	nameSourceData := make(map[string]interface{})
	configJson, err := json.Marshal(Config)
	if err != nil {
		return "", err
	}
	var config = make(map[string]interface{})
	json.Unmarshal(configJson, &config)
	nameSourceData["config"] = config

	ret, err = ParseValuesFromMap(nameSourceData, rule)

	return ret, err

}

// parses single value which key is between {}
func parseValuesFromMapSingle(config map[string]interface{}, path string) (string, error) {
	var ret interface{}
	if strings.Contains(path, "{") {
		path = strings.Replace(path, "{", "", -1)
		path = strings.Replace(path, "}", "", -1)
		path = strings.Replace(path, "result", "", -1)
		for i, s := range strings.Split(path, ".") {
			if i == 0 {
				ret = config[s]
			} else {
				res, err := GetSubInterface(ret, s)
				if err != nil {
					return "", err
				}
				ret = res
			}
		}
	} else {
		return path, nil
	}
	if ret == nil {
		return "", nil
	} else {
		return ret.(string), nil
	}
}

// Fetches the field string value by fieldName from the iface
func GetSubInterface(iface interface{}, subPath string) (interface{}, error) {
	val := reflect.ValueOf(iface)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	// if fieldName contains slice brackets, i.e. field[0] index is extracted
	fieldName := subPath
	var index string
	if strings.Contains(subPath, "[") {
		index = string(subPath[strings.Index(subPath, "[")+1 : strings.Index(subPath, "]")])
		fieldName = string(subPath[:strings.Index(subPath, "[")])
	}

	var fv reflect.Value
	if val.Kind() == reflect.Struct {
		fv = val.FieldByName(fieldName)
	} else if val.Kind() == reflect.Map {
		fv = reflect.ValueOf(iface.(map[string]interface{})[fieldName])
	} else if val.Kind() == reflect.Slice {
		fv = reflect.ValueOf(iface)
	} else {
		return "", fmt.Errorf("GetSubInterface: iface must be of type Struct or Map instead of '%s'", val.Kind())
	}
	return getReflectFieldValue(fv, fieldName, index)
}

// returns the reflect.Value
func getReflectFieldValue(fv reflect.Value, fieldName string, index string) (interface{}, error) {
	var res interface{}
	switch fv.Kind() {
	case reflect.Struct, reflect.Map:
		res = fv.Interface()
	case reflect.Bool:
		res = strconv.FormatBool(fv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		res = strconv.FormatInt(fv.Int(), 10)
	case reflect.Float32:
		res = strconv.FormatFloat(fv.Float(), 'f', -1, 32)
	case reflect.Float64:
		res = strconv.FormatFloat(fv.Float(), 'f', -1, 64)
	case reflect.String:
		res = fv.String()
	case reflect.Slice:
		if index == "" {
			return "", fmt.Errorf("getReflectFieldValue: Field '%s' recognized as slice. Element position must be provided, i.e. field[0]", fieldName)
		}
		i, err := strconv.Atoi(index)
		if err != nil {
			return "", fmt.Errorf("getReflectFieldValue: Error converting index to int from field '%s'", fieldName)
		}
		return getReflectFieldValue(reflect.ValueOf(reflect.ValueOf(fv.Interface()).Index(i).Interface()), fieldName, "")
	case reflect.Invalid:
		res = ""
	default:
		return "", fmt.Errorf("getReflectFieldValue: Field Value kind '%s' not supported. Please do implement", fv.Kind())
	}
	return res, nil
}

func StrSliceContains(ss interface{}, s string) bool {
	str := fmt.Sprintf("%v", ss)
	text := string(str[1 : len(str)-1])
	arr := strings.Split(text, " ")
	for _, v := range arr {
		if v == s {
			return true
		}
	}
	return false
}

var hashChars = []rune("abcdefghijklmnopqrstuvwxyz-1234567890")

func GenerateRandomHash(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = hashChars[rand.Intn(len(hashChars))]
	}
	return string(b)
}

type Currency string

const (
	CUR_EUR Currency = "EUR"
	CUR_GBP Currency = "GBP"
	CUR_USD Currency = "USD"
)

const (
	RATE_GBP2EUR float64 = 1.18
	RATE_USD2EUR float64 = 0.95
)

// Exchange converts value from one currency to the other
func Exchange(currencyFrom string, currencyTo string, value decimal.Decimal) (*decimal.Decimal, error) {
	if currencyFrom == currencyTo {
		return &value, nil
	}
	var res decimal.Decimal
	var curFrom = Currency(currencyFrom)
	var curTo = Currency(currencyTo)
	switch curFrom {
	case CUR_USD:
		switch curTo {
		case CUR_USD:
			res = value
		case CUR_EUR:
			res = value.Mul(decimal.NewFromFloat(RATE_USD2EUR))
		case CUR_GBP:
			res = value.Mul(decimal.NewFromFloat(RATE_USD2EUR / RATE_GBP2EUR))
		default:
			return nil, fmt.Errorf("Exchange rate TO '" + string(currencyTo) + "' is not known")
		}
	case CUR_GBP:
		switch curTo {
		case CUR_GBP:
			res = value
		case CUR_EUR:
			res = value.Mul(decimal.NewFromFloat(RATE_GBP2EUR))
		case CUR_USD:
			res = value.Mul(decimal.NewFromFloat(RATE_GBP2EUR / RATE_USD2EUR))
		default:
			return nil, fmt.Errorf("Exchange rate TO '" + string(currencyTo) + "' is not known")
		}
	case CUR_EUR:
		switch curTo {
		case CUR_EUR:
			res = value
		case CUR_GBP:
			res = value.Mul(decimal.NewFromFloat(1 / RATE_GBP2EUR))
		case CUR_USD:
			res = value.Mul(decimal.NewFromFloat(1 / RATE_USD2EUR))
		default:
			return nil, fmt.Errorf("Exchange rate TO '" + string(currencyTo) + "' is not known")
		}
	default:
		return nil, fmt.Errorf("Exchange rate FROM '" + string(currencyFrom) + "' is not known")
	}
	return &res, nil
}

// Get Monday as first day of the week
func GetFirstDayOfWeek(tm time.Time) int {
	_, _, day := tm.Date()
	weekday := tm.Weekday()
	if weekday == 0 {
		weekday = 7
	}
	firstDayOfWeek := day - int(weekday) + 1
	return firstDayOfWeek
}

// Get first day of the last week
func GetFirstDayOfLastWeek() time.Time {
	now := time.Now()
	nowLastWeek := now.Add(-7 * 24 * time.Hour)
	beginLastWeek := nowLastWeek.Add(-time.Duration(nowLastWeek.Weekday()) * 24 * time.Hour)
	year, month, day := beginLastWeek.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, now.Location())
}

// Get last day of the last week
func GetLastDayOfLastWeek() time.Time {
	now := time.Now()
	nowLastWeek := now.Add(-7 * 24 * time.Hour)
	endLastWeek := nowLastWeek.Add(time.Duration(7-nowLastWeek.Weekday()) * 24 * time.Hour)
	year, month, day := endLastWeek.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, now.Location())
}
