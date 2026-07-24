package notifications_transport_http

import (
	"fmt"
	"net/http"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
	identity_http_middleware "github.com/LisLisich/RESTAPI/internal/features/identity/transport/http/middleware"
)

type NotificationResponse struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}

func (handler *NotificationHTTPHandler) ListNotifications(
	rw http.ResponseWriter,
	request *http.Request,
) {
	responseHandler := core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(request.Context()),
		rw,
	)
	principal, ok := identity_http_middleware.PrincipalFromContext(request.Context())
	if !ok {
		responseHandler.ErrorResponse(
			fmt.Errorf("authenticated principal is required: %w", core_errors.ErrUnauthorized),
			"authentication required",
		)
		return
	}
	notifications, err := handler.reader.ListNotifications(
		request.Context(),
		principal.UserID,
	)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to list notifications")
		return
	}
	response := make([]NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		response = append(response, NotificationResponse{
			ID:        notification.ID,
			Title:     notification.Title,
			Body:      notification.Body,
			ReadAt:    notification.ReadAt,
			CreatedAt: notification.CreatedAt,
		})
	}
	responseHandler.JSONResponse(response, http.StatusOK)
}
