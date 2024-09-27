package main

import (
	"context"
	"fmt"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/config"
	"hr-tools-backend/controllers"
	"hr-tools-backend/fiberlog"
	"hr-tools-backend/initializers"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	initializers.InitAllServices(ctx)

	app := fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024, // limit of 100MB
	})
	app.Use(fiberRecover.New())

	swaggerCfg := swagger.Config{
		Path:     "/swagger",
		FilePath: "./docs/swagger.json",
	}
	app.Use(swagger.New(swaggerCfg))

	//api
	api := fiber.New()
	api.Use(fiberlog.New(*initializers.LoggerConfig))
	app.Mount("/api", api)
	api.Use(cors.New(cors.Config{
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PATCH, DELETE, PUT",
	}))
	controllers.InitRegRouters(api)
	controllers.InitOrgApiRouters(api)

	app.Hooks().OnShutdown()

	// gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	wg := sync.WaitGroup{}
	go func() {
		_ = <-c
		wg.Add(1)
		defer wg.Done()
		log.Info("Gracefully shutting down...")
		cancel()
		if err := app.Shutdown(); err != nil {
			log.WithError(err).Error("Error when try gracefully shutting down")
		}
		time.Sleep(time.Second)
		log.Info("Gracefully shutting down finished")
	}()

	// run HTTP server
	if err := app.Listen(fmt.Sprintf("%s:%d", config.Conf.App.ListenAddr, config.Conf.App.Port)); err != nil {
		log.Fatal(err)
	}

	wg.Wait()
	log.Info("HTTP server successfully stopped")
}
