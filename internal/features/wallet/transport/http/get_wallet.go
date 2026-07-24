package wallet_transport_http

import (
	"fmt"
	"net/http"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
	core_logger "github.com/LisLisich/RESTAPI/internal/core/logger"
	core_http_response "github.com/LisLisich/RESTAPI/internal/core/transport/http/response"
	identity_http_middleware "github.com/LisLisich/RESTAPI/internal/features/identity/transport/http/middleware"
)

type GetWalletResponse struct {
	Version      int    `json:"version"`
	BalanceMinor int64  `json:"balance_minor" format:"int64"`
	Currency     string `json:"currency"`
}

// GetWallet godoc
// @Summary Получить свой кошелек
// @Description Возвращает баланс целым числом копеек
// @Tags wallet
// @Produce json
// @Success 200 {object} GetWalletResponse "RUB-кошелек"
// @Failure 401 {object} core_http_response.ErrorResponse "Требуется аутентификация"
// @Failure 404 {object} core_http_response.ErrorResponse "Кошелек не найден"
// @Security BearerAuth
// @Security SessionCookie
// @Router /api/v2/wallet [get]
func (handler *WalletHTTPHandler) GetWallet(rw http.ResponseWriter, r *http.Request) {
	responseHandler := core_http_response.NewHTTPResponseHandler(
		core_logger.FromContext(r.Context()),
		rw,
	)
	principal, ok := identity_http_middleware.PrincipalFromContext(r.Context())
	if !ok {
		responseHandler.ErrorResponse(
			fmt.Errorf("authenticated principal is required: %w", core_errors.ErrUnauthorized),
			"authentication required",
		)
		return
	}
	wallet, err := handler.service.GetWallet(r.Context(), principal.UserID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get wallet")
		return
	}
	responseHandler.JSONResponse(
		GetWalletResponse{
			Version:      wallet.Version,
			BalanceMinor: wallet.Balance.MinorUnits,
			Currency:     string(wallet.Balance.Currency),
		},
		http.StatusOK,
	)
}
