package notifications_transport_http

import (
	"net/http"

	core_http_middleware "github.com/LisLisich/RESTAPI/internal/core/transport/http/middleware"
	core_http_server "github.com/LisLisich/RESTAPI/internal/core/transport/http/server"
	notifications_service "github.com/LisLisich/RESTAPI/internal/features/notifications/service"
)

type NotificationHTTPHandler struct {
	reader                   notifications_service.NotificationReader
	authenticationMiddleware core_http_middleware.Middleware
}

func NewNotificationHTTPHandler(
	reader notifications_service.NotificationReader,
	authenticationMiddleware core_http_middleware.Middleware,
) *NotificationHTTPHandler {
	return &NotificationHTTPHandler{
		reader:                   reader,
		authenticationMiddleware: authenticationMiddleware,
	}
}

func (handler *NotificationHTTPHandler) Routes() []core_http_server.Route {
	middleware := []core_http_middleware.Middleware{}
	if handler.authenticationMiddleware != nil {
		middleware = append(middleware, handler.authenticationMiddleware)
	}
	return []core_http_server.Route{{
		Method:     http.MethodGet,
		Path:       "/notifications",
		Handler:    handler.ListNotifications,
		Middleware: middleware,
	}}
}
