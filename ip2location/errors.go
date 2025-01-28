package ip2location

import "errors"

var (
	ErrInvalidIP       = errors.New("invalid IP address")
	ErrInvalidProvider = errors.New("invalid provider")
) 