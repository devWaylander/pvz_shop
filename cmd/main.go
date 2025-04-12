package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	errorgroup "github.com/devWaylander/coins_store/pkg/error_group"
	"github.com/devWaylander/pvz_store/api"
	"github.com/devWaylander/pvz_store/internal/handler"
	"github.com/devWaylander/pvz_store/internal/repo"
	"github.com/devWaylander/pvz_store/internal/service"
	"github.com/devWaylander/pvz_store/pkg/log"
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
	// Auth Middleware
	// Service
	service := service.New(repo)

	// Handlers
	// chi router and swagger req validator
	// Инициализация chi-роутера и middleware валидации запросов на основе OpenAPI
	r := chi.NewRouter()
	swagger, err := api.GetSwagger()
	if err != nil {
		log.Logger.Fatal().Msgf("failed to load swagger spec: %s", err)
	}
	r.Use(nethttpmiddleware.OapiRequestValidator(swagger))

	// Инициализация хендлеров бизнес-логики (реализация интерфейса StrictServerInterface)
	h := handler.New(service)

	// Создаем strict‑хендлер (объект, удовлетворяющий api.StrictServerInterface)
	strictHandler := api.NewStrictHandler(h, nil)

	// Регистрируем все эндпоинты в роутере chi:
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
