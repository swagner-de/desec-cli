package cmd

import "testing"

func TestValidateDynDNSUpdate(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		ipv4     string
		ipv6     string
		wantErr  bool
	}{
		{name: "valid", hostname: "example.dedyn.io", ipv4: "192.0.2.1", ipv6: "2001:db8::1"},
		{name: "preserve", hostname: "example.dedyn.io", ipv4: "preserve", ipv6: "preserve"},
		{name: "missing hostname", ipv4: "192.0.2.1", ipv6: "2001:db8::1", wantErr: true},
		{name: "missing ipv4", hostname: "example.dedyn.io", ipv6: "2001:db8::1", wantErr: true},
		{name: "missing ipv6", hostname: "example.dedyn.io", ipv4: "192.0.2.1", wantErr: true},
		{name: "empty addresses", hostname: "example.dedyn.io", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDynDNSUpdate(tt.hostname, tt.ipv4, tt.ipv6)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateDynDNSUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
