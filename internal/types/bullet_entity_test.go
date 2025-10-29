package types

import (
	"testing"
)

func TestBulletEntity_GetImageId(t *testing.T) {
	tests := []struct {
		name        string
		imageGetter IImageIdGetter
		expected    string
		expectError bool
	}{
		{
			name:        "Valid ImageGetter",
			imageGetter: &MockImageIdGetter{id: "bullet"},
			expected:    "bullet",
			expectError: false,
		},
		{
			name:        "Nil ImageGetter",
			imageGetter: nil,
			expected:    "",
			expectError: true,
		},
		{
			name:        "Empty ID",
			imageGetter: &MockImageIdGetter{id: ""},
			expected:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bullet := &BulletEntity{
				ImageGetter: tt.imageGetter,
			}

			result, err := bullet.GetImageId()

			if tt.expectError {
				if err == nil {
					t.Error("Ожидалась ошибка")
				}
			} else {
				if err != nil {
					t.Errorf("Не ожидалась ошибка: %v", err)
				}
				if result != tt.expected {
					t.Errorf("GetImageId() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestBulletEntity_GetSize(t *testing.T) {
	bullet := &BulletEntity{}
	size := bullet.GetSize()

	expectedWidth := 4
	expectedHeight := 4

	if size.Width != expectedWidth {
		t.Errorf("GetSize().Width = %v, want %v", size.Width, expectedWidth)
	}

	if size.Height != expectedHeight {
		t.Errorf("GetSize().Height = %v, want %v", size.Height, expectedHeight)
	}
}

func TestBulletEntity_GetPosition(t *testing.T) {
	expectedPos := Position{X: 100, Y: 200}
	bullet := &BulletEntity{
		Position: expectedPos,
	}

	result := bullet.GetPosition()
	if result != expectedPos {
		t.Errorf("GetPosition() = %v, want %v", result, expectedPos)
	}
}
