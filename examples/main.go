package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	goobs "github.com/todesdev/go-obs"
	"os"
)

func main() {

	app := fiber.New()

	err := goobs.Initialize(&goobs.Config{
		FiberApp:              app,
		ServiceName:           "example_app",
		ServiceVersion:        "0.0.1",
		MetricsEndpoint:       "",
		EnableFiberMiddleware: false,
		EnableMetricsHandler:  false,
		MetricsGRPC:           true,
		TracingGRPC:           true,
		GRPCEndpoint:          os.Getenv("OTLP_GRPC_ENDPOINT"),
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(app.Listen(":3000"))

}
