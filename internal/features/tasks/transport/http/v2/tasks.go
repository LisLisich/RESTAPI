package tasks_transport_http_v2

import (
	"fmt"
	"net/http"
	"time"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_errors "github.com/LisLisich/fintask/internal/core/errors"
	core_logger "github.com/LisLisich/fintask/internal/core/logger"
	core_http_request "github.com/LisLisich/fintask/internal/core/transport/http/request"
	core_http_response "github.com/LisLisich/fintask/internal/core/transport/http/response"
	identity_http_middleware "github.com/LisLisich/fintask/internal/features/identity/transport/http/middleware"
)

type CreateTaskRequest struct {
	Title       string  `json:"title" validate:"required,min=3,max=100" example:"Prepare interview"`
	Description *string `json:"description" validate:"omitempty,min=1,max=1000"`
}

type TaskResponse struct {
	ID          int        `json:"id"`
	Version     int        `json:"version"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// CreateTask godoc
// @Summary Создать свою задачу
// @Description Создает задачу для текущего аутентифицированного пользователя
// @Tags tasks-v2
// @Accept json
// @Produce json
// @Param request body CreateTaskRequest true "Новая задача"
// @Success 201 {object} TaskResponse "Созданная задача"
// @Failure 400 {object} ErrorResponse "Некорректный запрос"
// @Failure 401 {object} ErrorResponse "Требуется аутентификация"
// @Security BearerAuth
// @Security SessionCookie && CSRFToken
// @Router /api/v2/tasks [post]
func (h *TasksHTTPHandler) CreateTask(rw http.ResponseWriter, r *http.Request) {
	responseHandler := newResponseHandler(rw, r)
	principal, ok := identity_http_middleware.PrincipalFromContext(r.Context())
	if !ok {
		responseHandler.ErrorResponse(
			fmt.Errorf("authenticated principal is required: %w", core_errors.ErrUnauthorized),
			"authentication required",
		)
		return
	}

	var request CreateTaskRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate task request")
		return
	}

	task, err := h.tasksService.CreateTask(
		r.Context(),
		domain.NewTaskUninitialized(request.Title, request.Description, principal.UserID),
	)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create task")
		return
	}
	responseHandler.JSONResponse(taskResponse(task), http.StatusCreated)
}

// GetTasks godoc
// @Summary Получить свои задачи
// @Description Возвращает только задачи текущего аутентифицированного пользователя
// @Tags tasks-v2
// @Produce json
// @Param limit query int false "Максимальное число задач"
// @Param offset query int false "Смещение"
// @Success 200 {array} TaskResponse "Задачи"
// @Failure 400 {object} ErrorResponse "Некорректные параметры"
// @Failure 401 {object} ErrorResponse "Требуется аутентификация"
// @Security BearerAuth
// @Security SessionCookie
// @Router /api/v2/tasks [get]
func (h *TasksHTTPHandler) GetTasks(rw http.ResponseWriter, r *http.Request) {
	responseHandler := newResponseHandler(rw, r)
	principal, ok := identity_http_middleware.PrincipalFromContext(r.Context())
	if !ok {
		responseHandler.ErrorResponse(
			fmt.Errorf("authenticated principal is required: %w", core_errors.ErrUnauthorized),
			"authentication required",
		)
		return
	}

	limit, err := core_http_request.GetIntQueryParam(r, "limit")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to parse task limit")
		return
	}
	offset, err := core_http_request.GetIntQueryParam(r, "offset")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to parse task offset")
		return
	}

	userID := principal.UserID
	tasks, err := h.tasksService.GetTasks(r.Context(), &userID, limit, offset)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get tasks")
		return
	}
	response := make([]TaskResponse, len(tasks))
	for index := range tasks {
		response[index] = taskResponse(tasks[index])
	}
	responseHandler.JSONResponse(response, http.StatusOK)
}

func newResponseHandler(
	rw http.ResponseWriter,
	r *http.Request,
) *core_http_response.HTTPResponseHandler {
	return core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(r.Context()),
		rw,
	)
}

func taskResponse(task domain.Task) TaskResponse {
	return TaskResponse{
		ID:          task.ID,
		Version:     task.Version,
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
		CreatedAt:   task.CreatedAt,
		CompletedAt: task.CompletedAt,
	}
}
