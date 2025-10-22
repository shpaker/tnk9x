package repositories

import (
	"errors"
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/interfaces"
)

// MockAssetsRepository - простой мок для тестирования
type MockAssetsRepository struct {
	assets map[string][]byte
}

func NewMockAssetsRepository() *MockAssetsRepository {
	return &MockAssetsRepository{
		assets: make(map[string][]byte),
	}
}

func (m *MockAssetsRepository) ReadAsset(name string) ([]byte, error) {
	if data, exists := m.assets[name]; exists {
		return data, nil
	}
	return nil, errors.New("asset not found")
}

func (m *MockAssetsRepository) ReadImage(name string) (image.Image, error) {
	return nil, errors.New("not implemented")
}

func (m *MockAssetsRepository) AddAsset(name string, data []byte) {
	m.assets[name] = data
}

// MockSpritesRepository - простой мок для тестирования спрайтов
type MockSpritesRepository struct{}

func NewMockSpritesRepository() *MockSpritesRepository {
	return &MockSpritesRepository{}
}

func (m *MockSpritesRepository) GetSprite(group_id string, sprite_id string) (*ebiten.Image, error) {
	// Создаем простое изображение 8x8 для тестов
	img := ebiten.NewImage(8, 8)
	return img, nil
}

func TestGetLevel_Success(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := NewMockAssetsRepository()
	mockSpritesRepo := NewMockSpritesRepository()

	// Создаем тестовую карту 26x26
	levelData := []byte(`..........................
..........................
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..##@@##..##..##..
..##..##..##@@##..##..##..
..##..##..##..##..##..##..
..##..##..........##..##..
..##..##..........##..##..
..........##..##..........
..........##..##..........
##..####..........####..##
@@..####..........####..@@
..........##..##..........
..........######..........
..##..##..######..##..##..
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..........##..##..
..##..##..........##..##..
..##..##...####...##..##..
...........#..#...........
...........#..#...........`)

	mockRepo.AddAsset("levels/1", levelData)

	// Создаем сервис уровней
	levelsService := NewLevelsRepository(mockRepo, mockSpritesRepo)

	// Вызываем функцию
	level, err := levelsService.GetLevel(1)

	// Проверяем результат
	if err != nil {
		t.Fatalf("GetLevel вернул ошибку: %v", err)
	}

	if len(level) == 0 {
		t.Fatal("Уровень пустой")
	}
}

func TestGetLevel_InvalidSize(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := NewMockAssetsRepository()
	mockSpritesRepo := NewMockSpritesRepository()

	// Создаем карту с неправильным размером (25 строк)
	levelData := []byte(`..........................
..........................
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..##@@##..##..##..
..##..##..##@@##..##..##..
..##..##..##..##..##..##..
..##..##..........##..##..
..##..##..........##..##..
..........##..##..........
..........##..##..........
##..####..........####..##
@@..####..........####..@@
..........##..##..........
..........######..........
..##..##..######..##..##..
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..........##..##..
..##..##..........##..##..
..##..##...####...##..##..
...........#..#...........`)

	mockRepo.AddAsset("levels/1", levelData)

	// Создаем сервис уровней
	levelsService := NewLevelsRepository(mockRepo, mockSpritesRepo)

	// Вызываем функцию
	_, err := levelsService.GetLevel(1)

	// Проверяем, что получили ошибку
	if err == nil {
		t.Fatal("Ожидалась ошибка для неправильного размера")
	}
}

func TestGetLevel_Interface(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := NewMockAssetsRepository()
	mockSpritesRepo := NewMockSpritesRepository()

	// Создаем тестовую карту 26x26
	levelData := []byte(`..........................
..........................
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..##@@##..##..##..
..##..##..##@@##..##..##..
..##..##..##..##..##..##..
..##..##..........##..##..
..##..##..........##..##..
..........##..##..........
..........##..##..........
##..####..........####..##
@@..####..........####..@@
..........##..##..........
..........######..........
..##..##..######..##..##..
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..##..##..##..##..
..##..##..........##..##..
..##..##..........##..##..
..##..##...####...##..##..
...........#..#...........
...........#..#...........`)

	mockRepo.AddAsset("levels/1", levelData)

	// Создаем сервис уровней через интерфейс
	var levelsService interfaces.ILevelsDataService = NewLevelsRepository(mockRepo, mockSpritesRepo)

	// Вызываем функцию через интерфейс
	level, err := levelsService.GetLevel(1)

	// Проверяем результат
	if err != nil {
		t.Fatalf("GetLevel вернул ошибку: %v", err)
	}

	if len(level) == 0 {
		t.Fatal("Уровень пустой")
	}
}
