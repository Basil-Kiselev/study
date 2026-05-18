package interceptor

import (
	"context"
	"fmt"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	AuthorizationHeader = "authorization"
	BearerPrefix        = "Bearer "
)

type contextKey string

const UserIDKey contextKey = "user_id"

type JWTClaims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTInterceptor struct {
	jwtSecret []byte
}

var publicMethods = map[string]bool{
	"/gateway.Gateway/Register": true,
	"/gateway.Gateway/Login":    true,
	"/gateway.Gateway/Refresh":  true,
}

func NewJWTInterceptor(jwtSecret string) *JWTInterceptor {
	return &JWTInterceptor{
		jwtSecret: []byte(jwtSecret),
	}
}

func (i *JWTInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		userID, err := i.extractUserIDFromJWT(ctx)
		if err != nil {
			return nil, err
		}
		ctxWithUserID := context.WithValue(ctx, UserIDKey, userID)

		return handler(ctxWithUserID, req)
	}
}

func (i *JWTInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if publicMethods[info.FullMethod] {
			return handler(srv, stream)
		}

		userID, err := i.extractUserIDFromJWT(stream.Context())
		if err != nil {
			return err
		}

		ctxWithUserID := context.WithValue(stream.Context(), UserIDKey, userID)

		wrapperStream := &wrappedServerStream{
			ServerStream: stream,
			ctx:          ctxWithUserID,
		}

		return handler(srv, wrapperStream)
	}
}

func (i *JWTInterceptor) extractUserIDFromJWT(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	authHeaders := md.Get(AuthorizationHeader)
	if len(authHeaders) == 0 {
		return 0, status.Errorf(codes.Unauthenticated, "authorization header is not provided")
	}

	token := authHeaders[0]
	if !strings.HasPrefix(token, BearerPrefix) {
		return 0, status.Errorf(codes.Unauthenticated, "access token is empty")
	}

	accessToken := strings.TrimPrefix(token, BearerPrefix)
	if accessToken == "" {
		return 0, status.Errorf(codes.Unauthenticated, "access token is empty")
	}

	userID, err := i.parseAndValidateJWT(accessToken)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return userID, nil
}

func (i *JWTInterceptor) parseAndValidateJWT(tokenString string) (uint64, error) {
	claimsNew := JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claimsNew, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return i.jwtSecret, nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return 0, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return 0, fmt.Errorf("failed to extract claims")
	}

	if claims.UserID == 0 {
		return 0, fmt.Errorf("user not found in token")
	}

	return claims.UserID, nil
}

func (i *JWTInterceptor) ValidateTokenWithoutAuthService(token string) (uint64, error) {
	return i.parseAndValidateJWT(token)
}

func GetUserIDFromContext(ctx context.Context) (uint64, error) {
	userID, ok := ctx.Value(UserIDKey).(uint64)
	if !ok {
		return 0, fmt.Errorf("fail to get user in context")
	}
	return userID, nil
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
