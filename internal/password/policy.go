// Package password implements the application-wide password policy.
//
// Rules enforced (PCI DSS v4.0 + NIST SP 800-63B):
//   - Minimum 12 characters, maximum 128 characters.
//   - Must contain at least one uppercase letter, one lowercase letter, one digit, and one special character.
//   - Must not match any entry in the embedded common-passwords list.
//   - Must not contain the local part of the user's email address.
//
// The package also exports [BcryptCost], the work factor used throughout the application
// when calling bcrypt.GenerateFromPassword.
package password

import (
	_ "embed"
	"strings"
	"sync"
	"unicode"
)

// BcryptCost is the work factor used when hashing passwords.
// 12 meets financial industry security recommendations (above the bcrypt default of 10).
const BcryptCost = 12

//go:embed common_passwords.txt
var rawCommon string

var (
	once            sync.Once
	commonPasswords map[string]struct{}
)

func loadCommon() {
	once.Do(func() {
		lines := strings.Split(rawCommon, "\n")
		commonPasswords = make(map[string]struct{}, len(lines))
		for _, line := range lines {
			if line = strings.TrimSpace(line); line != "" {
				commonPasswords[line] = struct{}{}
			}
		}
	})
}

// Validate checks password against every rule in the financial-grade policy and
// returns a slice of all violations in Spanish.
// An empty slice means the password is acceptable.
// The email parameter is used to detect whether the password embeds the local part
// (the portion before "@") of the user's email address.
func Validate(password, email string) []string {
	loadCommon()

	var violations []string

	runes := []rune(password)
	length := len(runes)

	if length < 12 {
		violations = append(violations, "la contraseña debe tener al menos 12 caracteres")
	}
	if length > 128 {
		violations = append(violations, "la contraseña no puede exceder 128 caracteres")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range runes {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if !hasUpper {
		violations = append(violations, "la contraseña debe contener al menos una letra mayúscula")
	}
	if !hasLower {
		violations = append(violations, "la contraseña debe contener al menos una letra minúscula")
	}
	if !hasDigit {
		violations = append(violations, "la contraseña debe contener al menos un número")
	}
	if !hasSpecial {
		violations = append(violations, "la contraseña debe contener al menos un carácter especial")
	}

	if _, found := commonPasswords[strings.ToLower(password)]; found {
		violations = append(violations, "la contraseña es demasiado común")
	}

	if email != "" {
		if localPart := strings.SplitN(email, "@", 2)[0]; localPart != "" {
			if strings.Contains(strings.ToLower(password), strings.ToLower(localPart)) {
				violations = append(violations, "la contraseña no puede contener parte del email")
			}
		}
	}

	return violations
}
