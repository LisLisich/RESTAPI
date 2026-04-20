package statistics_transport_http

import (
	"fmt"
	"net/http"
	"time"

	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_http_request "github.com/LisLisich/RESTAPI/internal/core/transport/http/request"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
)

type GetStatisricsResponse StatisticsDTOResponse

func (h *StatisticsHTTPHandler) GetStatistics(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)
	userID, from, to, err := getUserIDFromToQueryParams(r)
	if err != nil {
		responseHandler.ErroResponse(
			err,
			"failed to get 'userID'/'from'/'to' query param",
		)
	}
	statisticsDomain, err := h.statisticsService.GetStatistics(ctx, userID, from, to)
	if err != nil {
		responseHandler.ErroResponse(
			err,
			"failed to get statistics",
		)
	}
	response := StatisticsDTOFromDomain(statisticsDomain)
	responseHandler.JSONResponse(response, http.StatusOK)
}

func getUserIDFromToQueryParams(r *http.Request) (*int, *time.Time, *time.Time, error) {
	const (
		userIDQueryParam = "user_id"
		fromQureParamKey = "from"
		toQueryParamKey  = "to"
	)
	userID, err := core_http_request.GetIntQueryParam(r, userIDQueryParam)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'user_id' query param: %w", err)
	}
	from, err := core_http_request.GetDateQueryParam(r, fromQureParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'from' query param: %w", err)
	}
	to, err := core_http_request.GetDateQueryParam(r, toQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'to' query param: %w", err)
	}
	return userID, from, to, nil
}
