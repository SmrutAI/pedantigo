// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"net"
)

// Network constraint types.
type (
	ipv4Constraint struct{}
	ipv6Constraint struct{}
)

// ipv4Constraint validates that a string is a valid IPv4 address.
func (c ipv4Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("ipv4 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse IP address
	ip := net.ParseIP(str)
	if ip == nil {
		return fmt.Errorf("must be a valid IPv4 address")
	}

	// Check if it's IPv4 (not IPv6)
	// IPv4 addresses return non-nil from To4()
	if ip.To4() == nil {
		return fmt.Errorf("must be a valid IPv4 address")
	}

	return nil
}

// ipv6Constraint validates that a string is a valid IPv6 address.
func (c ipv6Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("ipv6 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse IP address
	ip := net.ParseIP(str)
	if ip == nil {
		return fmt.Errorf("must be a valid IPv6 address")
	}

	// Check if it's IPv6 (not IPv4)
	// IPv6 addresses return nil from To4()
	if ip.To4() != nil {
		return fmt.Errorf("must be a valid IPv6 address")
	}

	return nil
}
