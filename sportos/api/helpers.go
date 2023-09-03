package api

import (
	L "backend/internal/logging"
	"backend/sportos"
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"net/http"
	"net/http/httputil"
	"runtime/debug"
	"strings"
	"time"

	"github.com/rs/cors"
)

const (
	HEADER_DATE           sportos.HeaderName = "Date"
	HEADER_CONTENT_LENGTH sportos.HeaderName = "Content-Length"
)

// aPIJSONResponse writes a response with added Content-Type application/json
func aPIJSONResponse(ctx context.Context, w http.ResponseWriter, code int, payload interface{}, Repo *crud.Repo) {
	w.WriteHeader(code)
	if payload != nil {
		res, err := json.Marshal(payload)
		if err != nil {
			return
		}
		UpdateApiJournal(ctx, code, w, res, Repo)
		w.Write(res)
	}
}

func UpdateApiJournal(ctx context.Context, code int, w http.ResponseWriter, body []byte, Repo *crud.Repo) error {

	apiJournalId, ok := ctx.Value(sportos.CONTEXT_API_JOURNAL_ID_KEY).(string)
	if !ok {
		return errors.New("ApiJournalKey not set up correctly")
	}
	/*apiJournal, err := Repo.ApiJournalCrud.GetById(ctx, apiJournalId, nil)
	if err != nil {
		return nil
	}*/
	//Response status and http version
	up := DR.ApiJournalUpdateParams{}
	up.Id = apiJournalId
	response := "HTTP/1.1 "
	response = fmt.Sprintf("%s%d", response, code)
	response = fmt.Sprintf("%s %s", response, http.StatusText(code))

	//Headers
	w.Header().Set(string(HEADER_DATE), fmt.Sprint(time.Now().UTC().Format(time.RFC1123)))
	if body != nil {
		w.Header().Set(string(HEADER_CONTENT_LENGTH), fmt.Sprint(len(body)))
	} else {
		w.Header().Set(string(HEADER_CONTENT_LENGTH), fmt.Sprint(0))
	}
	for key, value := range w.Header() {
		response = fmt.Sprintf("%s\n%s: %s", response, key, value[0])
	}

	if body != nil {
		responseJson := string(body)
		response = fmt.Sprintf("%s\n%s", response, responseJson)
		up.ResponseJson = &responseJson
	} else {
		up.ResponseJson = nil
	}

	up.Response = &response
	userId := DA.GetUserIdFromContext(ctx)
	if userId != "" {
		up.UserId = &userId
	}

	_, err := Repo.ApiJournalCrud.Update(ctx, up, nil, nil)

	return err
}

// aPIJSONResponseOK writes OK response
func aPIJSONResponseOK(ctx context.Context, w http.ResponseWriter, payload interface{}, Repo *crud.Repo) {
	aPIJSONResponse(ctx, w, http.StatusOK, payload, Repo)
}

// aPIJSONErrorzResponse writes an errorz response
func aPIJSONErrorResponse(ctx context.Context, w http.ResponseWriter, err DA.Error, Repo *crud.Repo) {
	if !err.IsEmpty() {
		aPIJSONResponse(ctx, w, err.GetHTTPCode(), err, Repo)
	} else {
		aPIJSONResponse(ctx, w, err.GetHTTPCode(), nil, Repo)
	}
}

func withCORSEnabled(handler http.Handler) http.Handler {
	corsOpts := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
	})

	return corsOpts.Handler(handler)
}

// jwtVerify Middleware function
func (m *Middleware) jwtVerify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			token := strings.TrimPrefix(r.Header.Get(string(sportos.HEADER_AUTHORIZATION)), "Bearer ")
			if token == "" {
				aPIJSONErrorResponse(r.Context(), w, DA.ErrorBadRequest().WithMessage("token is missing"), m.s.Repo)
				return
			}
			sp := DR.UserSearchParams{Token: &token}
			user, err := m.s.Repo.UserCrud.Search(r.Context(), sp, nil)
			if len(user) > 0 && user[0].EmailVerified < 0 {
				aPIJSONErrorResponse(r.Context(), w, DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MAIL_NOT_VERIFIED), m.s.Repo)
			}
			if err != nil || len(user) == 0 || (user[0].TokenValidUntil != nil && user[0].TokenValidUntil.Before(time.Now())) {
				aPIJSONErrorResponse(r.Context(), w, DA.ErrorUnauthorized(), m.s.Repo)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), sportos.CONTEXT_USER_ID_KEY, user[0].Username))
		}
		next.ServeHTTP(w, r)
	})
}

type Middleware struct {
	s *Server
}

// apiJournal Middleware function
func (m *Middleware) apiJournal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiJournal := DR.ApiJournal{
			EditInfoCU: DR.EditInfoCU{},
		}
		if r.Body != nil {
			var reqBody []byte
			reqBody, _ = io.ReadAll(r.Body)
			reqJsonStr := string(reqBody)
			apiJournal.RequestBodyString = &reqJsonStr
			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}
		rb, err := httputil.DumpRequest(r, true)
		if err != nil {
			return
		}
		srq := string(rb)
		apiJournal.Request = &srq

		IPAddress := r.Header.Get(string(sportos.HEADER_X_REAL_IP))
		if IPAddress == "" {
			IPAddress = strings.Split(r.RemoteAddr, ":")[0]
		}
		apiJournal.SourceIP = &IPAddress
		r = r.WithContext(context.WithValue(r.Context(), sportos.CONTEXT_SOURCE_IP_KEY, IPAddress))
		apiJournal, err = m.s.Repo.ApiJournalCrud.Create(r.Context(), apiJournal, nil, nil)
		if err != nil {
			L.L.WithRequestID(r.Context()).Error("Error", L.Any("Error", err), L.Any("Dump request", apiJournal))
		}
		r = r.WithContext(context.WithValue(r.Context(), sportos.CONTEXT_API_JOURNAL_ID_KEY, apiJournal.ApiJournalId))
		next.ServeHTTP(w, r)
	})
}

// TODO unused, check if should be removed
// withID puts the request ID into the current context.
// func withID(ctx context.Context, id string) context.Context {
// 	return context.WithValue(ctx, L.ContextKey, id)
// }

// commonMiddleware --Set content-type
func (m *Middleware) commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(L.GetContextWithReqIDFromHeader(r.Context(), r))
		ctx := r.Context()

		defer L.L.LogDuration(ctx, time.Now(), L.String("URI", r.RequestURI))

		L.L.WithRequestID(ctx).Info(">Incoming HTTP Request", L.String("URI", r.RequestURI), L.String("Method", r.Method), L.String("Remote Host", r.Host), L.String("ClientIP", r.RemoteAddr))

		buf, _ := io.ReadAll(r.Body)
		rdr2 := io.NopCloser(bytes.NewBuffer(buf))

		if len(buf) > 0 {
			L.L.WithRequestID(ctx).Info(">HTTP request body", L.Json("Body", buf))
			r.Body = rdr2
		}

		w.Header().Add(string(sportos.HEADER_CONTENT_TYPE), "application/json")
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) panicRecoveryHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				debug.PrintStack()
				httprequest, _ := httputil.DumpRequest(r, false)
				ctx := r.Context()
				L.L.WithRequestID(ctx).DPanic("Recover from panic", L.Any("Error", err), L.String("Dump request", string(httprequest)), L.Json("Stack", debug.Stack()))
				aPIJSONErrorResponse(ctx, w, DA.InternalServerError(fmt.Errorf("%v", err)), m.s.Repo)
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}
