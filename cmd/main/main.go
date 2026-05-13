package main

import (
	"log"
	"proxy/config"
	"proxy/iternal/entity"
	"proxy/iternal/repository"
	"proxy/iternal/transport/http"
	"proxy/iternal/usecase"
)

func main() {

	cfg := config.CreateConfig("./config.json")

	whiteList := entity.CreateList("./data/whitelist.json")
	blackList := entity.CreateList("./data/blacklist.json")
	grayList := entity.CreateList("./data/graylist.json")

	whiteListRepo := repository.CreateRepository(&whiteList)
	blackListRepo := repository.CreateRepository(&blackList)
	grayListRepo := repository.CreateRepository(&grayList)

	listUseCase := usecase.CreateListUseCase(whiteListRepo, blackListRepo, grayListRepo)
	listHandler := http.CreateIpHandler(listUseCase)

	router := http.SetupRoute(listHandler)
	if err := router.Run(cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
