package tasks_transport_http

import (
	"fmt"
	"net/http"

	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_http_request "github.com/LisLisich/RESTAPI/internal/core/transport/http/request"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
)

type GetTasksResponse []TaskDTOResponse

func (h *TasksHTTPHandler) GetTasks(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)
	userID, limit, offset, err := getUserIDLimitOffsetQueryParams(r)
	if err != nil {
		responseHandler.ErroResponse(
			err,
			"failed to get 'userID','limit','offset' query param",
		)
		return
	}
	taskDomain, err := h.tasksService.GetTasks(ctx, userID, limit, offset)
	if err != nil {
		responseHandler.ErroResponse(
			err,
			"failed to get task",
		)
		return
	}
	response := GetTasksResponse(tasksDTOFromDomains(taskDomain))
	responseHandler.JSONResponse(response, http.StatusOK)

}

func getUserIDLimitOffsetQueryParams(r *http.Request) (*int, *int, *int, error) {
	const (
		userIDQueryParam    = "user_id"
		limitQureParamKey   = "limit"
		offsetQueryParamKey = "offset"
	)
	userID, err := core_http_request.GetIntQueryParam(r, userIDQueryParam)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'user_id' query param: %w", err)
	}
	limit, err := core_http_request.GetIntQueryParam(r, limitQureParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'limit' query param: %w", err)
	}
	offset, err := core_http_request.GetIntQueryParam(r, offsetQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'offset' query param: %w", err)
	}
	return userID, limit, offset, nil
}
