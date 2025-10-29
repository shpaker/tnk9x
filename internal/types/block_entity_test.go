package types

import (
	"fmt"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// MockTilesetRepository для тестирования
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

// MockImageIDGetter для тестирования
type MockImageIDGetter struct {
	id string
}

func (mig *MockImageIDGetter) GetImageID() (string, error) {
	return mig.id, nil
}

func TestBlockEntity_GetImageID_NilImageGetter(t *testing.T) {
	// Создаем блок с nil ImageGetter
	block := &BlockEntity{
		ImageGetter: nil,
		Data: &BlockData{
			Name:     "test",
			Position: Position{X: 0, Y: 0},
		},
		Position: Position{X: 0, Y: 0},
		Altitude: SURFACE,
	}

	// Пытаемся получить ID изображения
	_, err := block.GetImageID()

	// Проверяем, что возвращается ошибка
	if err == nil {
		t.Error("Ожидалась ошибка для nil ImageGetter")
	}

	expectedError := "ImageGetter is nil"
	if err.Error() != expectedError {
		t.Errorf(
			"Ожидалась ошибка '%s', получена '%s'",
			expectedError,
			err.Error(),
		)
	}
}

func TestBlockEntity_GetImageID_EmptyImageID(t *testing.T) {
	// Создаем блок с ImageGetter, который возвращает пустой ID
	block := &BlockEntity{
		ImageGetter: &MockImageIDGetter{id: ""},
		Data: &BlockData{
			Name:     "test",
			Position: Position{X: 0, Y: 0},
		},
		Position: Position{X: 0, Y: 0},
		Altitude: SURFACE,
	}

	// Пытаемся получить ID изображения
	imageID, err := block.GetImageID()
	// Проверяем, что ошибки нет и возвращается пустая строка
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
	// Создаем блок с валидным ImageGetter
	block := &BlockEntity{
		ImageGetter: &MockImageIDGetter{id: "valid"},
		Data: &BlockData{
			Name:     "test",
			Position: Position{X: 0, Y: 0},
		},
		Position: Position{X: 0, Y: 0},
		Altitude: SURFACE,
	}

	// Пытаемся получить ID изображения
	imageID, err := block.GetImageID()
	// Проверяем, что ошибки нет и ID получен корректно
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}
	if imageID != "valid" {
		t.Errorf("Ожидался ID 'valid', получен '%s'", imageID)
	}
}
