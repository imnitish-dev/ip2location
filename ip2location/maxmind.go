package ip2location

import (
	"net"
)

func (s *Service) lookupMaxMind(ip net.IP) (*Location, error) {
	record, err := s.maxmindDB.City(ip)
	if err != nil {
		return nil, err
	}

	location := &Location{
		Country:     record.Country.Names["en"],
		City:        record.City.Names["en"],
		Region:      record.Subdivisions[0].Names["en"],
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
		CountryCode: record.Country.IsoCode,
	}

	return location, nil
} 