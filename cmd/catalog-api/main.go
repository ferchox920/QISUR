// @title QISUR Catalog API
// @version 0.1.0
// @description Catalog and identity API with WebSocket notifications.
// @host localhost:8080
// @BasePath /api/v1
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

// App encapsula las dependencias principales de la aplicacion.
type App struct {
	DB       *pgxpool.Pool
	Router   *http.Server
	WSHub    *ws.Hub
	HTTPPort string
	Logr     *slog.Logger
}

func bootstrap(ctx context.Context, cfg config.Config, logr *slog.Logger) (*App, error) {
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	wsHub := ws.NewHub(cfg.WSAllowedOrigins)
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
		verificationSender = &mailer.NoopVerificationSender{Logr: logr}
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
		TokenValidator:  httpapi.JWTValidatorAdapter{Provider: jwtProvider},
	}

	router := routerFactory.Build()
	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	return &App{
		DB:       dbPool,
		Router:   server,
		WSHub:    wsHub,
		HTTPPort: cfg.HTTPPort,
		Logr:     logr,
	}, nil
}

func main() {
	logr := logger.New()
	if err := godotenv.Load(); err != nil {
		logr.Warn("env file not loaded, using environment vars", "error", err)
	}
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := bootstrap(ctx, cfg, logr)
	if err != nil {
		logr.Error("failed to bootstrap", "error", err)
		os.Exit(1)
	}
	defer app.DB.Close()

	if app.WSHub != nil {
		go app.WSHub.Run(ctx)
	}

	go func() {
		if err := app.Router.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logr.Error("server stopped with error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()
	if err := app.Router.Shutdown(shutdownCtx); err != nil {
		logr.Error("graceful shutdown failed", "error", err)
	}
	cancel()
}
