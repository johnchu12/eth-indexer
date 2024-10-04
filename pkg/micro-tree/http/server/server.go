package server

import (
	"context"
	"fmt"
	"os"

	"hw/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

func NewHTTPServer(lc fx.Lifecycle) *fiber.App {
	logger.Infof("Executing Fiber.")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// 記錄錯誤或進行其他處理
	// ...

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Infof("Start HTTP server at %v", fmt.Sprintf(":%s", port))
			go app.Listen(fmt.Sprintf(":%s", port))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Infof("Stop HTTP server.")
			return app.Shutdown()
		},
	})
	return app
}

var Module = fx.Options(
	fx.Provide(NewHTTPServer),
)
