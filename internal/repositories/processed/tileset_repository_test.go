package processed

import (
	"errors"
	"image"
	"os"
	"testing"

	"github.com/shpaker/tnk9x/internal/repositories/raw"
)

type MockTilesetFileRepository struct {
	files  map[string][]byte
	images map[string]image.Image
}

func NewMockTilesetFileRepository() *MockTilesetFileRepository {
	return &MockTilesetFileRepository{
		files:  make(map[string][]byte),
		images: make(map[string]image.Image),
	}
}

func (m *MockTilesetFileRepository) ReadFile(name string) ([]byte, error) {
	if data, exists := m.files[name]; exists {
		return data, nil
	}
	return nil, errors.New("file not found")
}

func (m *MockTilesetFileRepository) ReadImage(
	name string,
) (image.Image, error) {
	if img, exists := m.images[name]; exists {
		return img, nil
	}
	return nil, errors.New("image not found")
}

func (m *MockTilesetFileRepository) AddFile(name string, data []byte) {
	m.files[name] = data
}

func (m *MockTilesetFileRepository) AddImage(name string, img image.Image) {
	m.images[name] = img
}

func (m *MockTilesetFileRepository) CountFiles(
	dirPath string,
	pattern string,
) (int, error) {
	return 0, nil
}

func createTestImage(width, height int) image.Image {
	return image.NewRGBA(image.Rect(0, 0, width, height))
}

func createTestBlocksConfig() []byte {
	return []byte(`---
size: 8

images:
  brick: [0, 0]
  steel: [0, 1]
  forest: [0, 2]

animations:
`)
}

func createTestPlayerConfig() []byte {
	return []byte(`---
size: 16

images:
  tank_base: [0, 0]
  tank_base_2: [1, 0]

animations:
  base_tank:
    duration: 100
    frames: [tank_base, tank_base_2]
`)
}

func createTestRegistryWithBlocks(
	mockFileRepo *MockTilesetFileRepository,
) (*TilesetRepositoryRegistry, error) {
	blocksRepo, err := NewTilesetDataRepository(mockFileRepo, "blocks")
	if err != nil {
		return nil, err
	}
	return &TilesetRepositoryRegistry{blocks: blocksRepo}, nil
}

func createTestRegistryWithPlayer(
	mockFileRepo *MockTilesetFileRepository,
) (*TilesetRepositoryRegistry, error) {
	playerRepo, err := NewTilesetDataRepository(mockFileRepo, "player")
	if err != nil {
		return nil, err
	}
	return &TilesetRepositoryRegistry{player: playerRepo}, nil
}

func createTestHUDConfig() []byte {
	return []byte(`---
size: 8

images:
  enemy_icon: [0, 0]
  life_icon: [1, 0]

animations:
`)
}

func createTestRegistryWithHUD(
	mockFileRepo *MockTilesetFileRepository,
) (*TilesetRepositoryRegistry, error) {
	hudRepo, err := NewTilesetDataRepository(mockFileRepo, "hud")
	if err != nil {
		return nil, err
	}
	return &TilesetRepositoryRegistry{hud: hudRepo}, nil
}

