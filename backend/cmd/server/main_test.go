package main

import "testing"

func TestIsProductionEnvironment(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"production", true},
		{"Production", true},
		{"", false},
		{"development", false},
	}

	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv("TOGGLY_ENV", tc.value)
			if got := isProductionEnvironment(); got != tc.want {
				t.Fatalf("isProductionEnvironment() with TOGGLY_ENV=%q = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestAdminConfigFromEnvironmentUsesConfiguredCredentials(t *testing.T) {
	t.Setenv("TOGGLY_ENV", "")
	t.Setenv("TOGGLY_ADMIN_USERNAME", "root")
	t.Setenv("TOGGLY_ADMIN_PASSWORD", "hunter22")

	cfg := adminConfigFromEnvironment()
	if cfg.Username != "root" || cfg.Password != "hunter22" {
		t.Fatalf("expected configured admin credentials, got %+v", cfg)
	}
}

func TestAdminConfigFromEnvironmentFallsBackInDevelopment(t *testing.T) {
	t.Setenv("TOGGLY_ENV", "")
	t.Setenv("TOGGLY_ADMIN_USERNAME", "")
	t.Setenv("TOGGLY_ADMIN_PASSWORD", "")

	cfg := adminConfigFromEnvironment()
	if cfg.Username != "admin" || cfg.Password != "admin123" {
		t.Fatalf("expected default insecure admin credentials, got %+v", cfg)
	}
}
