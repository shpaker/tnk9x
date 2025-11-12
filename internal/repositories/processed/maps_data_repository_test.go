package processed

import (
	"errors"
	"image"
	"strings"
	"testing"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

// MockTilesetRepositoryRegistry - мок фасада для тестирования
type MockTilesetRepositoryRegistry struct{}

func (mtr *MockTilesetRepositoryRegistry) GetBlocksImage(
	id string,
) (types.IImageProvider, error) {
	return &image_providers.StaticProvider{ImageID: id}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetBlocksAnimationData(
	id string,
) (types.AnimationData, error) {
	return types.AnimationData{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetBlocksAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return types.AnimationConfig{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetPlayerImage(
	id string,
) (types.IImageProvider, error) {
	return &image_providers.StaticProvider{ImageID: id}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetPlayerAnimationData(
	id string,
) (types.AnimationData, error) {
	return types.AnimationData{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetPlayerAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return types.AnimationConfig{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetEnemyImage(
	id string,
) (types.IImageProvider, error) {
	return &image_providers.StaticProvider{ImageID: id}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetEnemyAnimationData(
	id string,
) (types.AnimationData, error) {
	return types.AnimationData{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetEnemyAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return types.AnimationConfig{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetBulletImage(
	id string,
) (types.IImageProvider, error) {
	return &image_providers.StaticProvider{ImageID: id}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetBulletAnimationData(
	id string,
) (types.AnimationData, error) {
	return types.AnimationData{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetBulletAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return types.AnimationConfig{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetSpawnerImage(
	id string,
) (types.IImageProvider, error) {
	return &image_providers.StaticProvider{ImageID: id}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetSpawnerAnimationData(
	id string,
) (types.AnimationData, error) {
	return types.AnimationData{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetSpawnerAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return types.AnimationConfig{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetExplosionImage(
	id string,
) (types.IImageProvider, error) {
	return &image_providers.StaticProvider{ImageID: id}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetExplosionAnimationData(
	id string,
) (types.AnimationData, error) {
	return types.AnimationData{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetExplosionAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return types.AnimationConfig{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetHQImage(
	id string,
) (types.IImageProvider, error) {
	return &image_providers.StaticProvider{ImageID: id}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetHQAnimationData(
	id string,
) (types.AnimationData, error) {
	return types.AnimationData{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetHQAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return types.AnimationConfig{}, nil
}

func (mtr *MockTilesetRepositoryRegistry) GetImageData(
	tilesetType string,
	id string,
) (image.Image, error) {
	return nil, nil
}

// MockTilesetRepository - мок для тестирования (deprecated, используйте MockTilesetRepositoryRegistry)
type MockTilesetRepository struct{}

// MockImageProvider - мок для тестирования
type MockImageProvider struct {
	id string
}

func (mig *MockImageProvider) GetImageID() (string, error) {
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

func (m *MockFileRepository) CountFiles(
	dirPath string,
	pattern string,
) (int, error) {
	count := 0
	// Простая реализация для тестов - считаем файлы с нужным расширением
	patternExt := ""
	if strings.HasPrefix(pattern, "*") {
		patternExt = pattern[1:]
	}
	for name := range m.files {
		if strings.HasPrefix(name, dirPath+"/") {
			if patternExt != "" && strings.HasSuffix(name, patternExt) {
				count++
			}
		}
	}
	return count, nil
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

	mockFileRepo.AddFile("levels/1.bcmap", levelData)

	// Создаем сервис уровней
	mockTilesetRegistry := &MockTilesetRepositoryRegistry{}
	// Проверяем, что мок реализует интерфейс
	var _ interfaces.ITilesetRepositoryRegistry = mockTilesetRegistry
	mapsService := NewMapsDataRepository(mockFileRepo, mockTilesetRegistry)

	// Вызываем функцию
	tileBaseSize := 8
	mapEntity, err := mapsService.GetLevel(1, tileBaseSize)
	// Проверяем результат
	if err != nil {
		t.Fatalf("GetLevel вернул ошибку: %v", err)
	}

	if mapEntity == nil {
		t.Fatal("MapEntity равен nil")
	}

	blocks := mapEntity.GetBlocks()
	if len(blocks) == 0 {
		t.Fatal("Уровень пустой")
	}
}

func TestGetLevel_InvalidSize(t *testing.T) {
	// Создаем мок репозитория
	mockFileRepo := NewMockFileRepository()

	// Создаем карту с неправильным размером строки (последняя строка короче)
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
...........#..#`)

	mockFileRepo.AddFile("levels/1.bcmap", levelData)

	// Создаем сервис уровней
	mockTilesetRegistry := &MockTilesetRepositoryRegistry{}
	// Проверяем, что мок реализует интерфейс
	var _ interfaces.ITilesetRepositoryRegistry = mockTilesetRegistry
	mapsService := NewMapsDataRepository(mockFileRepo, mockTilesetRegistry)

	// Вызываем функцию
	tileBaseSize := 8
	_, err := mapsService.GetLevel(1, tileBaseSize)

	// Проверяем, что получили ошибку (неправильная длина строки)
	if err == nil {
		t.Fatal("Ожидалась ошибка для неправильного размера")
	}
}
