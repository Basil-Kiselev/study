package app

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"study/Auth/internal/config"
	"study/Auth/internal/repository"
	"study/Auth/internal/server"
	"study/Auth/internal/service"
	authpb "study/contracts/auth"

	_ "study/Auth/internal/migrations"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	cfg    *config.Config
	logger *zerolog.Logger

	authRepo    *repository.Repository
	authService *service.AuthService
	authServer  *server.Server
	grpcServer  *grpc.Server
}

func New(cfg *config.Config, logger *zerolog.Logger) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *App) Run(ctx context.Context) error {
	authService, err := a.getAuthService(ctx)
	if err != nil {
		return fmt.Errorf("fail to get auth service: %w", err)
	}

	acSrv := server.NewServer(authService, a.logger)
	a.grpcServer = getGRPCServer(acSrv)

	listenAddr := fmt.Sprintf("%s:%d", a.cfg.Host, a.cfg.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		a.logger.Fatal().Err(err).Msg("fail to listen")
		return err
	}

	a.logger.Info().Msg("gRPC server listen")

	serveErrCh := make(chan error, 1)
	go func() {
		// serveErrCh <- a.grpcServer.Serve(lis)
		err := a.grpcServer.Serve(lis)
		a.logger.Info().Err(err).Msg("serve stopped") // ← добавь
		serveErrCh <- err
	}()

	select {
	case <-ctx.Done():
		a.grpcServer.GracefulStop()
		return ctx.Err()
	case err := <-serveErrCh:
		if err != nil {
			a.logger.Error().Err(err).Msg("fail to serve")
		}
		return err
	}
}

func (a *App) getAuthService(ctx context.Context) (*service.AuthService, error) {
	if a.authService == nil {
		repo, err := a.getRepo(ctx)
		if err != nil {
			return nil, fmt.Errorf("fail get repo: %w", err)
		}
		return service.NewAuthService(repo, a.logger, *a.cfg), nil
	}

	return a.authService, nil
}

func (a *App) getRepo(ctx context.Context) (*repository.Repository, error) {
	if a.authRepo == nil {
		if err := a.runMigrations(ctx); err != nil {
			return nil, fmt.Errorf("fail to get repo: %w", err)
		}

		db, err := gorm.Open(postgres.Open(a.cfg.DB), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("gorm init fail: %w", err)
		}

		a.authRepo = repository.NewRepo(db, a.logger)
	}
	return a.authRepo, nil
}

func (a *App) runMigrations(ctx context.Context) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("fail select migrations dialect: %w", err)
	}

	dbGoose, err := sql.Open("postgres", a.cfg.DB)
	if err != nil {
		return fmt.Errorf("fail to create sql connection: %w", err)
	}

	if err := goose.Up(dbGoose, "internal/migrations"); err != nil {
		return fmt.Errorf("fail to run migrations: %w", err)
	}

	return nil
}

func (a *App) getServer(ctx context.Context) (*server.Server, error) {
	if a.authServer == nil {
		service, err := a.getAuthService(ctx)
		if err != nil {
			return nil, fmt.Errorf("fail to get service: %w", err)
		}

		return server.NewServer(service, a.logger), nil
	}
	return a.authServer, nil
}

func getGRPCServer(srv *server.Server) *grpc.Server {
	grpcSrv := grpc.NewServer()
	authpb.RegisterAuthServer(grpcSrv, srv)
	reflection.Register(grpcSrv)
	return grpcSrv
}

func (a *App) Close() error {
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}
	return nil
}
