package dto

import (
	L "backend/internal/logging"
	"backend/sportos"
	DR "backend/sportos/repo/dto"
	"context"
	"fmt"
	"net/http"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func BoolPtr(b bool) *bool {
	return &b
}

func GetParameterFromURLQuery(httpReq *http.Request, param string) *string {
	rawQuery := strings.Replace(httpReq.URL.RawQuery, "+", "%2B", -1)
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		L.L.Error("GetParameterFromURLQuery", L.Error(err))
	}
	strArr := query[param]
	if len(strArr) > 0 {
		return &strArr[0]
	}
	return nil
}

func GetParameterFromURLPath(httpReq *http.Request, param string) string {
	paramVal := mux.Vars(httpReq)[param]
	return paramVal
}

func GetHeaderFromURL(httpReq *http.Request, headerName sportos.HeaderName) *string {
	ret, ok := httpReq.Header[string(headerName)]
	if !ok || len(ret) == 0 {
		return nil
	}
	return &ret[0]
}

func LogParameterNullSafe(methodName string, paramName string, param interface{}) {
	switch v := param.(type) {
	case (*string):
		if v != nil {
			L.L.Info(methodName, L.String(paramName, *v))
		}
	case (*int):
		if v != nil {
			L.L.Info(methodName, L.Int(paramName, *v))
		}
	case (*int64):
		if v != nil {
			L.L.Info(methodName, L.Int64(paramName, *v))
		}
	case (*time.Time):
		if v != nil {
			L.L.Info(methodName, L.Time(paramName, *v))
		}
	}
}

// Returns the UserID that was set from token by middleware
func GetUserIdFromContext(ctx context.Context) string {
	val := ctx.Value(sportos.CONTEXT_USER_ID_KEY)
	if val == nil {
		return ""
	} else {
		return val.(string)
	}
}

// Returns the User PAM token that was set from token by middleware
func GetPamTokenFromContext(ctx context.Context) string {
	val := ctx.Value(sportos.CONTEXT_PAM_TOKEN_KEY)
	if val == nil {
		return ""
	} else {
		return val.(string)
	}
}

// Returns the Api Journal Id (that was saved from incoming webhook request) that was set from token by middleware
func GetContextApiJournalIdFromContext(ctx context.Context) string {
	val := ctx.Value(sportos.CONTEXT_API_JOURNAL_ID_KEY)
	if val == nil {
		return ""
	} else {
		return val.(string)
	}
}

func ParseDate(dateStr *string, paramName string) (*time.Time, string) {
	layoutNano := time.RFC3339Nano
	layout := "2006-01-02"
	if dateStr != nil {
		date, err := time.Parse(layoutNano, *dateStr)
		if err == nil {
			return &date, ""
		} else {
			date, err = time.Parse(layout, *dateStr)
			if err == nil {
				return &date, ""
			} else {
				return nil, paramName + ": '" + *dateStr + "' isn't formatted correctly"
			}
		}
	}
	return nil, ""
}

func ParseDateOnly(dateStr *string, paramName string) (*time.Time, string) {
	layout := "2006-01-02"
	if dateStr != nil {
		date, err := time.Parse(layout, *dateStr)
		if err == nil {
			return &date, ""
		} else {
			return nil, paramName + ": '" + *dateStr + "' isn't formatted correctly"
		}
	}
	return nil, ""
}

func ParseInt(integerStr *string, paramName string) (*int64, string) {
	if integerStr != nil {
		integer, err := strconv.ParseInt(*integerStr, 10, 0)
		if err == nil {
			return &integer, ""
		} else {
			return nil, paramName + ": '" + *integerStr + "' isn't formatted correctly"
		}
	}
	return nil, ""
}

func ParseBool(booleanStr *string, paramName string) (*bool, string) {
	if booleanStr != nil {
		boolean, err := strconv.ParseBool(*booleanStr)
		if err == nil {
			return &boolean, ""
		} else {
			return nil, paramName + ": '" + *booleanStr + "' isn't formatted correctly"
		}
	}
	return nil, ""
}

func ToLowerPointer(str *string) *string {
	if str == nil {
		return nil
	}
	ret := strings.ToLower(*str)
	return &ret
}
func ToUpperPointer(str *string) *string {
	if str == nil {
		return nil
	}
	ret := strings.ToUpper(*str)
	return &ret
}

func ToUpper(str *string) string {
	if str == nil {
		return ""
	}
	return strings.ToUpper(*str)
}

func TrimEmpty(arr []string) []string {
	ret := make([]string, 0)
	for _, elem := range arr {
		if elem != "" {
			ret = append(ret, elem)
		}
	}
	return ret
}

