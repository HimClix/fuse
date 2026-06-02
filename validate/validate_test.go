package validate

import "testing"

func TestCompareValidators(t *testing.T) {
	e := New()
	type cfg struct {
		Port  int     `validate:"required,gte=1,lte=65535"`
		Name  string  `validate:"required,min=3,max=50"`
		Rate  float64 `validate:"gte=0,lte=1"`
		Level string  `validate:"oneof=debug info warn error"`
	}

	tests := []struct {
		name    string
		input   cfg
		wantErr bool
	}{
		{"valid", cfg{Port: 8080, Name: "app", Rate: 0.5, Level: "info"}, false},
		{"port zero", cfg{Port: 0, Name: "app", Rate: 0.5, Level: "info"}, true},
		{"port too high", cfg{Port: 99999, Name: "app", Rate: 0.5, Level: "info"}, true},
		{"name too short", cfg{Port: 80, Name: "ab", Rate: 0.5, Level: "info"}, true},
		{"rate too high", cfg{Port: 80, Name: "app", Rate: 1.5, Level: "info"}, true},
		{"invalid level", cfg{Port: 80, Name: "app", Rate: 0.5, Level: "trace"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := e.Validate(&tt.input)
			if (errs != nil) != tt.wantErr {
				t.Errorf("errors = %v, wantErr = %v", errs, tt.wantErr)
			}
		})
	}
}

func TestStringValidators(t *testing.T) {
	e := New()
	type cfg struct {
		Alpha string `validate:"alpha"`
		Lower string `validate:"lowercase"`
		Pfx   string `validate:"startswith=http"`
		Sfx   string `validate:"endswith=.com"`
	}

	tests := []struct {
		name    string
		input   cfg
		wantErr bool
	}{
		{"all valid", cfg{Alpha: "abc", Lower: "hello", Pfx: "https://x", Sfx: "test.com"}, false},
		{"alpha with digits", cfg{Alpha: "abc123", Lower: "hello", Pfx: "http://x", Sfx: "x.com"}, true},
		{"not lowercase", cfg{Alpha: "abc", Lower: "Hello", Pfx: "http://x", Sfx: "x.com"}, true},
		{"wrong prefix", cfg{Alpha: "abc", Lower: "hello", Pfx: "ftp://x", Sfx: "x.com"}, true},
		{"wrong suffix", cfg{Alpha: "abc", Lower: "hello", Pfx: "http://x", Sfx: "x.org"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := e.Validate(&tt.input)
			if (errs != nil) != tt.wantErr {
				t.Errorf("errors = %v, wantErr = %v", errs, tt.wantErr)
			}
		})
	}
}

func TestFormatValidators(t *testing.T) {
	e := New()

	tests := []struct {
		name    string
		val     any
		wantErr bool
	}{
		{"valid email", &struct {
			E string `validate:"email"`
		}{"user@example.com"}, false},
		{"invalid email", &struct {
			E string `validate:"email"`
		}{"not-email"}, true},
		{"valid url", &struct {
			U string `validate:"url"`
		}{"https://example.com"}, false},
		{"invalid url", &struct {
			U string `validate:"url"`
		}{"no-scheme"}, true},
		{"valid ip", &struct {
			I string `validate:"ip"`
		}{"192.168.1.1"}, false},
		{"invalid ip", &struct {
			I string `validate:"ip"`
		}{"999.999.999.999"}, true},
		{"valid uuid", &struct {
			U string `validate:"uuid"`
		}{"550e8400-e29b-41d4-a716-446655440000"}, false},
		{"invalid uuid", &struct {
			U string `validate:"uuid"`
		}{"not-uuid"}, true},
		{"valid json", &struct {
			J string `validate:"json"`
		}{`{"a":1}`}, false},
		{"invalid json", &struct {
			J string `validate:"json"`
		}{`{bad`}, true},
		{"valid hostname_port", &struct {
			H string `validate:"hostname_port"`
		}{"localhost:8080"}, false},
		{"invalid hostname_port", &struct {
			H string `validate:"hostname_port"`
		}{"no-port"}, true},
		{"valid semver", &struct {
			S string `validate:"semver"`
		}{"1.2.3"}, false},
		{"invalid semver", &struct {
			S string `validate:"semver"`
		}{"abc"}, true},
		{"valid cidr", &struct {
			C string `validate:"cidr"`
		}{"10.0.0.0/8"}, false},
		{"valid base64", &struct {
			B string `validate:"base64"`
		}{"aGVsbG8="}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := e.Validate(tt.val)
			if (errs != nil) != tt.wantErr {
				t.Errorf("errors = %v, wantErr = %v", errs, tt.wantErr)
			}
		})
	}
}

