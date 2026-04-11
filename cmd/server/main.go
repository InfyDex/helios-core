package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/infydex/helios-core/internal/auth"
	"github.com/infydex/helios-core/internal/config"
	"github.com/infydex/helios-core/internal/handler"
	appmw "github.com/infydex/helios-core/internal/middleware"
	"github.com/infydex/helios-core/internal/user"
	"github.com/infydex/helios-core/pkg/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	queries := db.New(pool)
	userStore := user.NewStore(queries)
	authSvc := auth.NewService(cfg.GoogleClientID, cfg.JWTSecret, cfg.JWTExpiry, userStore)

	app := fiber.New(fiber.Config{
		AppName:      "Helios Core",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
			}
			return c.Status(code).JSON(fiber.Map{"error": "internal server error"})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(appmw.RequestID())

	core := app.Group(handler.CoreAPIPrefix)
	core.Get("/health", handler.Health)
	handler.NewAuth(core, authSvc)

	go func() {
		addr := ":" + cfg.Port
		log.Printf("listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