func ParseCommaSeparatedToUpper(str *string) []string {
	if str != nil {
		ret := make([]string, 0)
		for _, elem := range strings.Split(*str, ",") {
			ret = append(ret, strings.ToUpper(strings.TrimSpace(elem)))
		}
		return ret
	}
	return nil
}

func ParseCommaSeparated(str *string) []string {
	if str != nil && *str != "" {
		ret := make([]string, 0)
		for _, elem := range strings.Split(*str, ",") {
			ret = append(ret, strings.TrimSpace(elem))
		}
		return ret
	}
	return nil
}

func GenerateRangeHeader(cnt int, pp DR.PagingSearchParams) (map[sportos.HeaderName]string, error) {
	resMap := make(map[sportos.HeaderName]string)
	zero := int64(0)
	cnt64 := int64(cnt)
	if pp.Offset == nil {
		pp.Offset = &zero
	} else {
		if *pp.Offset < 0 {
			return nil, NewApiError().WithPredefinedError(PRE_ERR_WRONG_RANGE).WithMessage("bad range parameters")
		}
	}
	if pp.Limit == nil {
		pp.Limit = &cnt64
	} else {
		if *pp.Limit < 0 {
			return nil, NewApiError().WithPredefinedError(PRE_ERR_WRONG_RANGE).WithMessage("bad range parameters")
		}
	}
	if *pp.Offset >= int64(cnt) {
		if *pp.Offset > 0 {
			resMap[sportos.HEADER_RANGE] = fmt.Sprintf("-:-/%v", cnt)
		} else {
			resMap[sportos.HEADER_RANGE] = "0:0/0"
		}
	} else {
		if *pp.Offset+*pp.Limit > int64(cnt) {
			resMap[sportos.HEADER_RANGE] = fmt.Sprintf("%v:%v/%v", *pp.Offset+1, cnt, cnt)
		} else {
			resMap[sportos.HEADER_RANGE] = fmt.Sprintf("%v:%v/%v", *pp.Offset+1, *pp.Offset+*pp.Limit, cnt)
		}
	}
	return resMap, nil
}

func ParseApiJournalSortParams(sort *string, sp *DR.ApiJournalSearchParams) {
	scs := DR.ParseSortParams(sort)
	columnsSorted := false

	for _, cs := range scs {
		clone := cs.Clone()
		switch clone.Column {
		case "apiJournalId":
			sp.ApiJournalSortParams.ApiJournalId = &clone
			columnsSorted = true
		case "createdAt":
			sp.ApiJournalSortParams.EditInfoCUSortParams.EditInfoCSortParams.CreatedAt = &clone
			columnsSorted = true
		case "updatedAt":
			sp.ApiJournalSortParams.EditInfoCUSortParams.EditInfoUSortParams.UpdatedAt = &clone
			columnsSorted = true
		}
	}

	// default sort
	if !columnsSorted {
		cs := DR.SortColumn{
			Order:     1,
			Direction: -1,
		}
		sp.ApiJournalSortParams.EditInfoCSortParams.CreatedAt = &cs
	}
}

func ParseAuditSortParams(sort *string, sp *DR.AuditSearchParams) {
	scs := DR.ParseSortParams(sort)
	columnsSorted := false
	for _, cs := range scs {
		clone := cs.Clone()
		switch clone.Column {
		case "entityId":
			sp.AuditSortParams.EntityId = &clone
			columnsSorted = true
		case "entity":
			sp.AuditSortParams.Entity = &clone
			columnsSorted = true
		}
	}
	// default sort
	if !columnsSorted {
		cs := DR.SortColumn{
			Order:     1,
			Direction: -1,
		}
		sp.AuditSortParams.EditInfoCSortParams.CreatedAt = &cs
	}
}

func BeginingOfDay() time.Time {
	t := time.Now()
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func EndOfDay() time.Time {
	t := time.Now().Add(time.Hour * 24)
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func SendMail(message string, title string, to []string) {

	// Sender data.
	from := "andrija.novakovic.1998@gmail.com"
	password := "prexhpmjccsyaulp"

	// smtp server configuration.
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	sendMessage := fmt.Sprintf("From: <%s>\r\n", from) + fmt.Sprintf("To: <%s>\r\n", to) + fmt.Sprintf("Subject: %s\r\n", title) + "\r\n" + message

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, []byte(sendMessage))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Email Sent Successfully!")
}

func AfterEqual(time1, time2 time.Time) bool {
	return time1.After(time2) || time1.Equal(time2)
}

func BeforeEqual(time1, time2 time.Time) bool {
	return time1.Before(time2) || time1.Equal(time2)
}
