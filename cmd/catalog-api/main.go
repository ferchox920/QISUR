package main

import (
	"context"
	"log"
	"net/http"
	"os"

	httpapi "catalog-api/internal/http"
	"catalog-api/internal/identity"
	"catalog-api/internal/storage/postgres"
	"catalog-api/pkg/config"
	"catalog-api/pkg/crypto"
	"catalog-api/pkg/logger"

	socketio "github.com/googollee/go-socket.io"
	"github.com/jackc/pgx/v5/pgxpool"
)

// bootstrap sets up infrastructure; domain wiring remains minimal and TODO-driven.
func bootstrap(ctx context.Context) (*pgxpool.Pool, *httpapi.IdentityHandler, *httpapi.RouterFactory, *socketio.Server, error) {
	cfg := config.Load()
	logr := logger.New()

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	wsServer := socketio.NewServer(nil)
	go wsServer.Serve()

	identityRepo := postgres.NewIdentityRepository(dbPool)

	jwtProvider := crypto.JWTProvider{
		Secret: cfg.JWTSecret,
		Issuer: cfg.JWTIssuer,
		TTL:    cfg.JWTTTL,
	}

	idService := identity.NewService(identity.ServiceDeps{
		UserRepo:                 identityRepo,
		RoleRepo:                 identityRepo,
		PasswordHasher:           crypto.BcryptHasher{},
		VerificationSender:       &noopVerificationSender{logr: logr},
		VerificationCodeProvider: noopVerificationCodeGenerator{},
		TokenProvider:            jwtProvider,
	})

	if err := idService.SeedAdmin(ctx, identity.AdminSeedInput{
		Email:    cfg.AdminSeed.Email,
		Password: cfg.AdminSeed.Password,
		FullName: cfg.AdminSeed.FullName,
	}); err != nil {
		logr.Printf("admin seed skipped: %v", err)
	}

	identityHandler := httpapi.NewIdentityHandler(idService)

	routerFactory := &httpapi.RouterFactory{
		IdentityHandler: identityHandler,
		WSServer:        wsServer,
		TokenValidator:  jwtValidatorAdapter{provider: jwtProvider},
	}

	return dbPool, identityHandler, routerFactory, wsServer, nil
}

func main() {
	ctx := context.Background()

	dbPool, _, routerFactory, wsServer, err := bootstrap(ctx)
	if err != nil {
		log.Fatalf("failed to bootstrap: %v", err)
	}
	defer dbPool.Close()
	defer wsServer.Close()

	router := routerFactory.Build()

	addr := ":" + config.Load().HTTPPort
	// TODO: plug in graceful shutdown with context cancellation.
	if err := router.Run(addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server stopped with error: %v", err)
	}
}

// noopVerificationSender is a temporary placeholder until email provider is integrated.
type noopVerificationSender struct {
	logr *log.Logger
}

func (s *noopVerificationSender) SendVerification(ctx context.Context, email, code string) error {
	if s.logr != nil {
		s.logr.Printf("verification email to %s with code %s (noop)", email, code)
	}
	return nil
}

// noopVerificationCodeGenerator issues static codes as a placeholder.
type noopVerificationCodeGenerator struct{}

func (noopVerificationCodeGenerator) Generate(ctx context.Context, userID string) (string, error) {
	return os.Getenv("DEFAULT_VERIFICATION_CODE"), nil
}

// jwtValidatorAdapter bridges JWTProvider to HTTP middleware TokenValidator.
type jwtValidatorAdapter struct {
	provider crypto.JWTProvider
}

func (j jwtValidatorAdapter) Validate(token string) (httpapi.AuthContext, error) {
	claims, err := j.provider.Validate(token)
	if err != nil {
		return httpapi.AuthContext{}, err
	}
	return httpapi.AuthContext{
		UserID: claims.Subject,
		Role:   claims.Role,
	}, nil
}
