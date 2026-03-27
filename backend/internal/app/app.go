package app

import (
	"fmt"
	"net/http"
	"trackly-backend/internal/handler"
	"trackly-backend/internal/repository"
	"trackly-backend/internal/service"
	"trackly-backend/pkg/config"
	"trackly-backend/pkg/database"
	"trackly-backend/pkg/logger"
)

func Run() error {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.New(cfg.Database.URI, cfg.Database.Name)
	if err != nil {
		return err
	}

	brokerColl := db.Collection("brokers")
	brokerRepo := repository.NewBrokerRepository(brokerColl)
	brokerService := service.NewBrokerService(brokerRepo)
	brokerHandler := handler.NewBrokerHandler(brokerService)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	brokerHandler.Register(mux)

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Info("server running on " + addr)

	return http.ListenAndServe(addr, mux)
}
