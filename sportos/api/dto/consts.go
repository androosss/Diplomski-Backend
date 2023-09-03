// Package dto (api) contains data transfer objects (DTOs) used in api package
package dto

import (
	"backend/sportos"
	DR "backend/sportos/repo/dto"
	"time"
)

const (
	API_V1 string = "/v1"
	// API_V2  string    = "/v2"
)

const (
	//Login
	HN_LOGIN          string = "/login"
	HN_USER           string = "/user"
	HN_SOCIAL_USER    string = "/social-user"
	HN_SOCIAL_LOGIN   string = "/social-login"
	HN_VERIFY         string = "/verify"
	HN_LOGOUT         string = "/logout"
	HN_SEND_RESET     string = "/send-reset"
	HN_RESET_PASSWORD string = "/reset-password"
	HN_SPORTS         string = "/sports"
	//public
	HN_TOURNAMENT_ROUND string = "/tournament-round"
	HN_TOURNAMENTS      string = "/tournaments"
	HN_MATCHES          string = "/matches"
	HN_PRACTICES        string = "/practices"
	HN_PLACES           string = "/places"
	HN_COACHES          string = "/coaches"
	HN_TIMES            string = "/times"
	HN_IMAGES           string = "/assets/images/{id}"
	HN_USERPOSTS        string = "/userposts"
	HN_STATS            string = "/statistics"
	HN_TEAMS            string = "/teams"
	HN_REVIEWS          string = "/reviews"
	HN_NAME_ID          string = "/name/{id}"
	//Backoffice
	HN_API_JOURNALS string = "/api-journals"
	HN_AUDITS       string = "/audits"
)

const (
	HEADER_USER_ID  sportos.HeaderName = "User-Id"
	HEADER_PARTNERS sportos.HeaderName = "Partners"
)

// [swagger]

// ApiJournal
//
// API Journal records all the calls to TRI Pay API
// swagger:model ApiJournal
type ApiJournal struct {
	// API Journal ID
	ApiJournalId string `json:"apiJournalId,omitempty"`
	// Caller IP address
	SourceIP string `json:"sourceIP,omitempty"`
	// Full HTTP request that was received
	Request string `json:"request,omitempty"`
	// Full HTTP response that was replied by TRI Pay
	Response string `json:"response,omitempty"`
	// date and time the journal was created
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

func (aj *ApiJournal) InitWithDatabaseStruct(do *DR.ApiJournal) {
	aj.ApiJournalId = do.ApiJournalId
	aj.SourceIP = *do.SourceIP
	if do.Request != nil {
		aj.Request = *do.Request
	}
	if do.Response != nil {
		aj.Response = *do.Response
	}
	aj.CreatedAt = do.CreatedAt
}
