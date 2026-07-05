package statistics_transport_http

import (
	"fmt"
	"net/http"
	"time"

	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_http_request "github.com/LisLisich/RESTAPI/internal/core/transport/http/request"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
)

type GetStatisticsResponse StatisticsDTOResponse

// GetStatistics godoc
// @Summary      Получение статистики
// @Description  Получение статистики по задачам с опциональной фильтрацией по user_id и/или временному промежутку
// @Tags         statistics
// @Produce      json
// @Param        user_id  query     string     false "Фильтрация статистики по конкретному пользователю" Format(uuid)
// @Param        from     query     string  false "Начало промежутка рассмотрения статистики (включительно), формат: YYYY-MM-DD"
// @Param        to       query     string  false "Конец промежутся рассмотрения статистики (не включительно), формат: YYYY-MM-DD"
// @Success      200      {object}  GetStatisticsResponse "Успешное получение статистики"
// @Failure      400      {object}  core_http_response.ErrorResponse "Bad request"
// @Failure      500      {object}  core_http_response.ErrorResponse "Internal server error"
// @Router       /statistics [get]
func (h *StatisticsHTTPHandler) GetStatistics(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)
	userID, from, to, err := getUserIDFromToQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get 'userID'/'from'/'to' query param",
		)
		return
	}
	statisticsDomain, err := h.statisticsService.GetStatistics(ctx, userID, from, to)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get statistics",
		)
		return
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
