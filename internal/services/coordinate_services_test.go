package services

import "testing"

func TestCoordinateService_RoundToNearestMultipleOf4(t *testing.T) {
	service := NewCoordinateService()

	tests := []struct {
		name     string
		value    float64
		expected float64
	}{
		{"exact multiple stays", 100.0, 100.0},
		{"zero stays", 0.0, 0.0},
		{"rounds up past midpoint", 2.1, 4.0},
		{"rounds down before midpoint", 1.9, 0.0},
		{"rounds to nearest below", 101.33, 100.0},
		{"rounds to nearest above", 102.7, 104.0},
		{"half rounds away from zero", 2.0, 4.0},
		{"negative rounds to nearest", -1.9, 0.0},
		{"negative rounds down", -2.1, -4.0},
		{"fractional near multiple", 99.6, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.RoundToNearestMultipleOf4(tt.value)
			if result != tt.expected {
				t.Errorf(
					"RoundToNearestMultipleOf4(%v) = %v, expected %v",
					tt.value,
					result,
					tt.expected,
				)
			}
		})
	}
}