func TestGetHUDImage(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("hud.yml", createTestHUDConfig())
	mockFileRepo.AddImage("hud", createTestImage(40, 16))

	registry, err := createTestRegistryWithHUD(mockFileRepo)
	if err != nil {
		t.Fatalf("Не удалось создать фасад: %v", err)
	}

	provider, err := registry.GetHUDImage("enemy_icon")
	if err != nil {
		t.Fatalf("GetHUDImage вернул ошибку: %v", err)
	}

	imageID, err := provider.GetImageID()
	if err != nil {
		t.Fatalf("GetImageID вернул ошибку: %v", err)
	}

	img, err := registry.GetImageData("hud", imageID)
	if err != nil {
		t.Fatalf("GetImageData вернул ошибку: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 8 || bounds.Dy() != 8 {
		t.Errorf(
			"Ожидался размер 8x8, получен %dx%d",
			bounds.Dx(),
			bounds.Dy(),
		)
	}

	if _, err := registry.GetHUDImage("nonexistent"); err == nil {
		t.Error("Ожидалась ошибка для несуществующего изображения")
	}
}

func TestGetHUDImage_NotInitialized(t *testing.T) {
	registry := &TilesetRepositoryRegistry{}

	if _, err := registry.GetHUDImage("enemy_icon"); err == nil {
		t.Error("Ожидалась ошибка неинициализированного репозитория")
	}
}

func TestNewTilesetRepository_Success(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()

	mockFileRepo.AddFile("blocks.yml", createTestBlocksConfig())
	mockFileRepo.AddImage("blocks", createTestImage(32, 32))

	repo, err := NewTilesetDataRepository(mockFileRepo, "blocks")
	if err != nil {
		t.Fatalf("NewTilesetRepository вернул ошибку: %v", err)
	}

	if repo == nil {
		t.Fatal("Репозиторий не создан")
	}

	if len(repo.imagesCache) == 0 {
		t.Fatal("Изображения не закешированы")
	}

	if repo.animationsData == nil {
		t.Fatal("Поле animationsData не инициализировано")
	}
}

func TestNewTilesetRepository_ConfigNotFound(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddImage("blocks", createTestImage(32, 32))

	_, err := NewTilesetDataRepository(mockFileRepo, "blocks")

	if err == nil {
		t.Fatal("Ожидалась ошибка для отсутствующей конфигурации")
	}
}

func TestNewTilesetRepository_ImageNotFound(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("blocks.yml", createTestBlocksConfig())

	_, err := NewTilesetDataRepository(mockFileRepo, "blocks")

	if err == nil {
		t.Fatal("Ожидалась ошибка для отсутствующего изображения")
	}
}

func TestGetImage_Success(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("blocks.yml", createTestBlocksConfig())
	mockFileRepo.AddImage("blocks", createTestImage(32, 32))

	registry, err := createTestRegistryWithBlocks(mockFileRepo)
	if err != nil {
		t.Fatalf("Не удалось создать фасад: %v", err)
	}

	provider, err := registry.GetBlocksImage("brick")
	if err != nil {
		t.Fatalf("GetBlocksImage вернул ошибку: %v", err)
	}

	if provider == nil {
		t.Fatal("Провайдер не получен")
	}

	imageID, err := provider.GetImageID()
	if err != nil {
		t.Fatalf("GetImageID вернул ошибку: %v", err)
	}

	img, err := registry.GetImageData("blocks", imageID)
	if err != nil {
		t.Fatalf("GetImageData вернул ошибку: %v", err)
	}

	if img == nil {
		t.Fatal("Изображение не получено")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 8 || bounds.Dy() != 8 {
		t.Errorf("Ожидался размер 8x8, получен %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGetImage_NotFound(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("blocks.yml", createTestBlocksConfig())
	mockFileRepo.AddImage("blocks", createTestImage(32, 32))

	registry, err := createTestRegistryWithBlocks(mockFileRepo)
	if err != nil {
		t.Fatalf("Не удалось создать фасад: %v", err)
	}

	_, err = registry.GetBlocksImage("nonexistent")

	if err == nil {
		t.Fatal("Ожидалась ошибка для несуществующего изображения")
	}

	expectedError := "image 'nonexistent' not found"
	if err.Error() != expectedError {
		t.Errorf(
			"Ожидалась ошибка '%s', получена '%s'",
			expectedError,
			err.Error(),
		)
	}
}

func TestGetAnimationData_Success(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("player.yml", createTestPlayerConfig())
	mockFileRepo.AddImage("player", createTestImage(64, 32))

	registry, err := createTestRegistryWithPlayer(mockFileRepo)
	if err != nil {
		t.Fatalf("Не удалось создать фасад: %v", err)
	}

	animationData, err := registry.GetPlayerAnimationData("base_tank")
	if err != nil {
		t.Fatalf("GetPlayerAnimationData вернул ошибку: %v", err)
	}

	if len(animationData) == 0 {
		t.Fatal("Данные анимации не получены")
	}

	if animationData[0].Image != "tank_base" {
		t.Errorf(
			"Ожидался image 'tank_base', получен '%s'",
			animationData[0].Image,
		)
	}

	if animationData[0].Duration != 100 {
		t.Errorf(
			"Ожидалась длительность 100, получена %d",
			animationData[0].Duration,
		)
	}

	if len(animationData) < 2 {
		t.Fatal("Ожидалось 2 кадра")
	}
	if animationData[1].Image != "tank_base_2" {
		t.Errorf(
			"Ожидался image 'tank_base_2', получен '%s'",
			animationData[1].Image,
		)
	}
}

func TestGetAnimationData_NotFound(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("player.yml", createTestPlayerConfig())
	mockFileRepo.AddImage("player", createTestImage(64, 32))

	registry, err := createTestRegistryWithPlayer(mockFileRepo)
	if err != nil {
		t.Fatalf("Не удалось создать фасад: %v", err)
	}

	_, err = registry.GetPlayerAnimationData("nonexistent")

	if err == nil {
		t.Fatal("Ожидалась ошибка для несуществующей анимации")
	}

	expectedError := "animation 'nonexistent' not found"
	if err.Error() != expectedError {
		t.Errorf(
			"Ожидалась ошибка '%s', получена '%s'",
			expectedError,
			err.Error(),
		)
	}
}

func TestGetAnimationData_EmptyFrames(t *testing.T) {
	configWithEmptyFrames := []byte(`---
size: 16

images:
  tank_base: [0, 0]

animations:
  base_tank:
    duration: 100
    frames: []
`)

	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("player.yml", configWithEmptyFrames)
	mockFileRepo.AddImage("player", createTestImage(32, 32))

	registry, err := createTestRegistryWithPlayer(mockFileRepo)
	if err != nil {
		t.Fatalf("Не удалось создать фасад: %v", err)
	}

	animationData, err := registry.GetPlayerAnimationData("base_tank")
	if err != nil {
		t.Fatalf("GetPlayerAnimationData вернул ошибку: %v", err)
	}

	if len(animationData) != 0 {
		t.Errorf(
			"Ожидались пустые данные анимации, получено %d кадров",
			len(animationData),
		)
	}
}

func TestTilesetRepository_Integration(t *testing.T) {
	if _, err := os.Stat("assets/tiles/blocks.yml"); os.IsNotExist(err) {
		t.Skip(
			"Пропуск интеграционного теста: файлы assets/tiles/blocks.yml не найдены",
		)
	}

	fileRepo := raw.NewFileRepository("assets")

	registry, err := NewTilesetRepositoryRegistry(fileRepo)
	if err != nil {
		t.Fatalf("Не удалось создать фасад: %v", err)
	}

	provider, err := registry.GetBlocksImage("brick")
	if err != nil {
		t.Fatalf("Не удалось получить изображение brick: %v", err)
	}

	if provider == nil {
		t.Fatal("Провайдер brick не получен")
	}

	imageID, err := provider.GetImageID()
	if err != nil {
		t.Fatalf("GetImageID вернул ошибку: %v", err)
	}

	img, err := registry.GetImageData("blocks", imageID)
	if err != nil {
		t.Fatalf("GetImageData вернул ошибку: %v", err)
	}

	if img == nil {
		t.Fatal("Изображение brick не получено")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 8 || bounds.Dy() != 8 {
		t.Errorf("Ожидался размер 8x8, получен %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestTilesetRepository_Cache(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("blocks.yml", createTestBlocksConfig())
	mockFileRepo.AddImage("blocks", createTestImage(32, 32))

	registry, err := createTestRegistryWithBlocks(mockFileRepo)
	if err != nil {
		t.Fatalf("Не удалось создать фасад: %v", err)
	}

	expectedCacheSize := 3
	if len(registry.blocks.imagesCache) != expectedCacheSize {
		t.Errorf(
			"Ожидалось %d элементов в кэше, получено %d",
			expectedCacheSize,
			len(registry.blocks.imagesCache),
		)
	}

	provider1, err := registry.GetBlocksImage("brick")
	if err != nil {
		t.Fatalf("Не удалось получить изображение: %v", err)
	}

	if len(registry.blocks.imagesCache) != expectedCacheSize {
		t.Errorf(
			"Ожидалось %d элементов в кэше, получено %d",
			expectedCacheSize,
			len(registry.blocks.imagesCache),
		)
	}

	imageID1, err := provider1.GetImageID()
	if err != nil {
		t.Fatalf("GetImageID вернул ошибку: %v", err)
	}

	img1, err := registry.GetImageData("blocks", imageID1)
	if err != nil {
		t.Fatalf("GetImageData вернул ошибку: %v", err)
	}

	provider2, err := registry.GetBlocksImage("brick")
	if err != nil {
		t.Fatalf("Не удалось получить изображение: %v", err)
	}

	if len(registry.blocks.imagesCache) != expectedCacheSize {
		t.Errorf(
			"Ожидалось %d элементов в кэше, получено %d",
			expectedCacheSize,
			len(registry.blocks.imagesCache),
		)
	}

	imageID2, err := provider2.GetImageID()
	if err != nil {
		t.Fatalf("GetImageID вернул ошибку: %v", err)
	}

	img2, err := registry.GetImageData("blocks", imageID2)
	if err != nil {
		t.Fatalf("GetImageData вернул ошибку: %v", err)
	}

	if img1 != img2 {
		t.Error("Изображения должны быть одинаковыми (из кэша)")
	}

	if len(registry.blocks.imagesCache) != expectedCacheSize {
		t.Errorf(
			"Ожидалось %d элементов в кэше, получено %d",
			expectedCacheSize,
			len(registry.blocks.imagesCache),
		)
	}

	provider3, err := registry.GetBlocksImage("steel")
	if err != nil {
		t.Fatalf("Не удалось получить изображение steel: %v", err)
	}

	if len(registry.blocks.imagesCache) != expectedCacheSize {
		t.Errorf(
			"Ожидалось %d элементов в кэше, получено %d",
			expectedCacheSize,
			len(registry.blocks.imagesCache),
		)
	}

	imageID3, err := provider3.GetImageID()
	if err != nil {
		t.Fatalf("GetImageID вернул ошибку: %v", err)
	}

	img3, err := registry.GetImageData("blocks", imageID3)
	if err != nil {
		t.Fatalf("GetImageData вернул ошибку: %v", err)
	}

	if img1 == img3 {
		t.Error("Изображения brick и steel должны быть разными")
	}
}
