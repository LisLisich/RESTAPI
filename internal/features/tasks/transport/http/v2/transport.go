package tasks_transport_http_v2

import (
	"context"
	"net/http"

	"github.com/LisLisich/fintask/internal/core/domain"
	core_http_middleware "github.com/LisLisich/fintask/internal/core/transport/http/middleware"
	core_http_response "github.com/LisLisich/fintask/internal/core/transport/http/response"
	core_http_server "github.com/LisLisich/fintask/internal/core/transport/http/server"
)

type ErrorResponse = core_http_response.ErrorResponse

type TasksService interface {
	CreateTask(ctx context.Context, task domain.Task) (domain.Task, error)
	GetTasks(
		ctx context.Context,
		userID *int,
		limit *int,
		offset *int,
	) ([]domain.Task, error)
	GetOwnedTask(ctx context.Context, id int, userID int) (domain.Task, error)
	PatchOwnedTask(
		ctx context.Context,
		id int,
		userID int,
		patch domain.TaskPatch,
	) (domain.Task, error)
	DeleteOwnedTask(ctx context.Context, id int, userID int) error
}

type TasksHTTPHandler struct {
	tasksService      TasksService
	sessionMiddleware core_http_middleware.Middleware
}

func NewTasksHTTPHandler(
	tasksService TasksService,
	sessionMiddleware core_http_middleware.Middleware,
) *TasksHTTPHandler {
	return &TasksHTTPHandler{
		tasksService:      tasksService,
		sessionMiddleware: sessionMiddleware,
	}
}

func (h *TasksHTTPHandler) Routes() []core_http_server.Route {
	middleware := []core_http_middleware.Middleware{}
	if h.sessionMiddleware != nil {
		middleware = append(middleware, h.sessionMiddleware)
	}
	return []core_http_server.Route{
		{
			Method:     http.MethodPost,
			Path:       "/tasks",
			Handler:    h.CreateTask,
			Middleware: middleware,
		},
		{
			Method:     http.MethodGet,
			Path:       "/tasks",
			Handler:    h.GetTasks,
			Middleware: middleware,
		},
		{
			Method:     http.MethodGet,
			Path:       "/tasks/{id}",
			Handler:    h.GetTask,
			Middleware: middleware,
		},
		{
			Method:     http.MethodPatch,
			Path:       "/tasks/{id}",
			Handler:    h.PatchTask,
			Middleware: middleware,
		},
		{
			Method:     http.MethodDelete,
			Path:       "/tasks/{id}",
			Handler:    h.DeleteTask,
			Middleware: middleware,
		},
	}
}
