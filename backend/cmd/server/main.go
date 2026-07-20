package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/auth"
	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/config"
	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/db"
	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/game"
	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/gemini"
	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/mail"
	"github.com/KaichoHarry/AI-shiritori-app/backend/internal/settings"
)

func main() {
	// ローカルで `go run` する場合に .env を読み込む(docker-composeはenv_fileで注入するため任意)。
	_ = godotenv.Load("../.env", ".env")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	mailer := mail.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom, cfg.AdminEmail)

	authRepo := auth.NewRepository(pool)
	tokens := auth.NewTokenIssuer(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	authService := auth.NewService(authRepo, tokens, mailer, cfg.PasswordResetTTL)
	authHandler := auth.NewHandler(authService, tokens)

	settingsRepo := settings.NewRepository(pool)
	settingsHandler := settings.NewHandler(settingsRepo)

	geminiClient, err := gemini.New(ctx, cfg.GeminiAPIKey, cfg.GeminiModel, cfg.GeminiTimeout)
	if err != nil {
		log.Fatalf("create gemini client: %v", err)
	}
	gameRepo := game.NewRepository(pool)
	gameService := game.NewService(gameRepo, settingsRepo, geminiClient)
	gameHandler := game.NewHandler(gameService)

	if cfg.IMAPHost != "" {
		poller := mail.NewIMAPPoller(cfg.IMAPHost, cfg.IMAPPort, cfg.IMAPUsername, cfg.IMAPPassword, cfg.IMAPPollInterval, authService)
		go poller.Run(ctx)
	} else {
		log.Println("IMAP_HOST is not set: password reset approval polling is disabled")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api", func(api chi.Router) {
		api.Mount("/auth", authHandler.Routes())
		api.Mount("/settings", settingsHandler.Routes(authHandler.RequireAuth))
		api.Mount("/games", gameHandler.Routes(authHandler.RequireAuth))
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
