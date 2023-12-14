package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	goobs "github.com/todesdev/go-obs"
)

func main() {

	app := fiber.New()

	err := goobs.Initialize(&goobs.Config{
		FiberApp:              app,
		ServiceName:           "example_app",
		ServiceVersion:        "0.0.1",
		MetricsEndpoint:       "/metrics",
		EnableFiberMiddleware: true,
		EnableMetricsHandler:  true,
		MetricsPrometheus:     true,
		TracingGRPC:           false,
		//GRPCEndpoint:          os.Getenv("OTLP_GRPC_ENDPOINT"),
		GRPCEndpoint: "",
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(app.Listen(":3000"))

}
