package wallet_service

import (
	"context"
	"fmt"

	"github.com/LisLisich/fintask/internal/core/domain"
)

func (service *WalletService) GetWallet(
	ctx context.Context,
	userID int,
) (domain.Wallet, error) {
	wallet, err := service.repository.GetWallet(ctx, userID)
	if err != nil {
		return domain.Wallet{}, fmt.Errorf("get wallet: %w", err)
	}
	return wallet, nil
}
