package tasks_transport_http_v2

import (
	"fmt"
	"net/http"

	"github.com/LisLisich/RESTAPI/internal/core/domain"
	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_http_request "github.com/LisLisich/RESTAPI/internal/core/transport/http/request"
	core_http_types "github.com/LisLisich/RESTAPI/internal/core/transport/http/types"
	identity_service "github.com/LisLisich/RESTAPI/internal/features/identity/service"
	identity_http_middleware "github.com/LisLisich/RESTAPI/internal/features/identity/transport/http/middleware"
)

// GetTask godoc
// @Summary Получить свою задачу
// @Description Возвращает задачу только текущего пользователя
// @Tags tasks-v2
// @Produce json
// @Param id path int true "ID задачи"
// @Success 200 {object} TaskResponse "Задача"
// @Failure 401 {object} ErrorResponse "Требуется аутентификация"
// @Failure 404 {object} ErrorResponse "Задача не найдена"
// @Security BearerAuth
// @Security SessionCookie
// @Router /api/v2/tasks/{id} [get]
func (h *TasksHTTPHandler) GetTask(rw http.ResponseWriter, r *http.Request) {
	responseHandler := newResponseHandler(rw, r)
	principal, taskID, ok := principalAndTaskID(responseHandler, r)
	if !ok {
		return
	}

	task, err := h.tasksService.GetOwnedTask(r.Context(), taskID, principal.UserID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get owned task")
		return
	}
	responseHandler.JSONResponse(taskResponse(task), http.StatusOK)
}

type PatchTaskRequest struct {
	Title       core_http_types.Nullable[string] `json:"title" swaggertype:"string"`
	Description core_http_types.Nullable[string] `json:"description" swaggertype:"string"`
	Completed   core_http_types.Nullable[bool]   `json:"completed" swaggertype:"boolean"`
}

func (request *PatchTaskRequest) Validate() error {
	if !request.Title.Set && !request.Description.Set && !request.Completed.Set {
		return fmt.Errorf("at least one patch field is required")
	}
	if request.Title.Set {
		if request.Title.Value == nil {
			return fmt.Errorf("title cannot be null")
		}
		titleLength := len([]rune(*request.Title.Value))
		if titleLength < 3 || titleLength > 100 {
			return fmt.Errorf("title must contain between 3 and 100 symbols")
		}
	}
	if request.Description.Set && request.Description.Value != nil {
		descriptionLength := len([]rune(*request.Description.Value))
		if descriptionLength < 1 || descriptionLength > 1000 {
			return fmt.Errorf("description must contain between 1 and 1000 symbols")
		}
	}
	if request.Completed.Set && request.Completed.Value == nil {
		return fmt.Errorf("completed cannot be null")
	}
	return nil
}

func (request PatchTaskRequest) Domain() domain.TaskPatch {
	return domain.NewTaskPatch(
		request.Title.ToDomain(),
		request.Description.ToDomain(),
		request.Completed.ToDomain(),
	)
}

// PatchTask godoc
// @Summary Изменить свою задачу
// @Description Применяет nullable patch только к задаче текущего пользователя
// @Tags tasks-v2
// @Accept json
// @Produce json
// @Param id path int true "ID задачи"
// @Param request body PatchTaskRequest true "Изменяемые поля"
// @Success 200 {object} TaskResponse "Обновленная задача"
// @Failure 400 {object} ErrorResponse "Некорректный patch"
// @Failure 401 {object} ErrorResponse "Требуется аутентификация"
// @Failure 404 {object} ErrorResponse "Задача не найдена"
// @Failure 409 {object} ErrorResponse "Конкурентное изменение"
// @Security BearerAuth
// @Security SessionCookie && CSRFToken
// @Router /api/v2/tasks/{id} [patch]
func (h *TasksHTTPHandler) PatchTask(rw http.ResponseWriter, r *http.Request) {
	responseHandler := newResponseHandler(rw, r)
	principal, taskID, ok := principalAndTaskID(responseHandler, r)
	if !ok {
		return
	}

	var request PatchTaskRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate task patch")
		return
	}
	task, err := h.tasksService.PatchOwnedTask(
		r.Context(),
		taskID,
		principal.UserID,
		request.Domain(),
	)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to patch owned task")
		return
	}
	responseHandler.JSONResponse(taskResponse(task), http.StatusOK)
}

// DeleteTask godoc
// @Summary Удалить свою задачу
// @Description Удаляет задачу только текущего пользователя
// @Tags tasks-v2
// @Param id path int true "ID задачи"
// @Success 204 "Задача удалена"
// @Failure 401 {object} ErrorResponse "Требуется аутентификация"
// @Failure 404 {object} ErrorResponse "Задача не найдена"
// @Security BearerAuth
// @Security SessionCookie && CSRFToken
// @Router /api/v2/tasks/{id} [delete]
func (h *TasksHTTPHandler) DeleteTask(rw http.ResponseWriter, r *http.Request) {
	responseHandler := newResponseHandler(rw, r)
	principal, taskID, ok := principalAndTaskID(responseHandler, r)
	if !ok {
		return
	}

	if err := h.tasksService.DeleteOwnedTask(
		r.Context(),
		taskID,
		principal.UserID,
	); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete owned task")
		return
	}
	responseHandler.NoContentResponse()
}

func principalAndTaskID(
	responseHandler interface {
		ErrorResponse(error, string)
	},
	r *http.Request,
) (identity_service.Principal, int, bool) {
	principal, ok := identity_http_middleware.PrincipalFromContext(r.Context())
	if !ok {
		responseHandler.ErrorResponse(
			fmt.Errorf("authenticated principal is required: %w", core_errors.ErrUnauthorized),
			"authentication required",
		)
		return identity_service.Principal{}, 0, false
	}
	taskID, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to parse task id")
		return identity_service.Principal{}, 0, false
	}
	return principal, taskID, true
}
