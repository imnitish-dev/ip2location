package ip2location

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/ip2location/ip2location-go/v9"
	"github.com/oschwald/geoip2-golang"
)

type Location struct {
	Country     string  `json:"country"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	CountryCode string  `json:"country_code"`
}

type Provider string

const (
	MaxMindProvider   Provider = "maxmind"
	IP2LocationProvider Provider = "ip2location"
)

type Service struct {
	maxmindDB   *geoip2.Reader
	ip2locDB    *ip2location.DB
	provider    Provider
	mu          sync.RWMutex
}

// NewService creates a new IP2Location service
func NewService(provider Provider, dbPath string) (*Service, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("database path is empty")
	}

	// Check if file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file not found: %s", dbPath)
	}

	s := &Service{
		provider: provider,
	}

	var err error
	switch provider {
	case MaxMindProvider:
		s.maxmindDB, err = geoip2.Open(dbPath)
	case IP2LocationProvider:
		s.ip2locDB, err = ip2location.OpenDB(dbPath)
	default:
		return nil, ErrInvalidProvider
	}

	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.maxmindDB != nil {
		s.maxmindDB.Close()
	}
	if s.ip2locDB != nil {
		s.ip2locDB.Close()
	}
}

func (s *Service) Lookup(ipStr string) (*Location, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, ErrInvalidIP
	}

	switch s.provider {
	case MaxMindProvider:
		return s.lookupMaxMind(ip)
	case IP2LocationProvider:
		return s.lookupIP2Location(ip)
	default:
		return nil, ErrInvalidProvider
	}
} 