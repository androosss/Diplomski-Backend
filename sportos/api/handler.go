// Package api contains code that is used for (generic) handling of incoming http requests to TRI PAY.
package api

import (
	L "backend/internal/logging"
	"backend/sportos"
	DA "backend/sportos/api/dto"
	BO "backend/sportos/api/handlers/backoffice"
	LO "backend/sportos/api/handlers/login"
	CL "backend/sportos/api/handlers/public"
	DR "backend/sportos/repo/dto"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

func registerHandlers(s *Server) {
	for k, ser := range s.SubServers {
		ser.MuxRouter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			L.L.Error("URL not found by CLMux", L.String("URI", r.RequestURI), L.String("client IP", r.RemoteAddr))
			ctx := r.Context()
			aPIJSONErrorResponse(ctx, w, DA.ErrorNotFound(), s.Repo)
		})
		middleware := Middleware{s}
		ser.MuxRouter.Use(middleware.panicRecoveryHandler)
		ser.MuxRouter.Use(middleware.commonMiddleware)
		// ser.MuxRouter.Use(middleware.LogRequest)
		if k != DR.SUB_LO {
			//ser.MuxRouter.Use(middleware.apiJournal) // add api journal middleware
			ser.MuxRouter.Use(middleware.jwtVerify) // add JWT support
		}

		v1Mux := ser.MuxRouter.PathPrefix(DA.API_V1).Subrouter()
		if s.CorsEnable {
			v1Mux.Use(withCORSEnabled)
		}

		registerHandlersForMux(s, v1Mux, DA.API_V1, k)
	}
}

func registerHandlersForMux(s *Server, router *mux.Router, apiVersion string, subServer DR.SubServer) {
	//Backoffice
	router.HandleFunc(string(DA.HN_API_JOURNALS), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_API_JOURNALS, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_AUDITS), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_AUDITS, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_LOGIN), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_LOGIN, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_USER), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_USER, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_SOCIAL_LOGIN), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_SOCIAL_LOGIN, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_SOCIAL_USER), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_SOCIAL_USER, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_VERIFY), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_VERIFY, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_LOGOUT), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_LOGOUT, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_SEND_RESET), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_SEND_RESET, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_RESET_PASSWORD), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_RESET_PASSWORD, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_SPORTS), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_SPORTS, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_TOURNAMENT_ROUND), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_TOURNAMENT_ROUND, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_TOURNAMENTS), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_TOURNAMENTS, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_MATCHES), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_MATCHES, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_PRACTICES), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_PRACTICES, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_PLACES), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_PLACES, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_COACHES), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_COACHES, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_TIMES), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_TIMES, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_USERPOSTS), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_USERPOSTS, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_STATS), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_STATS, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_TEAMS), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_TEAMS, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_REVIEWS), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_REVIEWS, apiVersion, subServer)
	})
	router.HandleFunc(string(DA.HN_NAME_ID), func(w http.ResponseWriter, r *http.Request) {
		HandleRequest(w, r, s, DA.HN_NAME_ID, apiVersion, subServer)
	})
	absPath, _ := filepath.Abs("../../assets/images")
	fs := http.FileServer(http.Dir(absPath))
	router.Handle(string(DA.HN_IMAGES), http.StripPrefix("/v1/assets/images", fs)).Methods(http.MethodGet)
}

