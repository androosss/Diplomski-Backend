package dto

import (
	L "backend/internal/logging"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"
)

type RequestInfo struct {
	Context    context.Context
	Method     string
	HURL       string
	APIVersion string
	SubServer  DR.SubServer
}

func NewRequestInfo(httpHandler string, apiVersion string, subServer DR.SubServer, httpReq *http.Request) (handlerInfo RequestInfo) {
	L.L.WithRequestID(httpReq.Context()).Info("NewRequestInfo")
	handlerInfo.Context = httpReq.Context()
	handlerInfo.Method = httpReq.Method
	handlerInfo.HURL = httpHandler
	handlerInfo.APIVersion = apiVersion
	handlerInfo.SubServer = subServer
	return handlerInfo
}
