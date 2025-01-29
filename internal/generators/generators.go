package generators

import (
	"math/rand"
	"time"

	"github.com/robmorgan/infraspec/internal/config"
)

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
}

// RandomResourceName generates a random resource name with the given prefix
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
		result[i] = charset[rand.Intn(len(charset))]
	}

	return prefix + string(result)
}

// RandomDNSName generates a DNS-compliant random name
func RandomDNSName(prefix string, cfg config.RandomStringConfig) string {
	// DNS names can only contain letters, numbers, and hyphens
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"

	return RandomResourceName(prefix, config.RandomStringConfig{
		Length:  cfg.Length,
		Charset: charset,
	})
}

// RandomPassword generates a random password meeting common requirements
func RandomPassword(cfg config.RandomStringConfig) string {
	// Ensure password includes required character types
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	special := "!@#$%^&*"

	length := cfg.Length
	if length < 8 {
		length = 8 // minimum safe password length
	}

	// Generate password with at least one of each required type
	password := make([]byte, length)

	// First four characters are guaranteed to have one of each type
	password[0] = lowercase[rand.Intn(len(lowercase))]
	password[1] = uppercase[rand.Intn(len(uppercase))]
	password[2] = numbers[rand.Intn(len(numbers))]
	password[3] = special[rand.Intn(len(special))]

	// Fill the rest randomly
	allChars := lowercase + uppercase + numbers + special
	for i := 4; i < length; i++ {
		password[i] = allChars[rand.Intn(len(allChars))]
	}

	// Shuffle the password to avoid predictable patterns
	rand.Shuffle(len(password), func(i, j int) {
		password[i], password[j] = password[j], password[i]
	})

	return string(password)
}
