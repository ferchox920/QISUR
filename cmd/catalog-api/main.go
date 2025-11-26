// @title QISUR Catalog API
// @version 0.1.0
// @description Catalog and identity API with WebSocket notifications.
// @BasePath /api/v1
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"log"
	"net/http"

	docs "catalog-api/docs/swagger"
	"catalog-api/internal/catalog"
	httpapi "catalog-api/internal/http"
	"catalog-api/internal/identity"
	"catalog-api/internal/storage/postgres"
	"catalog-api/pkg/config"
	"catalog-api/pkg/crypto"
	"catalog-api/pkg/logger"
	"catalog-api/pkg/mailer"

	socketio "github.com/googollee/go-socket.io"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// bootstrap sets up infrastructure; domain wiring remains minimal and TODO-driven.
func bootstrap(ctx context.Context) (*pgxpool.Pool, *httpapi.IdentityHandler, *httpapi.RouterFactory, *socketio.Server, error) {
	cfg := config.Load()
	logr := logger.New()
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	wsServer := socketio.NewServer(nil)
	go wsServer.Serve()
	eventEmitter := httpapi.NewSocketEmitter(wsServer)

	identityRepo := postgres.NewIdentityRepository(dbPool)
	catalogRepo := postgres.NewCatalogRepository(dbPool)

	jwtProvider := crypto.JWTProvider{
		Secret: cfg.JWTSecret,
		Issuer: cfg.JWTIssuer,
		TTL:    cfg.JWTTTL,
	}

	var verificationSender identity.VerificationSender
	if smtpSender := mailer.NewGomailVerificationSender(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.From, cfg.SMTP.SkipTLS); smtpSender != nil {
		verificationSender = smtpSender
	} else {
		logr.Printf("warning: SMTP not configured, falling back to noop verification sender")
		verificationSender = &noopVerificationSender{logr: logr}
	}

	codeGenerator := crypto.RandomDigitsGenerator{Length: 6}

	idService := identity.NewService(identity.ServiceDeps{
		UserRepo:                 identityRepo,
		RoleRepo:                 identityRepo,
		PasswordHasher:           crypto.BcryptHasher{},
		VerificationSender:       verificationSender,
		VerificationCodeProvider: codeGenerator,
		TokenProvider:            jwtProvider,
	})

	if err := idService.SeedAdmin(ctx, identity.AdminSeedInput{
		Email:    cfg.AdminSeed.Email,
		Password: cfg.AdminSeed.Password,
		FullName: cfg.AdminSeed.FullName,
	}); err != nil {
		logr.Printf("admin seed skipped: %v", err)
	}

	catService := catalog.NewService(catalog.ServiceDeps{
		CategoryRepo: catalogRepo,
		ProductRepo:  catalogRepo,
	})
	catalogHandler := httpapi.NewCatalogHandler(catService, eventEmitter)

	identityHandler := httpapi.NewIdentityHandler(idService)

	routerFactory := &httpapi.RouterFactory{
		CatalogHandler:  catalogHandler,
		IdentityHandler: identityHandler,
		WSServer:        wsServer,
		TokenValidator:  jwtValidatorAdapter{provider: jwtProvider},
	}

	return dbPool, identityHandler, routerFactory, wsServer, nil
}

func main() {
	_ = godotenv.Load()

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
