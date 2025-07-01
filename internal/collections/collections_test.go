package collections

import (
	"reflect"
	"testing"
)

func TestIntersection(t *testing.T) {
	tests := []struct {
		name     string
		a        []int
		b        []int
		expected []int
	}{
		{
			name:     "empty slices",
			a:        []int{},
			b:        []int{},
			expected: []int{},
		},
		{
			name:     "one empty slice",
			a:        []int{1, 2, 3},
			b:        []int{},
			expected: []int{},
		},
		{
			name:     "no common elements",
			a:        []int{1, 2, 3},
			b:        []int{4, 5, 6},
			expected: []int{},
		},
		{
			name:     "duplicates in first slice",
			a:        []int{1, 2, 2, 3, 3},
			b:        []int{2, 3},
			expected: []int{2, 3},
		},
		{
			name:     "duplicates in second slice",
			a:        []int{1, 2, 3},
			b:        []int{2, 2, 3, 3},
			expected: []int{2, 3},
		},
		{
			name:     "single common element",
			a:        []int{1, 2, 3},
			b:        []int{3, 4, 5},
			expected: []int{3},
		},
		{
			name:     "all elements common",
			a:        []int{1, 2, 3},
			b:        []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Intersection(tt.a, tt.b)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Intersection() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIntersectionWithStrings(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{
			name:     "empty string slices",
			a:        []string{},
			b:        []string{},
			expected: []string{},
		},
		{
			name:     "strings with spaces",
			a:        []string{"hello world", "test", "go"},
			b:        []string{"hello world", "golang", "test case"},
			expected: []string{"hello world"},
		},
		{
			name:     "case sensitive strings",
			a:        []string{"Test", "test", "TEST"},
			b:        []string{"test", "Test"},
			expected: []string{"Test", "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Intersection(tt.a, tt.b)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Intersection() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	tests := []struct {
		name     string
		a        []int
		b        []int
		expected []int
	}{
		{
			name:     "empty slices",
			a:        []int{},
			b:        []int{},
			expected: []int{},
		},
		{
			name:     "empty first slice",
			a:        []int{},
			b:        []int{1, 2, 3},
			expected: []int{},
		},
		{
			name:     "empty second slice",
			a:        []int{1, 2, 3},
			b:        []int{},
			expected: []int{1, 2, 3},
		},
		{
			name:     "no elements to subtract",
			a:        []int{1, 2, 3},
			b:        []int{4, 5, 6},
			expected: []int{1, 2, 3},
		},
		{
			name:     "subtract all elements",
			a:        []int{1, 2, 3},
			b:        []int{1, 2, 3},
			expected: []int{},
		},
		{
			name:     "subtract some elements",
			a:        []int{1, 2, 3, 4, 5},
			b:        []int{2, 4},
			expected: []int{1, 3, 5},
		},
		{
			name:     "duplicates in first slice",
			a:        []int{1, 2, 2, 3, 3},
			b:        []int{2},
			expected: []int{1, 3, 3},
		},
		{
			name:     "duplicates in second slice",
			a:        []int{1, 2, 3},
			b:        []int{2, 2, 3, 3},
			expected: []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Subtract(tt.a, tt.b)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Subtract() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSubtractWithStrings(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{
			name:     "subtract strings with spaces",
			a:        []string{"hello world", "test case", "golang"},
			b:        []string{"test case"},
			expected: []string{"hello world", "golang"},
		},
		{
			name:     "case sensitive strings",
			a:        []string{"Test", "test", "TEST"},
			b:        []string{"test"},
			expected: []string{"Test", "TEST"},
		},
		{
			name:     "special characters",
			a:        []string{"@#$", "123", "abc"},
			b:        []string{"123", "@#$"},
			expected: []string{"abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Subtract(tt.a, tt.b)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Subtract() = %v, want %v", result, tt.expected)
			}
		})
	}
}
func TestRandomElement(t *testing.T) {
	tests := []struct {
		name         string
		slice        []interface{}
		wantValue    interface{}
		wantOk       bool
		iterations   int
		expectUnique bool
	}{
		{
			name:         "empty slice",
			slice:        []interface{}{},
			wantValue:    nil,
			wantOk:       false,
			iterations:   1,
			expectUnique: false,
		},
		{
			name:         "single element slice",
			slice:        []interface{}{42},
			wantValue:    42,
			wantOk:       true,
			iterations:   5,
			expectUnique: false,
		},
		{
			name:         "multiple elements slice",
			slice:        []interface{}{1, 2, 3, 4, 5},
			wantValue:    nil,
			wantOk:       true,
			iterations:   100,
			expectUnique: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.iterations == 1 {
				value, ok := RandomElement(tt.slice)
				if ok != tt.wantOk {
					t.Errorf("RandomElement() ok = %v, want %v", ok, tt.wantOk)
				}
				if !tt.wantOk {
					return
				}
				if value != tt.wantValue && tt.wantValue != nil {
					t.Errorf("RandomElement() = %v, want %v", value, tt.wantValue)
				}
			} else {
				seen := make(map[interface{}]bool)
				for i := 0; i < tt.iterations; i++ {
					value, ok := RandomElement(tt.slice)
					if !ok {
						t.Errorf("RandomElement() ok = false, want true")
						return
					}
					seen[value] = true
				}
				if tt.expectUnique && len(seen) < 2 {
					t.Errorf("RandomElement() generated only one unique value in %d iterations", tt.iterations)
				}
			}
		})
	}
}

func TestRandomElementTyped(t *testing.T) {
	tests := []struct {
		name   string
		slice  interface{}
		wantOk bool
	}{
		{
			name:   "string slice",
			slice:  []string{"a", "b", "c"},
			wantOk: true,
		},
		{
			name:   "float slice",
			slice:  []float64{1.1, 2.2, 3.3},
			wantOk: true,
		},
		{
			name:   "struct slice",
			slice:  []struct{ val int }{{1}, {2}, {3}},
			wantOk: true,
		},
		{
			name:   "empty struct slice",
			slice:  []struct{}{},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch slice := tt.slice.(type) {
			case []string:
				_, ok := RandomElement(slice)
				if ok != tt.wantOk {
					t.Errorf("RandomElement() ok = %v, want %v", ok, tt.wantOk)
				}
			case []float64:
				_, ok := RandomElement(slice)
				if ok != tt.wantOk {
					t.Errorf("RandomElement() ok = %v, want %v", ok, tt.wantOk)
				}
			case []struct{ val int }:
				_, ok := RandomElement(slice)
				if ok != tt.wantOk {
					t.Errorf("RandomElement() ok = %v, want %v", ok, tt.wantOk)
				}
			case []struct{}:
				_, ok := RandomElement(slice)
				if ok != tt.wantOk {
					t.Errorf("RandomElement() ok = %v, want %v", ok, tt.wantOk)
				}
			}
		})
	}
}
