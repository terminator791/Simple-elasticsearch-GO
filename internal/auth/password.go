package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidHash     = errors.New("invalid hash format")
)

// PasswordConfig contains configuration for password hashing
type PasswordConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultPasswordConfig returns the default password configuration
func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// PasswordService handles password hashing and verification
type PasswordService struct {
	config *PasswordConfig
}

// NewPasswordService creates a new password service
func NewPasswordService(config *PasswordConfig) *PasswordService {
	if config == nil {
		config = DefaultPasswordConfig()
	}
	return &PasswordService{config: config}
}

// HashPassword hashes a password using Argon2id
func (p *PasswordService) HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", ErrInvalidPassword
	}

	// Generate a random salt
	salt := make([]byte, p.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Hash the password
	hash := argon2.IDKey([]byte(password), salt, p.config.Iterations, p.config.Memory, p.config.Parallelism, p.config.KeyLength)

	// Encode the salt and hash to base64
	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	// Return the encoded hash in the format: $argon2id$v=19$m=memory,t=iterations,p=parallelism$salt$hash
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		p.config.Memory, p.config.Iterations, p.config.Parallelism, saltEncoded, hashEncoded), nil
}

// VerifyPassword verifies a password against a hash
func (p *PasswordService) VerifyPassword(password, hash string) (bool, error) {
	if len(password) == 0 || len(hash) == 0 {
		return false, ErrInvalidPassword
	}

	// Parse the hash
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false, ErrInvalidHash
	}

	// Parse parameters
	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, ErrInvalidHash
	}

	// Decode salt and stored hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrInvalidHash
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, ErrInvalidHash
	}

	// Hash the password with the same parameters
	passwordHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(storedHash)))

	// Compare the hashes using constant-time comparison
	return subtle.ConstantTimeCompare(passwordHash, storedHash) == 1, nil
}

// ValidatePasswordStrength validates password strength
func (p *PasswordService) ValidatePasswordStrength(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	if len(password) > 128 {
		return errors.New("password must be less than 128 characters long")
	}

	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one digit
	hasDigit := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("password must contain at least one uppercase letter, one lowercase letter, and one digit")
	}

	return nil
}