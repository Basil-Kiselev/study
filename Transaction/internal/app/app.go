package app

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"study/Transaction/internal/account"
	"study/Transaction/internal/config"
	"study/Transaction/internal/repository"
	"study/Transaction/internal/server"
	"study/Transaction/internal/service"

	accountpb "study/contracts/account"
	transactionpb "study/contracts/transaction"

	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	logger             *zerolog.Logger
	cfg                *config.Config
	transactionRepo    *repository.Repository
	transactionService *service.TransactionService
	transactionServer  *server.Server
	grpcServer         *grpc.Server
	accountService     *account.Service
}

func New(logger *zerolog.Logger, cfg *config.Config) *App {
	return &App{
		logger: logger,
		cfg:    cfg,
	}
}

func (a *App) Run(ctx context.Context) error {
	tServer, err := a.getServer(ctx)
	if err != nil {
		return fmt.Errorf("fail to get transaction server: %w", err)
	}

	a.grpcServer = getGRPCServer(tServer)

	listenAddr := fmt.Sprintf("%s:%d", a.cfg.Host, a.cfg.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("fail to listen: %w", err)
	}

	a.logger.Info().Msg("gRPC server listening")

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- a.grpcServer.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		a.grpcServer.GracefulStop()
		return ctx.Err()
	case err := <-serverErrCh:
		if err != nil {
			a.logger.Error().Msg("fail to serve")
		}
	}

	return err
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

func (a *App) getRepo(ctx context.Context) (*repository.Repository, error) {
	if a.transactionRepo == nil {
		if err := a.runMigrations(ctx); err != nil {
			return nil, fmt.Errorf("fail to get repo: %w", err)
		}

		db, err := gorm.Open(postgres.Open(a.cfg.DB), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("gorm init fail: %w", err)
		}

		a.transactionRepo = repository.NewRepo(db, a.logger)
	}

	return a.transactionRepo, nil
}

func (a *App) getAccService(ctx context.Context) (*account.Service, error) {
	if a.accountService == nil {
		conn, err := grpc.Dial(a.cfg.AccountGrpcHost, grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("fail to connect account service: %w", err)
		}

		client := accountpb.NewAccountClient(conn)
		a.accountService = account.New(client)
	}

	return a.accountService, nil
}

func (a *App) getTransactionService(ctx context.Context) (*service.TransactionService, error) {
	if a.transactionService == nil {
		repo, err := a.getRepo(ctx)
		if err != nil {
			return nil, fmt.Errorf("fail to get repo for transaction service: %w", err)
		}

		aService, err := a.getAccService(ctx)
		if err != nil {
			return nil, fmt.Errorf("fail to get account service")
		}

		service := service.New(repo, a.logger, aService)
		a.transactionService = service
	}

	return a.transactionService, nil
}

func (a *App) getServer(ctx context.Context) (*server.Server, error) {
	if a.transactionServer == nil {
		tService, err := a.getTransactionService(ctx)
		if err != nil {
			return nil, fmt.Errorf("fail to get service for server: %w", err)
		}

		server := server.NewServer(tService, a.logger)
		a.transactionServer = server
	}

	return a.transactionServer, nil
}

func getGRPCServer(srv *server.Server) *grpc.Server {
	grpcSrv := grpc.NewServer()
	transactionpb.RegisterTransactionServiceServer(grpcSrv, srv)

	return grpcSrv
}

func (a *App) Close() error {
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}
	return nil
}
