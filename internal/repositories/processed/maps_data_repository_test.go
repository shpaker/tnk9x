package processed

import (
	"errors"
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// MockFileRepository - простой мок для тестирования
type MockFileRepository struct {
	files map[string][]byte
}

func NewMockFileRepository() *MockFileRepository {
	return &MockFileRepository{
		files: make(map[string][]byte),
	}
}

func (m *MockFileRepository) ReadFile(name string) ([]byte, error) {
	if data, exists := m.files[name]; exists {
		return data, nil
	}
	return nil, errors.New("file not found")
}

func (m *MockFileRepository) ReadImage(name string) (image.Image, error) {
	return nil, errors.New("not implemented")
}

func (m *MockFileRepository) AddFile(name string, data []byte) {
	m.files[name] = data
}

// MockSpritesRepository - простой мок для тестирования спрайтов
type MockSpritesRepository struct{}

func NewMockSpritesRepository() *MockSpritesRepository {
	return &MockSpritesRepository{}
}

func (m *MockSpritesRepository) GetSprite(groupID string, spriteID string) (*ebiten.Image, error) {
	// Создаем простое изображение 8x8 для тестов
	img := ebiten.NewImage(8, 8)
	return img, nil
}

func TestGetLevel_Success(t *testing.T) {
	// Создаем мок репозитория
	mockFileRepo := NewMockFileRepository()
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

	mockFileRepo.AddFile("levels/1", levelData)

	// Создаем сервис уровней
	mapsService := NewMapsDataRepository(mockFileRepo, mockSpritesRepo)

	// Вызываем функцию
	level, err := mapsService.GetLevel(1)

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
	mockFileRepo := NewMockFileRepository()
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

	mockFileRepo.AddFile("levels/1", levelData)

	// Создаем сервис уровней
	mapsService := NewMapsDataRepository(mockFileRepo, mockSpritesRepo)

	// Вызываем функцию
	_, err := mapsService.GetLevel(1)

	// Проверяем, что получили ошибку
	if err == nil {
		t.Fatal("Ожидалась ошибка для неправильного размера")
	}
}
