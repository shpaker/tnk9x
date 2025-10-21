package repositories

import (
	"errors"
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
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

func TestGetLevel_Positive(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := NewMockAssetsRepository()
	mockSpritesRepo := NewMockSpritesRepository()

	// Создаем тестовую карту 26x26 с несколькими блоками (676 блоков)
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
...........#..#...........
..........................`)

	mockRepo.AddAsset("levels/1", levelData)

	// Создаем сервис уровней
	levelsService := NewLevelsRepository(mockRepo, mockSpritesRepo)

	// Вызываем функцию
	level, err := levelsService.GetLevel(1)

	// Проверяем результат
	if err != nil {
		t.Fatalf("GetLevel вернул ошибку: %v", err)
	}

	// level - это массив, не может быть nil

	// Проверяем, что уровень имеет правильный размер массива
	if len(level) != 676 {
		t.Fatalf("Ожидался массив размером 676, получено: %d", len(level))
	}

	// Подсчитываем количество непустых блоков
	nonEmptyBlocks := 0
	for _, block := range level {
		if block.Data != nil {
			nonEmptyBlocks++
		}
	}

	// Проверяем, что есть блоки (должно быть больше 0)
	if nonEmptyBlocks == 0 {
		t.Fatal("Уровень не содержит блоков")
	}

	// Проверяем, что есть блоки типа brick
	foundBrick := false
	foundSteel := false
	for _, block := range level {
		if block.Data != nil {
			if block.Data.Name == types.Brick {
				foundBrick = true
			}
			if block.Data.Name == types.Steel {
				foundSteel = true
			}
		}
	}

	if !foundBrick {
		t.Error("Не найдены блоки типа brick")
	}

	if !foundSteel {
		t.Error("Не найдены блоки типа steel")
	}
}

func TestGetLevel_InvalidCharacter(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := NewMockAssetsRepository()
	mockSpritesRepo := NewMockSpritesRepository()

	// Создаем карту с недопустимым символом (правильного размера 26x26)
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
...........#..#........X..`)

	mockRepo.AddAsset("levels/1", levelData)

	// Создаем сервис уровней
	levelsService := NewLevelsRepository(mockRepo, mockSpritesRepo)

	// Вызываем функцию
	_, err := levelsService.GetLevel(1)

	// Проверяем, что получили ошибку
	if err == nil {
		t.Fatal("Ожидалась ошибка для недопустимого символа")
	}

	if !contains(err.Error(), "неизвестный символ") {
		t.Errorf("Ожидалась ошибка о неизвестном символе, получено: %v", err)
	}
}

func TestGetLevel_ArrayOverflow(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := NewMockAssetsRepository()
	mockSpritesRepo := NewMockSpritesRepository()

	// Создаем карту с очень большим количеством блоков (больше 676)
	// 30 строк по 30 символов = 900 блоков, что больше чем 676
	levelData := []byte(`##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################
##############################`)

	mockRepo.AddAsset("levels/1", levelData)

	// Создаем сервис уровней
	levelsService := NewLevelsRepository(mockRepo, mockSpritesRepo)

	// Вызываем функцию
	_, err := levelsService.GetLevel(1)

	// Проверяем, что получили ошибку переполнения массива
	if err == nil {
		t.Fatal("Ожидалась ошибка переполнения массива")
	}

	if !contains(err.Error(), "превышен размер массива уровня") {
		t.Errorf("Ожидалась ошибка о превышении размера массива, получено: %v", err)
	}
}

func TestILevelsService_Interface(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := NewMockAssetsRepository()
	mockSpritesRepo := NewMockSpritesRepository()

	// Создаем тестовую карту
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

	// level - это массив, не может быть nil

	// Проверяем, что уровень не пустой
	if len(level) == 0 {
		t.Fatal("Уровень пустой")
	}
}

// Вспомогательная функция для проверки содержания строки
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			contains(s[1:], substr))))
}
