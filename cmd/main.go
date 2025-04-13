package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/devWaylander/pvz_store/api"
	"github.com/devWaylander/pvz_store/internal/handler"
	auth "github.com/devWaylander/pvz_store/internal/middleware/auth"
	"github.com/devWaylander/pvz_store/internal/middleware/cors"
	"github.com/devWaylander/pvz_store/internal/middleware/logger"
	"github.com/devWaylander/pvz_store/internal/repo"
	"github.com/devWaylander/pvz_store/internal/service"
	errorgroup "github.com/devWaylander/pvz_store/pkg/error_group"
	"github.com/devWaylander/pvz_store/pkg/log"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/devWaylander/pvz_store/config"
	_ "github.com/lib/pq"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

func main() {
	// Config
	cfg, err := config.Parse()
	if err != nil {
		log.Logger.Fatal().Msg(err.Error())
	}

	// Graceful shutdown init
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		cancel()
	}()

	// DB
	db, err := sqlx.Connect("postgres", cfg.DB.DBUrl)
	if err != nil {
		log.Logger.Fatal().Msg(err.Error())
	}
	db.SetMaxOpenConns(cfg.DB.DBMaxConnections)
	db.SetMaxIdleConns(cfg.DB.DBMaxConnections)
	db.SetConnMaxLifetime(cfg.DB.DBLifeTimeConnection)
	db.SetConnMaxIdleTime(cfg.DB.DBMaxConnIdleTime)

	// Repositories
	repo := repo.New(db)
	authRepo := auth.NewRepo(db)
	// Service
	service := service.New(repo)
	// Auth
	authMiddlewares := auth.NewMiddleware(authRepo, cfg.Common.JWTSecret)

	// Handlers
	// chi router and swagger req validator
	// Инициализация chi-роутера
	r := chi.NewRouter()

	// middleware логирования
	r.Use(logger.Middleware())
	// middleware CORS
	r.Use(cors.Middleware())
	// middleware обогащения контекста AuthPrincipal
	r.Use(authMiddlewares.AuthContextEnrichingMiddleware)
	// middleware валидации запросов на основе OpenAPI
	opts := &nethttpmiddleware.Options{
		SilenceServersWarning: true,
		Options: openapi3filter.Options{
			AuthenticationFunc: authMiddlewares.Middleware(),
		},
	}
	swagger, err := api.GetSwagger()
	if err != nil {
		log.Logger.Fatal().Msgf("failed to load swagger spec: %s", err)
	}
	r.Use(nethttpmiddleware.OapiRequestValidatorWithOptions(swagger, opts))

	// Инициализация хендлеров бизнес-логики (реализация интерфейса StrictServerInterface)
	h := handler.New(authMiddlewares, service)

	// strict‑хендлер (объект, удовлетворяющий api.StrictServerInterface)
	strictHandler := api.NewStrictHandler(h, nil)

	// Регистрация эндпоинтов в роутере chi:
	h.RegisterStrictHandlers(r, strictHandler)

	// HTTP-сервер
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Common.Port),
		Handler: r,
	}

	// Graceful shutdown run
	g, gCtx := errorgroup.EGWithContext(ctx)
	g.Go(func() error {
		log.Logger.Info().Msgf("Server is up on port: %s", cfg.Common.Port)
		return httpServer.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		log.Logger.Info().Msgf("Server on port %s is shutting down", cfg.Common.Port)
		return httpServer.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		log.Logger.Info().Msgf("exit reason: %s \\n", err)
	}
}
