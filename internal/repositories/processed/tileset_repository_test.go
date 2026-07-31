package processed

import (
	"errors"
	"image"
	"os"
	"testing"

	"github.com/shpaker/tnk9x/internal/repositories/raw"
	"github.com/shpaker/tnk9x/internal/types"
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

// createTestRegistry собирает реестр с одним тайлсетом заданного типа
func createTestRegistry(
	mockFileRepo *MockTilesetFileRepository,
	tilesetType types.TilesetType,
	tilesetName string,
) (*TilesetRepositoryRegistry, error) {
	repo, err := NewTilesetDataRepository(mockFileRepo, tilesetName)
	if err != nil {
		return nil, err
	}
	return &TilesetRepositoryRegistry{
		tilesets: map[types.TilesetType]*TilesetDataRepository{
			tilesetType: repo,
		},
	}, nil
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

func TestGetImageData_HUD(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("hud.yml", createTestHUDConfig())
	mockFileRepo.AddImage("hud", createTestImage(40, 16))

	registry, err := createTestRegistry(
		mockFileRepo,
		types.TilesetTypeHUD,
		"hud",
	)
	if err != nil {
		t.Fatalf("Не удалось создать реестр: %v", err)
	}

	img, err := registry.GetImageData(types.TilesetTypeHUD, "enemy_icon")
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

	_, err = registry.GetImageData(types.TilesetTypeHUD, "nonexistent")
	if err == nil {
		t.Error("Ожидалась ошибка для несуществующего изображения")
	}
}

func TestGetImageData_UnknownTileset(t *testing.T) {
	registry := &TilesetRepositoryRegistry{}

	_, err := registry.GetImageData(types.TilesetTypeHUD, "enemy_icon")
	if err == nil {
		t.Error("Ожидалась ошибка для незагруженного тайлсета")
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

func TestGetImageData_Success(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("blocks.yml", createTestBlocksConfig())
	mockFileRepo.AddImage("blocks", createTestImage(32, 32))

	registry, err := createTestRegistry(
		mockFileRepo,
		types.TilesetTypeBlocks,
		"blocks",
	)
	if err != nil {
		t.Fatalf("Не удалось создать реестр: %v", err)
	}

	img, err := registry.GetImageData(types.TilesetTypeBlocks, "brick")
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

func TestGetImageData_NotFound(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("blocks.yml", createTestBlocksConfig())
	mockFileRepo.AddImage("blocks", createTestImage(32, 32))

	registry, err := createTestRegistry(
		mockFileRepo,
		types.TilesetTypeBlocks,
		"blocks",
	)
	if err != nil {
		t.Fatalf("Не удалось создать реестр: %v", err)
	}

	_, err = registry.GetImageData(types.TilesetTypeBlocks, "nonexistent")

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

func TestGetImageIDs(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("blocks.yml", createTestBlocksConfig())
	mockFileRepo.AddImage("blocks", createTestImage(32, 32))

	registry, err := createTestRegistry(
		mockFileRepo,
		types.TilesetTypeBlocks,
		"blocks",
	)
	if err != nil {
		t.Fatalf("Не удалось создать реестр: %v", err)
	}

	ids := registry.GetImageIDs(types.TilesetTypeBlocks)
	expected := []string{"brick", "forest", "steel"}
	if len(ids) != len(expected) {
		t.Fatalf("Ожидалось %d id, получено %v", len(expected), ids)
	}
	for i, id := range expected {
		if ids[i] != id {
			t.Errorf("id[%d]: ожидался %q, получен %q", i, id, ids[i])
		}
	}

	if ids := registry.GetImageIDs(types.TilesetTypeHUD); ids != nil {
		t.Errorf("Для незагруженного тайлсета ожидался nil, получено %v", ids)
	}
}

func TestGetAnimationData_Success(t *testing.T) {
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("player.yml", createTestPlayerConfig())
	mockFileRepo.AddImage("player", createTestImage(64, 32))

	registry, err := createTestRegistry(
		mockFileRepo,
		types.TilesetTypePlayer,
		"player",
	)
	if err != nil {
		t.Fatalf("Не удалось создать реестр: %v", err)
	}

	animationData, err := registry.GetAnimationData(
		types.TilesetTypePlayer,
		"base_tank",
	)
	if err != nil {
		t.Fatalf("GetAnimationData вернул ошибку: %v", err)
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

	registry, err := createTestRegistry(
		mockFileRepo,
		types.TilesetTypePlayer,
		"player",
	)
	if err != nil {
		t.Fatalf("Не удалось создать реестр: %v", err)
	}

	_, err = registry.GetAnimationData(
		types.TilesetTypePlayer,
		"nonexistent",
	)

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

	registry, err := createTestRegistry(
		mockFileRepo,
		types.TilesetTypePlayer,
		"player",
	)
	if err != nil {
		t.Fatalf("Не удалось создать реестр: %v", err)
	}

	animationData, err := registry.GetAnimationData(
		types.TilesetTypePlayer,
		"base_tank",
	)
	if err != nil {
		t.Fatalf("GetAnimationData вернул ошибку: %v", err)
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

	fileRepo := raw.NewFileRepository(os.DirFS("assets"))

	registry, err := NewTilesetRepositoryRegistry(fileRepo)
	if err != nil {
		t.Fatalf("Не удалось создать реестр: %v", err)
	}

	img, err := registry.GetImageData(types.TilesetTypeBlocks, "brick")
	if err != nil {
		t.Fatalf("Не удалось получить изображение brick: %v", err)
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

	registry, err := createTestRegistry(
		mockFileRepo,
		types.TilesetTypeBlocks,
		"blocks",
	)
	if err != nil {
		t.Fatalf("Не удалось создать реестр: %v", err)
	}

	blocks := registry.tilesets[types.TilesetTypeBlocks]
	expectedCacheSize := 3
	if len(blocks.imagesCache) != expectedCacheSize {
		t.Errorf(
			"Ожидалось %d элементов в кэше, получено %d",
			expectedCacheSize,
			len(blocks.imagesCache),
		)
	}

	img1, err := registry.GetImageData(types.TilesetTypeBlocks, "brick")
	if err != nil {
		t.Fatalf("Не удалось получить изображение: %v", err)
	}

	img2, err := registry.GetImageData(types.TilesetTypeBlocks, "brick")
	if err != nil {
		t.Fatalf("Не удалось получить изображение: %v", err)
	}

	if img1 != img2 {
		t.Error("Изображения должны быть одинаковыми (из кэша)")
	}

	img3, err := registry.GetImageData(types.TilesetTypeBlocks, "steel")
	if err != nil {
		t.Fatalf("Не удалось получить изображение steel: %v", err)
	}

	if img1 == img3 {
		t.Error("Изображения brick и steel должны быть разными")
	}

	if len(blocks.imagesCache) != expectedCacheSize {
		t.Errorf(
			"Ожидалось %d элементов в кэше, получено %d",
			expectedCacheSize,
			len(blocks.imagesCache),
		)
	}
}
