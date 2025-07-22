package collections

import (
	"crypto/rand"
	"math/big"
	"slices"
)

// Intersection returns a new slice with the common elements between a and b.
func Intersection[T comparable](a, b []T) []T {
	if len(a) == 0 || len(b) == 0 {
		return []T{}
	}

	set := make(map[T]bool)
	for _, item := range b {
		set[item] = true
	}

	var result []T
	seen := make(map[T]bool)
	for _, item := range a {
		if set[item] && !seen[item] {
			result = append(result, item)
			seen[item] = true
		}
	}

	// satisfies reflect.DeepEqual
	if len(result) == 0 {
		return []T{}
	}
	return result
}

// Subtract removes all the items in b from a.
func Subtract[T comparable](a, b []T) []T {
	if len(a) == 0 || len(b) == 0 {
		return a
	}

	var result []T
	for _, item := range a {
		if !slices.Contains(b, item) {
			result = append(result, item)
		}
	}

	// satisfies reflect.DeepEqual
	if len(result) == 0 {
		return []T{}
	}

	return result
}

// cryptoRandInt generates a cryptographically secure random integer in range [0, limit)
func cryptoRandInt(limit int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(limit)))
	if err != nil {
		panic(err) // TODO - we may wish to handle this more gracefully
	}
	return int(n.Int64())
}

// RandomElement returns a random element from the slice, and a boolean indicating whether the slice was empty.
func RandomElement[T any](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}

	index := cryptoRandInt(len(slice))
	return slice[index], true
}
