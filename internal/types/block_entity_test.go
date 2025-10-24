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

func (mtr *MockTilesetRepository) GetAnimationData(id string) (AnimationData, error) {
	return AnimationData{}, nil
}

// MockImageIdGetter для тестирования
type MockImageIdGetter struct {
	id string
}

func (mig *MockImageIdGetter) GetImageId() (string, error) {
	return mig.id, nil
}

func TestBlockEntity_GetImageId_NilImageGetter(t *testing.T) {
	// Создаем блок с nil ImageGetter
	block := &BlockEntity{
		ImageGetter: nil,
		Data: &BlockData{
			Name:     "test",
			Position: Position{X: 0, Y: 0},
		},
		Properties: &BlockProperties{
			Collidable: true,
		},
		WorldPosition: Position{X: 0, Y: 0},
	}

	// Пытаемся получить ID изображения
	_, err := block.GetImageId()

	// Проверяем, что возвращается ошибка
	if err == nil {
		t.Error("Ожидалась ошибка для nil ImageGetter")
	}

	expectedError := "ImageGetter is nil"
	if err.Error() != expectedError {
		t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedError, err.Error())
	}
}

func TestBlockEntity_GetImageId_EmptyImageId(t *testing.T) {
	// Создаем блок с ImageGetter, который возвращает пустой ID
	block := &BlockEntity{
		ImageGetter: &MockImageIdGetter{id: ""},
		Data: &BlockData{
			Name:     "test",
			Position: Position{X: 0, Y: 0},
		},
		Properties: &BlockProperties{
			Collidable: true,
		},
		WorldPosition: Position{X: 0, Y: 0},
	}

	// Пытаемся получить ID изображения
	imageId, err := block.GetImageId()

	// Проверяем, что ошибки нет и возвращается пустая строка
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}
	if imageId != "" {
		t.Errorf("Ожидалась пустая строка для пустого ImageId, получена '%s'", imageId)
	}
}

func TestBlockEntity_GetImageId_ValidImageId(t *testing.T) {
	// Создаем блок с валидным ImageGetter
	block := &BlockEntity{
		ImageGetter: &MockImageIdGetter{id: "valid"},
		Data: &BlockData{
			Name:     "test",
			Position: Position{X: 0, Y: 0},
		},
		Properties: &BlockProperties{
			Collidable: true,
		},
		WorldPosition: Position{X: 0, Y: 0},
	}

	// Пытаемся получить ID изображения
	imageId, err := block.GetImageId()

	// Проверяем, что ошибки нет и ID получен корректно
	if err != nil {
		t.Errorf("Не ожидалась ошибка: %v", err)
	}
	if imageId != "valid" {
		t.Errorf("Ожидался ID 'valid', получен '%s'", imageId)
	}
}
