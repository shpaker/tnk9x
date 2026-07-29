package types_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
)

func TestBulletEntity_GetImageID(t *testing.T) {
	tests := []struct {
		name        string
		imageGetter types.IImageProvider
		expected    string
		expectError bool
	}{
		{
			name: "Valid Image",
			imageGetter: &testutil.FakeImageProvider{
				ImageID: "bullet",
			},
			expected:    "bullet",
			expectError: false,
		},
		{
			name:        "Nil Image",
			imageGetter: nil,
			expected:    "",
			expectError: true,
		},
		{
			name:        "Empty ID",
			imageGetter: &testutil.FakeImageProvider{ImageID: ""},
			expected:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bullet := types.NewBulletEntity(
				types.Position{},
				types.Size{},
				types.SURFACE,
				tt.imageGetter,
				types.DirectionUp,
				nil,
				nil,
			)

			result, err := bullet.GetImageID()

			if tt.expectError {
				if err == nil {
					t.Error("Ожидалась ошибка")
				}
			} else {
				if err != nil {
					t.Errorf("Не ожидалась ошибка: %v", err)
				}
				if result != tt.expected {
					t.Errorf("GetImageID() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestBulletEntity_GetSize(t *testing.T) {
	bullet := types.NewBulletEntity(
		types.Position{},
		types.Size{},
		types.SURFACE,
		nil,
		types.DirectionUp,
		nil,
		nil,
	)
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
	expectedPos := types.Position{X: 100, Y: 200}
	bullet := types.NewBulletEntity(
		expectedPos,
		types.Size{},
		types.SURFACE,
		nil,
		types.DirectionUp,
		nil,
		nil,
	)

	result := bullet.GetPosition()
	if result != expectedPos {
		t.Errorf("GetPosition() = %v, want %v", result, expectedPos)
	}
}
