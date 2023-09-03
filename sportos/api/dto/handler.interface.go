package dto

import (
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"
)

type Handler interface {
	// Which HTTP method is supported by handler
	SupportedMethod() string
	// On which subServers is the handler exposed
	SupportedSubservers() []DR.SubServer
	// Init should read the http request data and store the data into the Handler struct
	Init(*http.Request) Error
	// Validation of Handler data that were initialized
	Validate(context.Context, *crud.Repo) Error
	// Main work is done here
	Process(context.Context, *crud.Repo) (interface{}, Error)
}
