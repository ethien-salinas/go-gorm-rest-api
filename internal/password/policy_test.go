package password

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		email          string
		wantViolations []string
		wantClean      bool
	}{
		{
			name:      "valid strong password",
			password:  "Str0ng!Pass#24",
			email:     "user@example.com",
			wantClean: true,
		},
		{
			name:           "too short",
			password:       "Sh0rt!",
			email:          "",
			wantViolations: []string{"la contraseña debe tener al menos 12 caracteres"},
		},
		{
			name:           "too long",
			password:       strings.Repeat("A1a!", 33), // 132 chars
			email:          "",
			wantViolations: []string{"la contraseña no puede exceder 128 caracteres"},
		},
		{
			name:           "missing uppercase",
			password:       "str0ng!pass#24",
			email:          "",
			wantViolations: []string{"la contraseña debe contener al menos una letra mayúscula"},
		},
		{
			name:           "missing lowercase",
			password:       "STR0NG!PASS#24",
			email:          "",
			wantViolations: []string{"la contraseña debe contener al menos una letra minúscula"},
		},
		{
			name:           "missing digit",
			password:       "Strong!Pass#XY",
			email:          "",
			wantViolations: []string{"la contraseña debe contener al menos un número"},
		},
		{
			name:           "missing special character",
			password:       "Str0ngPassXY24",
			email:          "",
			wantViolations: []string{"la contraseña debe contener al menos un carácter especial"},
		},
		{
			name:           "common password",
			password:       "password",
			email:          "",
			wantViolations: []string{"la contraseña es demasiado común"},
		},
		{
			name:           "contains email local part",
			password:       "Str0ng!john#24",
			email:          "john@example.com",
			wantViolations: []string{"la contraseña no puede contener parte del email"},
		},
		{
			name:      "email check skipped when email is empty",
			password:  "Str0ng!Pass#24",
			email:     "",
			wantClean: true,
		},
		{
			name:  "collects all violations",
			password: "short",
			email:    "",
			// short (no upper, no digit, no special, too short)
			wantClean: false,
		},
		{
			name:      "exactly 12 characters is valid",
			password:  "Exactly12!Ab",
			email:     "",
			wantClean: true,
		},
		{
			name:      "exactly 128 characters is valid",
			password:  strings.Repeat("A1a!", 32), // 128 chars
			email:     "",
			wantClean: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := Validate(tt.password, tt.email)
			if tt.wantClean {
				assert.Empty(t, violations, "expected no violations")
			} else if len(tt.wantViolations) > 0 {
				for _, want := range tt.wantViolations {
					assert.Contains(t, violations, want)
				}
			} else {
				assert.NotEmpty(t, violations, "expected at least one violation")
			}
		})
	}
}

func TestBcryptCost(t *testing.T) {
	assert.Equal(t, 12, BcryptCost, "BcryptCost must be 12 for financial compliance")
}
