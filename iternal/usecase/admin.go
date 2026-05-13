package usecase

import (
	"proxy/iternal/repository"
)

type ListsUseCase struct {
	WhiteRepo repository.ListRepository
	GrayRepo  repository.ListRepository
	BlackRepo repository.ListRepository
}

func CreateListUseCase(whiteRepo, grayRepo, blackRepo repository.ListRepository) *ListsUseCase {
	return &ListsUseCase{WhiteRepo: whiteRepo, GrayRepo: grayRepo, BlackRepo: blackRepo}
}
