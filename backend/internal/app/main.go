package app

import (
	"fmt"
	"net/http"
	"trackly-backend/pkg/config"
	"trackly-backend/pkg/database"
	"trackly-backend/pkg/logger"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.App.Env)

	db, err := database.NewDatabase(
		cfg.Database.Host,
		cfg.Database.Database,
		cfg.Database.User,
		cfg.Database.Password,
	)
	if err != nil {
		return err
	}
	defer db.Close()

	// router
	mux := http.NewServeMux()

	// routes.Setup(
	// 	mux,
	// 	companyHandler,
	// 	authHandler,
	// 	cfg.Jwt.Secret,
	// 	userService,
	// )

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Info("server running on " + addr)

	return http.ListenAndServe(addr, mux)
}
