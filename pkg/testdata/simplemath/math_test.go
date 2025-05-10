package simplemath

import "testing"

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 5, 3, 8},
		{"negative numbers", -2, -3, -5},
		{"mixed signs", -5, 3, -2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Add(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Add(%d, %d) = %d; want %d", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	result := Subtract(5, 3)
	if result != 2 {
		t.Errorf("Subtract(5, 3) = %d; want 2", result)
	}
}

func TestMultiply(t *testing.T) {
	result := Multiply(5, 3)
	if result != 15 {
		t.Errorf("Multiply(5, 3) = %d; want 15", result)
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"normal division", 6, 3, 2},
		{"zero divisor", 5, 0, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Divide(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Divide(%d, %d) = %d; want %d", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}
