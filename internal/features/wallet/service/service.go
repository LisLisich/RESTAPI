package wallet_service

import (
	"context"

	"github.com/LisLisich/fintask/internal/core/domain"
)

type WalletRepository interface {
	GetWallet(ctx context.Context, userID int) (domain.Wallet, error)
}

type WalletService struct {
	repository WalletRepository
}

func NewWalletService(repository WalletRepository) *WalletService {
	return &WalletService{repository: repository}
}
