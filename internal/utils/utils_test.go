package utils

import (
	"testing"
)

func TestRoundToEven(t *testing.T) {
	tests := []struct {
		name     string
		number   float64
		up       bool
		expected int
	}{
		{
			name:     "округление четного числа вверх",
			number:   4.0,
			up:       true,
			expected: 4,
		},
		{
			name:     "округление четного числа вниз",
			number:   4.0,
			up:       false,
			expected: 4,
		},
		{
			name:     "округление нечетного числа вверх",
			number:   3.0,
			up:       true,
			expected: 4,
		},
		{
			name:     "округление нечетного числа вниз",
			number:   3.0,
			up:       false,
			expected: 2,
		},
		{
			name:     "округление дробного числа вверх",
			number:   3.7,
			up:       true,
			expected: 4,
		},
		{
			name:     "округление дробного числа вниз",
			number:   3.7,
			up:       false,
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundToEven(tt.number, tt.up)
			if result != tt.expected {
				t.Errorf(
					"RoundToEven(%v, %v) = %v, ожидалось %v",
					tt.number,
					tt.up,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestRoundToDivisible(t *testing.T) {
	tests := []struct {
		name     string
		number   float64
		divisor  int
		up       bool
		expected int
	}{
		{
			name:     "округление до числа кратного 5 вверх",
			number:   12.0,
			divisor:  5,
			up:       true,
			expected: 15,
		},
		{
			name:     "округление до числа кратного 5 вниз",
			number:   12.0,
			divisor:  5,
			up:       false,
			expected: 10,
		},
		{
			name:     "округление до числа кратного 3 вверх",
			number:   7.0,
			divisor:  3,
			up:       true,
			expected: 9,
		},
		{
			name:     "округление до числа кратного 3 вниз",
			number:   7.0,
			divisor:  3,
			up:       false,
			expected: 6,
		},
		{
			name:     "округление дробного числа до кратного 4 вверх",
			number:   6.8,
			divisor:  4,
			up:       true,
			expected: 8,
		},
		{
			name:     "округление дробного числа до кратного 4 вниз",
			number:   6.8,
			divisor:  4,
			up:       false,
			expected: 4,
		},
		{
			name:     "округление числа уже кратного делителю",
			number:   15.0,
			divisor:  5,
			up:       true,
			expected: 15,
		},
		{
			name:     "округление числа уже кратного делителю вниз",
			number:   15.0,
			divisor:  5,
			up:       false,
			expected: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundToDivisible(tt.number, tt.divisor, tt.up)
			if result != tt.expected {
				t.Errorf(
					"RoundToDivisible(%v, %v, %v) = %v, ожидалось %v",
					tt.number,
					tt.divisor,
					tt.up,
					result,
					tt.expected,
				)
			}
		})
	}
}
