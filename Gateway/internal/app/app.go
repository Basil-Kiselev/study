package app

import (
	"context"
	"fmt"
	"net"
	"study/Gateway/internal/account"
	"study/Gateway/internal/auth"
	"study/Gateway/internal/config"
	"study/Gateway/internal/interceptor"
	"study/Gateway/internal/server"
	"study/Gateway/internal/service"
	accpb "study/contracts/account"
	authpb "study/contracts/auth"
	gatewaypb "study/contracts/gateway"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	cfg            *config.Config
	logger         *zerolog.Logger
	service        *service.GatewayService
	authService    *auth.Service
	accountService *account.Service
	server         *server.Server
	grpcServer     *grpc.Server
}

func New(logger *zerolog.Logger, cfg *config.Config) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *App) Run(ctx context.Context) error {
	gatewayServer, err := a.getGatewayServer(ctx)
	if err != nil {
		return fmt.Errorf("fail to get gateway server: %w", err)
	}
	a.grpcServer = a.getGrpcServer(gatewayServer)

	listenAddr := fmt.Sprintf("%s:%s", a.cfg.Host, a.cfg.Port)
	listen, err := net.Listen("tcp", listenAddr)
	if err != nil {
		a.logger.Fatal().Err(err).Msg("fail to listen")
		return err
	}

	a.logger.Info().Msg("server listen")

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- a.grpcServer.Serve(listen)
	}()

	select {
	case <-ctx.Done():
		a.grpcServer.GracefulStop()
		return ctx.Err()
	case err := <-serverErrCh:
		if err != nil {
			a.logger.Error().Err(err).Msg("fail to serve")
		}

		return err
	}
}

func (a *App) getGatewayService(ctx context.Context, logger *zerolog.Logger) (*service.GatewayService, error) {
	if a.service != nil {
		return a.service, nil
	}

	authSrv, err := a.getAuthService(ctx)
	if err != nil {
		return nil, err
	}

	accSrv, err := a.getAccountService(ctx)
	if err != nil {
		return nil, err
	}

	a.service = service.New(logger, authSrv, accSrv)
	return a.service, nil
}

func (a *App) getAuthService(ctx context.Context) (*auth.Service, error) {
	if a.authService != nil {
		return a.authService, nil
	}

	conn, err := grpc.NewClient(
		a.cfg.AuthGrpcHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, fmt.Errorf("fail to create auth service: %w", err)
	}

	client := authpb.NewAuthClient(conn)
	a.authService = auth.NewService(client)

	return a.authService, nil
}

func (a *App) getAccountService(ctx context.Context) (*account.Service, error) {
	if a.accountService != nil {
		return a.accountService, nil
	}

	conn, err := grpc.NewClient(
		a.cfg.AccountGrpcHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, fmt.Errorf("fail to create acc service: %w", err)
	}

	client := accpb.NewAccountClient(conn)
	a.accountService = account.NewService(client)

	return a.accountService, nil
}

func (a *App) getGatewayServer(ctx context.Context) (*server.Server, error) {
	if a.server != nil {
		return a.server, nil
	}

	svc, err := a.getGatewayService(ctx, a.logger)
	if err != nil {
		return nil, err
	}

	a.server = server.New(svc, a.logger)
	return a.server, nil
}

func (a *App) getGrpcServer(server *server.Server) *grpc.Server {
	jwtInterceptor := interceptor.NewJWTInterceptor(a.cfg.JWTSecret)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(jwtInterceptor.UnaryInterceptor()),
	)

	gatewaypb.RegisterGatewayServer(grpcServer, server)
	return grpcServer
}

func (a *App) Close() error {
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}

	return nil
}
