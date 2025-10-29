package processed

import (
	"errors"
	"image"
	"testing"

	"github.com/shpaker/gonflict/internal/types"
)

// MockTilesetRepository - мок для тестирования
type MockTilesetRepository struct{}

func (mtr *MockTilesetRepository) GetImage(id string) (image.Image, error) {
	return nil, nil
}

func (mtr *MockTilesetRepository) GetAnimationData(
	id string,
) (types.AnimationData, error) {
	return types.AnimationData{}, nil
}

func (mtr *MockTilesetRepository) GetAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return types.AnimationConfig{}, nil
}

// MockImageIDGetter - мок для тестирования
type MockImageIDGetter struct {
	id string
}

func (mig *MockImageIDGetter) GetImageID() (string, error) {
	return mig.id, nil
}

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

func TestGetLevel_Success(t *testing.T) {
	// Создаем мок репозитория
	mockFileRepo := NewMockFileRepository()

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
	mockTilesetRepo := &MockTilesetRepository{}
	// Проверяем, что мок реализует интерфейс
	var _ ITilesetRepository = mockTilesetRepo
	mapsService := NewMapsDataRepository(mockFileRepo, mockTilesetRepo)

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
	mockTilesetRepo := &MockTilesetRepository{}
	// Проверяем, что мок реализует интерфейс
	var _ ITilesetRepository = mockTilesetRepo
	mapsService := NewMapsDataRepository(mockFileRepo, mockTilesetRepo)

	// Вызываем функцию
	_, err := mapsService.GetLevel(1)

	// Проверяем, что получили ошибку
	if err == nil {
		t.Fatal("Ожидалась ошибка для неправильного размера")
	}
}
