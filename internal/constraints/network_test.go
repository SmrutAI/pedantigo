package constraints

import (
	"testing"
)

// TestIpv4Constraint tests ipv4Constraint.Validate() for valid IPv4 addresses.
func TestIpv4Constraint(t *testing.T) {
	runSimpleConstraintTests(t, ipv4Constraint{}, []simpleTestCase{
		// Valid IPv4 addresses
		{"valid IPv4 - localhost", "127.0.0.1", false},
		{"valid IPv4 - private range", "192.168.1.1", false},
		{"valid IPv4 - private range 10", "10.0.0.1", false},
		{"valid IPv4 - private range 172", "172.16.0.1", false},
		{"valid IPv4 - zeros", "0.0.0.0", false},
		{"valid IPv4 - broadcast", "255.255.255.255", false},
		{"valid IPv4 - google DNS", "8.8.8.8", false},
		{"valid IPv4 - public IP", "1.1.1.1", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid IPv4 addresses
		{"invalid IPv4 - out of range", "256.1.1.1", true},
		{"invalid IPv4 - too few octets", "192.168.1", true},
		{"invalid IPv4 - too many octets", "192.168.1.1.1", true},
		{"invalid IPv4 - letters", "192.168.a.1", true},
		{"invalid IPv4 - empty octets", "192.168..1", true},
		// IPv6 addresses - should fail
		{"IPv6 address fails", "::1", true},
		{"IPv6 full address fails", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"IPv6 compressed fails", "2001:db8::1", true},
		// Other invalid formats
		{"hostname not IP", "example.com", true},
		{"CIDR notation not IP", "192.168.1.0/24", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestIpv6Constraint tests ipv6Constraint.Validate() for valid IPv6 addresses.
func TestIpv6Constraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid IPv6 addresses - full form
		{name: "valid IPv6 - localhost loopback", value: "::1", wantErr: false},
		{name: "valid IPv6 - unspecified", value: "::", wantErr: false},
		{name: "valid IPv6 - full form", value: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", wantErr: false},

		// Valid IPv6 addresses - compressed form
		{name: "valid IPv6 - compressed", value: "2001:db8::1", wantErr: false},
		{name: "valid IPv6 - compressed form 2", value: "2001:db8::8a2e:370:7334", wantErr: false},
		{name: "valid IPv6 - link-local", value: "fe80::1", wantErr: false},
		{name: "valid IPv6 - multicast", value: "ff02::1", wantErr: false},

		// Valid IPv6 addresses - zone ID variants (some may fail depending on implementation)
		{name: "valid IPv6 - with numbers", value: "1234:5678:90ab:cdef:1234:5678:90ab:cdef", wantErr: false},

		// Empty string - should be skipped
		{name: "empty string", value: "", wantErr: false},

		// IPv4 addresses - should fail
		{name: "IPv4 localhost fails", value: "127.0.0.1", wantErr: true},
		{name: "IPv4 private fails", value: "192.168.1.1", wantErr: true},
		{name: "IPv4 google DNS fails", value: "8.8.8.8", wantErr: true},

		// Invalid IPv6 addresses
		{name: "invalid IPv6 - too many colons", value: "2001::db8:::1", wantErr: true},
		{name: "invalid IPv6 - invalid hex", value: "2001:db8::gggg", wantErr: true},
		{name: "invalid IPv6 - too many groups", value: "2001:0db8:85a3:0000:0000:8a2e:0370:7334:extra", wantErr: true},
		{name: "invalid IPv6 - incomplete", value: "2001:db8:", wantErr: true},

		// Other invalid formats
		{name: "hostname not IP", value: "example.com", wantErr: true},
		{name: "IPv6 with port fails", value: "[2001:db8::1]:8080", wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := ipv6Constraint{}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}
