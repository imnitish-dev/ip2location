package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

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

	// Get database paths with absolute paths
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	config := &Config{
		Port:            getEnv("PORT", "3000"),
		Host:            getEnv("HOST", "0.0.0.0"),
		MaxMindDBPath:   getEnv("MAXMIND_DB_PATH", filepath.Join(workDir, "MaxMind.mmdb")),
		IP2LocationPath: getEnv("IP2LOCATION_DB_PATH", filepath.Join(workDir, "IP2LOCATION.BIN")),
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
	DeviceBrowser DeviceInfo `json:"deviceBrowser,omitempty"`
	Ip string `json:"ip,omitempty"`
}

// App holds the application dependencies
type App struct {
	maxmindService *ip2location.Service
	ip2locService  *ip2location.Service
	fiber          *fiber.App
}

// NewApp initializes the application
func NewApp(maxmindPath, ip2locPath string) (*App, error) {
	var app App

	maxmindService, err := ip2location.NewService(ip2location.MaxMindProvider, maxmindPath)
	if err != nil {
		log.Printf("Warning: Failed to initialize MaxMind service: %v", err)
	}
	app.maxmindService = maxmindService

	ip2locService, err := ip2location.NewService(ip2location.IP2LocationProvider, ip2locPath)
	if err != nil {
		log.Printf("Warning: Failed to initialize IP2Location service: %v", err)
	}
	app.ip2locService = ip2locService

	app.fiber = fiber.New(fiber.Config{
		ErrorHandler:          errorHandler,
		JSONEncoder:          json.Marshal,
		JSONDecoder:          json.Unmarshal,
		DisableStartupMessage: true,
	})

	// Only proceed if at least one service is initialized
	if app.maxmindService == nil && app.ip2locService == nil {
		return nil, fmt.Errorf("failed to initialize both MaxMind and IP2Location services")
	}

	app.setupRoutes()
	return &app, nil
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
	a.fiber.Get("/", a.handleIp)
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

	deviceBrowser := getDeviceInfo(c.Get("User-Agent"))

	return c.JSON(Response{
		MaxMind:     maxmindLoc,
		IP2Location: ip2locLoc,
		DeviceBrowser: deviceBrowser,
	})
}

func (a *App) lookupConcurrent(ip string) (*ip2location.Location, *ip2location.Location) {
	var (
		wg          sync.WaitGroup
		maxmindLoc  *ip2location.Location
		ip2locLoc   *ip2location.Location
		maxmindErr  error
		ip2locErr   error
	)

	// Start both lookups concurrently
	wg.Add(2)

	// MaxMind lookup
	go func() {
		defer wg.Done()
		if a.maxmindService != nil {
			maxmindLoc, maxmindErr = a.maxmindService.Lookup(ip)
		} else {
			log.Printf("Warning: MaxMind service is not initialized")
		}
	}()

	// IP2Location lookup
	go func() {
		defer wg.Done()
		if a.ip2locService != nil {
			ip2locLoc, ip2locErr = a.ip2locService.Lookup(ip)
		} else {
			log.Printf("Warning: IP2Location service is not initialized")
		}
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
		log.Printf("Warning: Lookup timeout occurred")
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

func (a *App) handleIp(c *fiber.Ctx) error {
	// Get client IP with fallback logic
	ip := getClientIP(c)
	
	// Only proceed with external IP lookup if needed
	if isLocalIP(ip) {
		ip = getPublicIP()
	}
	
	// If we couldn't determine the IP, return error
	if ip == "" {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Message: "Could not determine valid IP address",
		})
	}

	// Concurrent lookup using existing IP
	maxmindLoc, ip2locLoc := a.lookupConcurrent(ip)

	if maxmindLoc == nil && ip2locLoc == nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{
			Message: "Failed to lookup IP address",
		})
	}

	deviceBrowser := getDeviceInfo(c.Get("User-Agent"))

	return c.JSON(Response{
		MaxMind:     maxmindLoc,
		IP2Location: ip2locLoc,
		DeviceBrowser: deviceBrowser,
		Ip: ip,
	})
}

// getClientIP attempts to get the real client IP from various headers
func getClientIP(c *fiber.Ctx) string {
	// First try the standard IP
	if ip := c.IP(); !isLocalIP(ip) {
		return ip
	}
	
	// Try X-Forwarded-For
	if forwardedFor := c.Get("x-forwarded-for"); forwardedFor != "" {
		// Take the first IP if there are multiple
		if idx := strings.Index(forwardedFor, ","); idx != -1 {
			return strings.TrimSpace(forwardedFor[:idx])
		}
		return forwardedFor
	}
	
	// Finally try X-Client-IP
	return c.Get("x-client-ip")
}

// isLocalIP checks if the IP is a localhost address
func isLocalIP(ip string) bool {
	return ip == "" || ip == "127.0.0.1" || ip == "::1" || ip == "localhost"
}

// getPublicIP fetches the public IP using ipify with timeout
func getPublicIP() string {
	log.Println("Getting public IP from ipify")
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	
	return string(body)
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