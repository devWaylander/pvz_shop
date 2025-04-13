package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/devWaylander/pvz_store/api"
	"github.com/devWaylander/pvz_store/config"
	"github.com/devWaylander/pvz_store/internal/grpc"
	"github.com/devWaylander/pvz_store/internal/handler"
	auth "github.com/devWaylander/pvz_store/internal/middleware/auth"
	"github.com/devWaylander/pvz_store/internal/middleware/cors"
	"github.com/devWaylander/pvz_store/internal/middleware/logger"
	"github.com/devWaylander/pvz_store/internal/pb/pvz_v1"
	"github.com/devWaylander/pvz_store/internal/repo"
	"github.com/devWaylander/pvz_store/internal/service"
	errorgroup "github.com/devWaylander/pvz_store/pkg/error_group"
	"github.com/devWaylander/pvz_store/pkg/log"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
	grpcPkg "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	// Запуск gRPC сервера
	grpcServer := grpcPkg.NewServer()
	go func(grpcServer *grpcPkg.Server) {
		// Настроим gRPC сервер
		lis, err := net.Listen("tcp", ":3000")
		if err != nil {
			log.Logger.Fatal().Msgf("не удалось начать прослушку порта gRPC: %v", err)
		}

		grpcService := grpc.New(repo)
		pvz_v1.RegisterPVZServiceServer(grpcServer, grpcService)

		reflection.Register(grpcServer)

		log.Logger.Info().Msgf("gRPC сервер запущен на порту 3000")
		if err := grpcServer.Serve(lis); err != nil {
			log.Logger.Fatal().Msgf("ошибка при запуске gRPC сервера: %v", err)
		}
	}(grpcServer)

	// Graceful shutdown run
	g, gCtx := errorgroup.EGWithContext(ctx)
	g.Go(func() error {
		log.Logger.Info().Msgf("HTTP сервер запущен на порту: %s", cfg.Common.Port)
		return httpServer.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		log.Logger.Info().Msgf("HTTP сервер на порту %s завершает работу", cfg.Common.Port)
		return httpServer.Shutdown(context.Background())
	})

	g.Go(func() error {
		<-gCtx.Done()
		log.Logger.Info().Msgf("gRPC сервер на порту 3000 завершает работу")
		grpcServer.GracefulStop()
		return nil
	})

	// Ожидание завершения всех горутин
	if err := g.Wait(); err != nil {
		log.Logger.Info().Msgf("Причина выхода: %s", err)
	}
}