// Decodes any TRI Pay request. Returns handler interface depending on URL (API version / HURL) and http method
func makeHandler(requestInfo *DA.RequestInfo, r *http.Request) (h DA.Handler, err DA.Error) {
	L.L.WithRequestID(r.Context()).Info("makeHandler")
	switch requestInfo.APIVersion {
	case DA.API_V1:
		switch requestInfo.HURL {
		case DA.HN_LOGIN:
			switch r.Method {
			case http.MethodPost:
				h = &LO.LoginPostHandler{}
			case http.MethodPut:
				h = &LO.LoginPutHandler{}
			}
		case DA.HN_SPORTS:
			switch r.Method {
			case http.MethodGet:
				h = &LO.SportsGetHandler{}
			}
		case DA.HN_TOURNAMENTS:
			switch r.Method {
			case http.MethodPost:
				h = &CL.TournamentPostHandler{}
			case http.MethodGet:
				h = &CL.TournamentsGetHandler{}
			case http.MethodPatch:
				h = &CL.TournamentPatchHandler{}
			}
		case DA.HN_TOURNAMENT_ROUND:
			switch r.Method {
			case http.MethodPost:
				h = &CL.TournamentRoundPostHandler{}
			}
		case DA.HN_MATCHES:
			switch r.Method {
			case http.MethodPost:
				h = &CL.MatchPostHandler{}
			case http.MethodGet:
				h = &CL.MatchGetHandler{}
			case http.MethodPatch:
				h = &CL.MatchPatchHandler{}
			}
		case DA.HN_PRACTICES:
			switch r.Method {
			case http.MethodPost:
				h = &CL.PracticePostHandler{}
			case http.MethodGet:
				h = &CL.PracticeGetHandler{}
			case http.MethodPatch:
				h = &CL.PracticePatchHandler{}
			}
		case DA.HN_REVIEWS:
			switch r.Method {
			case http.MethodGet:
				h = &CL.ReviewsGetHandler{}
			case http.MethodPatch:
				h = &CL.ReviewsPatchHandler{}
			}
		case DA.HN_PLACES:
			switch r.Method {
			case http.MethodGet:
				h = &CL.PlacesGetHandler{}
			}
		case DA.HN_COACHES:
			switch r.Method {
			case http.MethodGet:
				h = &CL.CoachsGetHandler{}
			}
		case DA.HN_USERPOSTS:
			switch r.Method {
			case http.MethodPost:
				h = &CL.UserpostPostHandler{}
			case http.MethodGet:
				h = &CL.UserpostGetHandler{}
			}
		case DA.HN_TIMES:
			switch r.Method {
			case http.MethodGet:
				h = &CL.TimesGetHandler{}
			}
		case DA.HN_VERIFY:
			switch r.Method {
			case http.MethodPost:
				h = &LO.UserVerifyPostHandler{}
			}
		case DA.HN_USER:
			switch r.Method {
			case http.MethodPost:
				h = &LO.UserPostHandler{}
			}
		case DA.HN_SOCIAL_USER:
			switch r.Method {
			case http.MethodPost:
				h = &LO.SocialUserPostHandler{}
			}
		case DA.HN_SOCIAL_LOGIN:
			switch r.Method {
			case http.MethodPost:
				h = &LO.SocialLoginPostHandler{}
			}
		case DA.HN_SEND_RESET:
			switch r.Method {
			case http.MethodPost:
				h = &LO.SendResetPostHandler{}
			}
		case DA.HN_RESET_PASSWORD:
			switch r.Method {
			case http.MethodPost:
				h = &LO.ResetPasswordPostHandler{}
			}
		case DA.HN_LOGOUT:
			switch r.Method {
			case http.MethodPost:
				h = &LO.LogoutPostHandler{}
			}
		case DA.HN_API_JOURNALS:
			switch r.Method {
			case http.MethodGet:
				h = &BO.ApiJournalsGetHandler{}
			}
		case DA.HN_AUDITS:
			switch r.Method {
			case http.MethodGet:
				h = &BO.AuditsGetHandler{}
			}
		case DA.HN_STATS:
			switch r.Method {
			case http.MethodGet:
				h = &CL.StatisticsGetHandler{}
			}
		case DA.HN_TEAMS:
			switch r.Method {
			case http.MethodGet:
				h = &CL.TeamsGetHandler{}
			case http.MethodPost:
				h = &CL.TeamsPostHandler{}
			case http.MethodPatch:
				h = &CL.TeamsPatchHandler{}
			}
		case DA.HN_NAME_ID:
			switch r.Method {
			case http.MethodGet:
				h = &CL.NameGetHandler{}
			}
		}
	}
	if h != nil {
		err = h.Init(r)
	} else {
		L.L.WithRequestID(r.Context()).Error("makeHandler: unknown handler path", L.String("path", requestInfo.HURL))
		err = DA.NewApiError().WithInternalError(fmt.Errorf("method or path not recognized"))
	}
	return h, err
}

// Common handler for any route
// It receives a http request and writes response to ResponseWriter
// hurl represents a sportos handler identificator
// apiVersion determines the version of the TRI Pay API that will be used
// subServer determines the type of sportos handler (public, backoffice) that will be created. It is checked if that subServer can serve the request
func HandleRequest(w http.ResponseWriter, r *http.Request, s *Server, hurl string, apiVersion string, subServer DR.SubServer) {

	L.L.WithRequestID(r.Context()).Info("handleRequest", L.Any("hurl", hurl), L.Any("mux.vars", mux.Vars(r)), L.Any("Body", r.Body))
	requestInfo := DA.NewRequestInfo(hurl, apiVersion, subServer, r)
	h, err := makeHandler(&requestInfo, r)
	ctx := r.Context()
	if err != nil {
		L.L.Error("handleRequest: error making handler", L.Error(err.GetInternalError()))
		aPIJSONErrorResponse(ctx, w, err, s.Repo)
		return
	}
	w.Header().Set(string(sportos.HEADER_CONTENT_TYPE), "application/json")
	if h.SupportedMethod() != requestInfo.Method {
		aPIJSONErrorResponse(ctx, w, DA.ErrorMethodNotAllowed(), s.Repo)
		return
	}
	if subServer != "" && !sportos.StrSliceContains(h.SupportedSubservers(), string(subServer)) {
		aPIJSONErrorResponse(ctx, w, DA.ErrorNotFound(), s.Repo)
		return
	}
	apiErr := h.Validate(requestInfo.Context, s.Repo)
	if apiErr != nil {
		aPIJSONErrorResponse(ctx, w, apiErr, s.Repo)
		return
	}
	res, apiErr := h.Process(requestInfo.Context, s.Repo)
	if apiErr != nil {
		aPIJSONErrorResponse(ctx, w, apiErr, s.Repo)
		return
	}
	responseMap, ok := res.(map[string]interface{})
	if !ok {
		aPIJSONErrorResponse(ctx, w, DA.InternalServerError(fmt.Errorf("bad response format")), s.Repo)
		return
	}
	_, exists := responseMap["headers"]
	headerMap, ok := responseMap["headers"].(map[sportos.HeaderName]string)
	if exists && !ok {
		aPIJSONErrorResponse(ctx, w, DA.InternalServerError(fmt.Errorf("bad response format")), s.Repo)
		return
	}

	for headerName, headerValue := range headerMap {
		w.Header().Set(string(headerName), headerValue)
	}

	aPIJSONResponseOK(ctx, w, responseMap["body"], s.Repo)
}
