package types

import (
	"fmt"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

type MockTilesetRepository struct{}

func (mtr *MockTilesetRepository) GetImage(id string) (*ebiten.Image, error) {
	if id == "notfound" {
		return nil, fmt.Errorf("image '%s' not found", id)
	}
	return ebiten.NewImage(8, 8), nil
}

func (mtr *MockTilesetRepository) GetAnimationData(
	id string,
) (AnimationData, error) {
	return AnimationData{}, nil
}

type MockImageProvider struct {
	id string
}

func (mig *MockImageProvider) GetImageID() (string, error) {
	return mig.id, nil
}

func TestBlockEntity_GetImageID_NilImage(t *testing.T) {
	block := &BlockEntity{
		Image: nil,
		Data: &BlockData{
			Name:     "test",
			Position: Position{X: 0, Y: 0},
		},
		Position: Position{X: 0, Y: 0},
		Altitude: SURFACE,
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
	block := &BlockEntity{
		Image: &MockImageProvider{id: ""},
		Data: &BlockData{
			Name:     "test",
			Position: Position{X: 0, Y: 0},
		},
		Position: Position{X: 0, Y: 0},
		Altitude: SURFACE,
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
	block := &BlockEntity{
		Image: &MockImageProvider{id: "valid"},
		Data: &BlockData{
			Name:     "test",
			Position: Position{X: 0, Y: 0},
		},
		Position: Position{X: 0, Y: 0},
		Altitude: SURFACE,
	}

	imageID, err := block.GetImageID()
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}
	if imageID != "valid" {
		t.Errorf("Ожидался ID 'valid', получен '%s'", imageID)
	}
}