func TestCrossFieldValidators(t *testing.T) {
	e := New()
	type cfg struct {
		Env      string `validate:"required"`
		Replicas int    `validate:"required_if=Env prod"`
		Password string `validate:"required"`
		Confirm  string `validate:"eqfield=Password"`
	}

	tests := []struct {
		name    string
		input   cfg
		wantErr bool
	}{
		{"prod with replicas", cfg{Env: "prod", Replicas: 3, Password: "abc", Confirm: "abc"}, false},
		{"prod without replicas", cfg{Env: "prod", Replicas: 0, Password: "abc", Confirm: "abc"}, true},
		{"dev without replicas (ok)", cfg{Env: "dev", Replicas: 0, Password: "abc", Confirm: "abc"}, false},
		{"password mismatch", cfg{Env: "dev", Password: "abc", Confirm: "xyz"}, true},
		{"password match", cfg{Env: "dev", Password: "abc", Confirm: "abc"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := e.Validate(&tt.input)
			if (errs != nil) != tt.wantErr {
				t.Errorf("errors = %v, wantErr = %v", errs, tt.wantErr)
			}
		})
	}
}

func TestOmitempty(t *testing.T) {
	e := New()
	type cfg struct {
		Name  string `validate:"required"`
		Email string `validate:"omitempty,email"`
	}

	tests := []struct {
		name    string
		input   cfg
		wantErr bool
	}{
		{"empty email skipped", cfg{Name: "test", Email: ""}, false},
		{"valid email", cfg{Name: "test", Email: "a@b.com"}, false},
		{"invalid non-empty email", cfg{Name: "test", Email: "bad"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := e.Validate(&tt.input)
			if (errs != nil) != tt.wantErr {
				t.Errorf("errors = %v, wantErr = %v", errs, tt.wantErr)
			}
		})
	}
}

func TestOrCombinator(t *testing.T) {
	e := New()
	type cfg struct {
		Contact string `validate:"email|url"`
	}

	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"email passes", "user@example.com", false},
		{"url passes", "https://example.com", false},
		{"neither fails", "not-either", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := e.Validate(&cfg{Contact: tt.val})
			if (errs != nil) != tt.wantErr {
				t.Errorf("errors = %v, wantErr = %v", errs, tt.wantErr)
			}
		})
	}
}

func TestCollectAllErrors(t *testing.T) {
	e := New()
	type cfg struct {
		A string `validate:"required"`
		B string `validate:"required"`
		C int    `validate:"required"`
	}
	errs := e.Validate(&cfg{})
	if len(errs) != 3 {
		t.Errorf("expected 3 errors, got %d: %v", len(errs), errs)
	}
}

func TestCustomValidator(t *testing.T) {
	e := New()
	e.Register("is_port", func(info FieldInfo) bool {
		v, ok := info.Value.(int)
		return ok && v >= 1 && v <= 65535
	})

	type cfg struct {
		Port int `validate:"is_port"`
	}

	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"valid port", 8080, false},
		{"too high", 99999, true},
		{"zero", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := e.Validate(&cfg{Port: tt.port})
			if (errs != nil) != tt.wantErr {
				t.Errorf("errors = %v, wantErr = %v", errs, tt.wantErr)
			}
		})
	}
}

func TestUniqueValidator(t *testing.T) {
	e := New()
	type cfg struct {
		Tags []string `validate:"unique"`
	}

	tests := []struct {
		name    string
		tags    []string
		wantErr bool
	}{
		{"unique", []string{"a", "b", "c"}, false},
		{"duplicate", []string{"a", "b", "a"}, true},
		{"empty", []string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := e.Validate(&cfg{Tags: tt.tags})
			if (errs != nil) != tt.wantErr {
				t.Errorf("errors = %v, wantErr = %v", errs, tt.wantErr)
			}
		})
	}
}
