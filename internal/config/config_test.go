package config

import "testing"

func TestValidateBaseURL(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		flag    string // CONNECTORS_HU_ALLOW_INSECURE
		wantErr bool
	}{
		{"https remote", "https://api.connectors.hu", "", false},
		{"https localhost", "https://localhost:8443", "", false},
		{"http remote refused", "http://api.connectors.hu", "", true},
		{"http remote refused even with the flag", "http://api.connectors.hu", "1", true},
		{"http localhost refused without the flag", "http://localhost:54321", "", true},
		{"http localhost allowed with the flag", "http://localhost:54321", "1", false},
		{"http 127.0.0.1 refused without the flag", "http://127.0.0.1:54321", "", true},
		{"http 127.0.0.1 allowed with the flag", "http://127.0.0.1:54321", "1", false},
		{"http LOCALHOST is case-insensitive", "http://LOCALHOST:54321", "1", false},
		{"flag value other than 1 does not count", "http://localhost:54321", "true", true},
		{"unknown scheme refused", "ftp://api.connectors.hu", "", true},
		{"no scheme refused", "api.connectors.hu", "", true},
		// Documented: the IPv6 loopback is not in the dev exception.
		{"http [::1] refused without the flag", "http://[::1]:54321", "", true},
		{"http [::1] refused even with the flag", "http://[::1]:54321", "1", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CONNECTORS_HU_ALLOW_INSECURE", tc.flag)
			err := validateBaseURL(tc.url)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateBaseURL(%q) with flag %q: err=%v, wantErr=%v", tc.url, tc.flag, err, tc.wantErr)
			}
		})
	}
}
