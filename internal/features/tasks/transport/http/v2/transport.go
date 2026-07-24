package tasks_transport_http_v2

import (
	"context"
	"net/http"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_http_middleware "github.com/LisLisich/RESTAPI/internal/core/transport/http/middleware"
	core_http_server "github.com/LisLisich/RESTAPI/internal/core/transport/http/server"
)

type TasksService interface {
	CreateTask(ctx context.Context, task domain.Task) (domain.Task, error)
	GetTasks(
		ctx context.Context,
		userID *int,
		limit *int,
		offset *int,
	) ([]domain.Task, error)
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
	}
}
