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
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	docs "catalog-api/docs/swagger"
	"catalog-api/internal/catalog"
	httpapi "catalog-api/internal/http"
	"catalog-api/internal/identity"
	"catalog-api/internal/storage/postgres"
	"catalog-api/internal/ws"
	"catalog-api/pkg/config"
	"catalog-api/pkg/crypto"
	"catalog-api/pkg/logger"
	"catalog-api/pkg/mailer"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// bootstrap levanta la infraestructura; el wiring de dominio sigue siendo minimo y con TODOs.
func bootstrap(ctx context.Context) (*pgxpool.Pool, *httpapi.IdentityHandler, *httpapi.RouterFactory, *ws.Hub, error) {
	cfg := config.Load()
	logr := logger.New()
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	wsHub := ws.NewHub(cfg.WSAllowedOrigins)
	go wsHub.Run(ctx)
	eventEmitter := httpapi.NewSocketEmitter(wsHub)

	identityRepo := postgres.NewIdentityRepository(dbPool)
	catalogRepo := postgres.NewCatalogRepository(dbPool)

	jwtProvider := crypto.JWTProvider{
		Secret: cfg.JWTSecret,
		Issuer: cfg.JWTIssuer,
		TTL:    cfg.JWTTTL,
	}

	var verificationSender identity.VerificationSender
	if smtpSender := mailer.NewMailVerificationSender(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.From, cfg.SMTP.SkipTLS); smtpSender != nil {
		verificationSender = smtpSender
	} else {
		logr.Warn("SMTP not configured; falling back to noop verification sender")
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
		logr.Warn("admin seed skipped", "error", err)
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
		WSHub:           wsHub,
		TokenValidator:  jwtValidatorAdapter{provider: jwtProvider},
	}

	return dbPool, identityHandler, routerFactory, wsHub, nil
}

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // fallback; main handles shutdown below.

	dbPool, _, routerFactory, _, err := bootstrap(ctx)
	if err != nil {
		log.Fatalf("failed to bootstrap: %v", err)
	}
	defer dbPool.Close()

	router := routerFactory.Build()

	addr := ":" + config.Load().HTTPPort
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server stopped with error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	cancel()
}

// noopVerificationSender es un placeholder temporal hasta integrar proveedor de email.
type noopVerificationSender struct {
	logr *slog.Logger
}

func (s *noopVerificationSender) SendVerification(ctx context.Context, email, code string) error {
	if s.logr != nil {
		s.logr.Info("verification email noop sender", "email", email)
	}
	return nil
}

// jwtValidatorAdapter conecta JWTProvider con el middleware TokenValidator.
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
