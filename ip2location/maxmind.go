package ip2location

import (
	"net"
)

func (s *Service) lookupMaxMind(ip net.IP) (*Location, error) {
	record, err := s.maxmindDB.City(ip)
	if err != nil {
		return nil, err
	}

	var region string
	if len(record.Subdivisions) > 0 {
		region = record.Subdivisions[0].Names["en"]
	}

	location := &Location{
		Country:     record.Country.Names["en"],
		City:        record.City.Names["en"],
		Region:      region,
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
		CountryCode: record.Country.IsoCode,
	}

	return location, nil
} 