package app

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"study/Account/internal/config"
	"study/Account/internal/repository"
	"study/Account/internal/server"
	"study/Account/internal/service"
	accountpb "study/contracts/account"

	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	cfg    *config.Config
	logger *zerolog.Logger

	accountRepo    *repository.Repository
	accountService *service.AccountService
	accountServer  *server.Server
	grpcServer     *grpc.Server
}

func New(cfg *config.Config, logger *zerolog.Logger) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *App) Run(ctx context.Context) error {
	accountService, err := a.getAccountService(ctx)
	if err != nil {
		return fmt.Errorf("fail to get account service: %w", err)
	}

	acSrv := server.NewServer(accountService, a.logger)
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
		serveErrCh <- a.grpcServer.Serve(lis)
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

func (a *App) getAccountService(ctx context.Context) (*service.AccountService, error) {
	if a.accountService == nil {
		repo, err := a.getRepo(ctx)
		if err != nil {
			return nil, fmt.Errorf("fail get repo: %w", err)
		}
		return service.NewAccountService(repo, a.logger), nil
	}

	return a.accountService, nil
}

func (a *App) getRepo(ctx context.Context) (*repository.Repository, error) {
	if a.accountRepo == nil {
		if err := a.runMigrations(ctx); err != nil {
			return nil, fmt.Errorf("fail to get repo: %w", err)
		}

		db, err := gorm.Open(postgres.Open(a.cfg.DB), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("gorm init fail: %w", err)
		}

		a.accountRepo = repository.NewRepository(db, a.logger)
	}
	return a.accountRepo, nil
}

func (a *App) runMigrations(ctx context.Context) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("fail select migrations dialect: %w", err)
	}

	dbGoose, err := sql.Open("postgres", a.cfg.DB)
	if err != nil {
		return fmt.Errorf("fail to create sql connection: %w", err)
	}

	if err := goose.Up(dbGoose, "Account/internal/migrations"); err != nil {
		return fmt.Errorf("fail to run migrations: %w", err)
	}

	return nil
}

func (a *App) getServer(ctx context.Context) (*server.Server, error) {
	if a.accountServer == nil {
		service, err := a.getAccountService(ctx)
		if err != nil {
			return nil, fmt.Errorf("fail to get service: %w", err)
		}

		return server.NewServer(service, a.logger), nil
	}
	return a.accountServer, nil
}

func getGRPCServer(srv *server.Server) *grpc.Server {
	grpcSrv := grpc.NewServer()
	accountpb.RegisterAccountServer(grpcSrv, srv)
	return grpcSrv
}

func (a *App) Close() error {
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}
	return nil
}
