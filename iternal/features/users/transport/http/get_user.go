package users_transport_http

import (
	"net/http"

	core_logger "github.com/LisLisich/RESTAPI/iternal/core/logger"
	core_http_response "github.com/LisLisich/RESTAPI/iternal/core/transport/http/response"
	core_http_utils "github.com/LisLisich/RESTAPI/iternal/core/transport/http/utils"
)

type GetUserResponse UserDTOResponse

func (h *UsersHTTPHandler) GetUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)
	userID, err := core_http_utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErroResponse(
			err,
			"failed to get userID path value",
		)
		return
	}
	userDomain, err := h.usersService.GetUser(ctx, userID)
	if err != nil {
		responseHandler.ErroResponse(
			err,
			"failed to get user",
		)
		return
	}
	response := GetUserResponse(userDTOFromDomain(userDomain))
	responseHandler.JSONResponse(response, http.StatusOK)
}
