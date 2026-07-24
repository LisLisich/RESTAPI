package health_transport_http

import (
	"net/http"

	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
	core_http_server "github.com/LisLisich/RESTAPI/internal/core/transport/http/server"
)

type HealthHTTPHandler struct{}

func NewHealthHTTPHandler() *HealthHTTPHandler {
	return &HealthHTTPHandler{}
}

func (handler *HealthHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{{
		Method:  http.MethodGet,
		Path:    "/healthz",
		Handler: handler.Health,
	}}
}

func (handler *HealthHTTPHandler) Health(
	rw http.ResponseWriter,
	request *http.Request,
) {
	responseHandler := core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(request.Context()),
		rw,
	)
	responseHandler.JSONResponse(map[string]string{"status": "ok"}, http.StatusOK)
}
