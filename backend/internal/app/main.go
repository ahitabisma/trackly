package app

import (
	"fmt"
	"net/http"

	"trackly-backend/internal/handler"
	"trackly-backend/internal/repository"
	"trackly-backend/internal/routes"
	"trackly-backend/internal/service"
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

	// dependency
	companyRepo := repository.NewCompanyRepository(db)
	companyService := service.NewCompanyService(companyRepo, log)
	companyHandler := handler.NewCompanyHandler(companyService)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, log, cfg.Jwt.Secret)
	authHandler := handler.NewAuthHandler(userService)

	// router
	mux := http.NewServeMux()

	routes.Setup(
		mux,
		companyHandler,
		authHandler,
		cfg.Jwt.Secret,
		userService,
	)

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Info("server running on " + addr)

	return http.ListenAndServe(addr, mux)
}
