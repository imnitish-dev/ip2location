package ip2location

import (
	"net"
)

func (s *Service) lookupIP2Location(ip net.IP) (*Location, error) {
	results, err := s.ip2locDB.Get_all(ip.String())
	if err != nil {
		return nil, err
	}

	location := &Location{
		Country:     results.Country_long,
		City:        results.City,
		Region:      results.Region,
		Latitude:    float64(results.Latitude),
		Longitude:   float64(results.Longitude),
		CountryCode: results.Country_short,
	}

	return location, nil
} 