package generators

import (
	"crypto/rand"
	"math/big"

	"github.com/robmorgan/infraspec/internal/config"
)

var minPasswordLength = 8

// cryptoRandInt generates a cryptographically secure random integer in range [0, limit)
func cryptoRandInt(limit int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(limit)))
	if err != nil {
		panic(err) // In practice, you might want to handle this more gracefully
	}
	return int(n.Int64())
}

// cryptoShuffle performs a cryptographically secure Fisher-Yates shuffle
func cryptoShuffle(slice []byte) {
	for i := len(slice) - 1; i > 0; i-- {
		j := cryptoRandInt(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// RandomResourceName generates a random resource name with the given prefix.
func RandomResourceName(prefix string, cfg config.RandomStringConfig) string {
	// Use default values if not specified
	length := cfg.Length
	if length == 0 {
		length = 5 // default length
	}

	charset := cfg.Charset
	if charset == "" {
		charset = "abcdef0123456789" // default charset
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[cryptoRandInt(len(charset))]
	}

	return prefix + string(result)
}

// RandomDNSName generates a DNS-compliant random name.
func RandomDNSName(prefix string, cfg config.RandomStringConfig) string {
	// DNS names can only contain letters, numbers, and hyphens
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"

	return RandomResourceName(prefix, config.RandomStringConfig{
		Length:  cfg.Length,
		Charset: charset,
	})
}

// RandomPassword generates a random password meeting common requirements.
func RandomPassword(cfg config.RandomStringConfig) string {
	// Ensure password includes required character types
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	special := "!@#$%^&*"

	length := cfg.Length
	if length < minPasswordLength {
		length = minPasswordLength // minimum safe password length
	}

	// Generate password with at least one of each required type
	password := make([]byte, length)

	// First four characters are guaranteed to have one of each type
	password[0] = lowercase[cryptoRandInt(len(lowercase))]
	password[1] = uppercase[cryptoRandInt(len(uppercase))]
	password[2] = numbers[cryptoRandInt(len(numbers))]
	password[3] = special[cryptoRandInt(len(special))]

	// Fill the rest randomly
	allChars := lowercase + uppercase + numbers + special
	for i := 4; i < length; i++ {
		password[i] = allChars[cryptoRandInt(len(allChars))]
	}

	// Shuffle the password to avoid predictable patterns
	cryptoShuffle(password)

	return string(password)
}
