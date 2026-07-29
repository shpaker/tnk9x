package types_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
)

func TestBlockEntity_GetImageID_NilImage(t *testing.T) {
	block := &types.BlockEntity{
		Image: nil,
		Data: &types.BlockData{
			Name:     "test",
			Position: types.Position{X: 0, Y: 0},
		},
		Position: types.Position{X: 0, Y: 0},
		Altitude: types.SURFACE,
	}

	_, err := block.GetImageID()

	if err == nil {
		t.Error("Ожидалась ошибка для nil Image")
	}

	expectedError := "image is nil"
	if err.Error() != expectedError {
		t.Errorf(
			"Ожидалась ошибка '%s', получена '%s'",
			expectedError,
			err.Error(),
		)
	}
}

func TestBlockEntity_GetImageID_EmptyImageID(t *testing.T) {
	block := &types.BlockEntity{
		Image: &testutil.FakeImageProvider{ImageID: ""},
		Data: &types.BlockData{
			Name:     "test",
			Position: types.Position{X: 0, Y: 0},
		},
		Position: types.Position{X: 0, Y: 0},
		Altitude: types.SURFACE,
	}

	imageID, err := block.GetImageID()
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}
	if imageID != "" {
		t.Errorf(
			"Ожидалась пустая строка для пустого ImageID, получена '%s'",
			imageID,
		)
	}
}

func TestBlockEntity_GetImageID_ValidImageID(t *testing.T) {
	block := &types.BlockEntity{
		Image: &testutil.FakeImageProvider{ImageID: "valid"},
		Data: &types.BlockData{
			Name:     "test",
			Position: types.Position{X: 0, Y: 0},
		},
		Position: types.Position{X: 0, Y: 0},
		Altitude: types.SURFACE,
	}

	imageID, err := block.GetImageID()
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}
	if imageID != "valid" {
		t.Errorf("Ожидался ID 'valid', получен '%s'", imageID)
	}
}
