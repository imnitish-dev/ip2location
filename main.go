package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/imnitish-dev/ip2location/ip2location"
	pb "github.com/imnitish-dev/ip2location/proto"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

// Config holds the application configuration
type Config struct {
	Port            string
	Host            string
	MaxMindDBPath   string
	IP2LocationPath string
	GRPCPort        string
}

// loadConfig loads the configuration from environment variables
func loadConfig() (*Config, error) {
	env := getEnv("ENV", "development")
	envFile := fmt.Sprintf(".env.%s", env)

	// Try environment-specific file first
	if err := godotenv.Load(envFile); err != nil {
		// Fall back to default .env
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: no .env file found, using environment variables")
		}
	}

	config := &Config{
		Port:            getEnv("PORT", "3000"),
		Host:            getEnv("HOST", "0.0.0.0"),
		MaxMindDBPath:   getEnv("MAXMIND_DB_PATH", "./MaxMind.mmdb"),
		IP2LocationPath: getEnv("IP2LOCATION_DB_PATH", "./IP2LOCATION.BIN"),
		GRPCPort:        getEnv("GRPC_PORT", "50051"),
	}

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

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
		return nil, fmt.Errorf("failed to initialize MaxMind service: %w", err)
	}

	ip2locService, err := ip2location.NewService(ip2location.IP2LocationProvider, ip2locPath)
	if err != nil {
		maxmindService.Close() // Clean up if second init fails
		return nil, fmt.Errorf("failed to initialize IP2Location service: %w", err)
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

func sanitizeIP(rawIp string) (string, error) {
	ip, err := url.QueryUnescape(rawIp)
	if err != nil {
		return "", err
	}

	// Retrieve IP parameter
	ip = strings.TrimSpace(ip)
	// Remove spaces
	ip = strings.TrimSpace(ip)
	
	// URL encode the IP
	ip = url.QueryEscape(ip)
	
	// Remove dashes and replace with dots
	ip = strings.ReplaceAll(ip, "-", ".")
	
	// Validate IP address format
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid IP address format")
	}
	
	return ip, nil
}

func (a *App) handleIPLookup(c *fiber.Ctx) error {
	ip, err := sanitizeIP(c.Params("ip"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Message: err.Error(),
		})
	}
	maxmindLoc, ip2locLoc := a.lookupConcurrent(ip)

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

func (a *App) lookupConcurrent(ip string) (*ip2location.Location, *ip2location.Location) {
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

// GRPCServer implements the gRPC service
type GRPCServer struct {
	pb.UnimplementedIP2LocationServiceServer
	app *App
}

// LookupIP implements the gRPC lookup method
func (s *GRPCServer) LookupIP(ctx context.Context, req *pb.LookupRequest) (*pb.LookupResponse, error) {
	maxmindLoc, ip2locLoc := s.app.lookupConcurrent(req.Ip)

	response := &pb.LookupResponse{}

	if maxmindLoc == nil && ip2locLoc == nil {
		response.Message = "Failed to lookup IP address"
		return response, nil
	}

	if maxmindLoc != nil {
		response.Maxmind = &pb.Location{
			Country:     maxmindLoc.Country,
			City:       maxmindLoc.City,
			Region:     maxmindLoc.Region,
			Latitude:   maxmindLoc.Latitude,
			Longitude:  maxmindLoc.Longitude,
			CountryCode: maxmindLoc.CountryCode,
		}
	}

	if ip2locLoc != nil {
		response.Ip2Location = &pb.Location{
			Country:     ip2locLoc.Country,
			City:       ip2locLoc.City,
			Region:     ip2locLoc.Region,
			Latitude:   ip2locLoc.Latitude,
			Longitude:  ip2locLoc.Longitude,
			CountryCode: ip2locLoc.CountryCode,
		}
	}

	return response, nil
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize application
	app, err := NewApp(config.MaxMindDBPath, config.IP2LocationPath)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	// Start gRPC server
	go func() {
		grpcAddr := fmt.Sprintf("%s:%s", config.Host, config.GRPCPort)
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("Failed to listen for gRPC: %v", err)
		}

		grpcServer := grpc.NewServer()
		pb.RegisterIP2LocationServiceServer(grpcServer, &GRPCServer{app: app})

		log.Printf("gRPC server starting on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP server
	address := fmt.Sprintf("%s:%s", config.Host, config.Port)
	log.Printf("HTTP server starting on %s!", address)

	if err := app.fiber.Listen(address); err != nil {
		log.Fatal(err)
	}
} 