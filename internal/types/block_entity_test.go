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

func (mig *MockImageIdGetter) GetImageId() string {
	return mig.id
}

func TestBlockEntity_GetImage_NilImageGetter(t *testing.T) {
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

	// Пытаемся получить изображение
	_, err := block.GetImage(&MockTilesetRepository{})

	// Проверяем, что ошибка выбрасывается
	if err == nil {
		t.Fatal("Ожидалась ошибка для nil ImageGetter")
	}

	expectedError := "no image getter available"
	if err.Error() != expectedError {
		t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedError, err.Error())
	}
}

func TestBlockEntity_GetImage_EmptyImageId(t *testing.T) {
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

	// Пытаемся получить изображение
	_, err := block.GetImage(&MockTilesetRepository{})

	// Проверяем, что ошибка выбрасывается
	if err == nil {
		t.Fatal("Ожидалась ошибка для пустого ImageId")
	}

	expectedError := "empty image ID"
	if err.Error() != expectedError {
		t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedError, err.Error())
	}
}

func TestBlockEntity_GetImage_ImageNotFound(t *testing.T) {
	// Создаем блок с ImageGetter, который возвращает несуществующий ID
	block := &BlockEntity{
		ImageGetter: &MockImageIdGetter{id: "notfound"},
		Data: &BlockData{
			Name:     "test",
			Position: Position{X: 0, Y: 0},
		},
		Properties: &BlockProperties{
			Collidable: true,
		},
		WorldPosition: Position{X: 0, Y: 0},
	}

	// Пытаемся получить изображение
	_, err := block.GetImage(&MockTilesetRepository{})

	// Проверяем, что ошибка выбрасывается
	if err == nil {
		t.Fatal("Ожидалась ошибка для несуществующего изображения")
	}

	expectedError := "image 'notfound' not found"
	if err.Error() != expectedError {
		t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedError, err.Error())
	}
}

func TestBlockEntity_GetImage_Success(t *testing.T) {
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

	// Пытаемся получить изображение
	image, err := block.GetImage(&MockTilesetRepository{})

	// Проверяем, что ошибки нет
	if err != nil {
		t.Fatalf("Не ожидалась ошибка: %v", err)
	}

	// Проверяем, что изображение получено
	if image == nil {
		t.Fatal("Изображение не должно быть nil")
	}
}
