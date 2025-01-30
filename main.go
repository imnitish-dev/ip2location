package main

import (
	"log"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/imnitish-dev/ip2location/ip2location"
)

// Response holds the API response structure
type Response struct {
	Message     string                `json:"message,omitempty"`
	MaxMind     *ip2location.Location `json:"maxmind,omitempty"`
	IP2Location *ip2location.Location `json:"ip2location,omitempty"`
}

// App holds the application dependencies
type App struct {
	maxmindService *ip2location.Service
	ip2locService  *ip2location.Service
	fiber          *fiber.App
}

// NewApp initializes the application
func NewApp(maxmindPath, ip2locPath string) (*App, error) {
	maxmindService, err := ip2location.NewService(ip2location.MaxMindProvider, maxmindPath)
	if err != nil {
		return nil, err
	}

	ip2locService, err := ip2location.NewService(ip2location.IP2LocationProvider, ip2locPath)
	if err != nil {
		maxmindService.Close() // Clean up if second init fails
		return nil, err
	}

	app := &App{
		maxmindService: maxmindService,
		ip2locService:  ip2locService,
		fiber: fiber.New(fiber.Config{
			ErrorHandler: errorHandler,
			// Optimize for JSON responses
			JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
			// Disable startup message
			DisableStartupMessage: true,
		}),
	}

	app.setupRoutes()
	return app, nil
}

// Close releases all resources
func (a *App) Close() {
	if a.maxmindService != nil {
		a.maxmindService.Close()
	}
	if a.ip2locService != nil {
		a.ip2locService.Close()
	}
}

func (a *App) setupRoutes() {
	// Add logger middleware
	a.fiber.Use(logger.New(logger.Config{
		Format: "${time} ${status} - ${latency} ${method} ${path}\n",
	}))

	// Define routes
	a.fiber.Get("/lookup/:ip", a.handleIPLookup)
	a.fiber.Get("/health", handleHealth)
}

func (a *App) handleIPLookup(c *fiber.Ctx) error {
	ip := c.Params("ip")
	maxmindLoc, ip2locLoc := a.lookupConcurrent( ip)

	// If both lookups failed
	if maxmindLoc == nil && ip2locLoc == nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Message: "Failed to lookup IP address",
		})
	}

	return c.JSON(Response{
		MaxMind:     maxmindLoc,
		IP2Location: ip2locLoc,
	})
}

func (a *App) lookupConcurrent( ip string) (*ip2location.Location, *ip2location.Location) {
	var (
		wg          sync.WaitGroup
		maxmindLoc  *ip2location.Location
		ip2locLoc   *ip2location.Location
		maxmindErr  error
		ip2locErr   error
		maxmindOnce sync.Once
		ip2locOnce  sync.Once
	)

	// Start both lookups concurrently
	wg.Add(2)

	// MaxMind lookup
	go func() {
		defer wg.Done()
		maxmindOnce.Do(func() {
			maxmindLoc, maxmindErr = a.maxmindService.Lookup(ip)
		})
	}()

	// IP2Location lookup
	go func() {
		defer wg.Done()
		ip2locOnce.Do(func() {
			ip2locLoc, ip2locErr = a.ip2locService.Lookup(ip)
		})
	}()

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Both lookups completed
	case <-time.After(2 * time.Second):
		// Timeout occurred, return whatever we have
	}

	// Log errors if any
	if maxmindErr != nil && maxmindErr != ip2location.ErrInvalidIP {
		log.Printf("MaxMind lookup error: %v", maxmindErr)
	}
	if ip2locErr != nil && ip2locErr != ip2location.ErrInvalidIP {
		log.Printf("IP2Location lookup error: %v", ip2locErr)
	}

	return maxmindLoc, ip2locLoc
}

func handleHealth(c *fiber.Ctx) error {
	return c.JSON(Response{
		Message: "Service is healthy",
	})
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(Response{
		Message: err.Error(),
	})
}

func main() {
	app, err := NewApp("./MaxMind.mmdb", "./IP2LOCATION.BIN")
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	log.Println("Server starting on :3000")
	if err := app.fiber.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
} 