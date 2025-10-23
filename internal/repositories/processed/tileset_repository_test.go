package processed

import (
	"errors"
	"image"
	"os"
	"testing"

	"github.com/shpaker/gonflict/internal/repositories/raw"
)

// MockTilesetFileRepository - мок для тестирования TilesetRepository
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

func (m *MockTilesetFileRepository) ReadImage(name string) (image.Image, error) {
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

// Создаем тестовое изображение
func createTestImage(width, height int) image.Image {
	return image.NewRGBA(image.Rect(0, 0, width, height))
}

// Создаем тестовую конфигурацию блоков
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

// Создаем тестовую конфигурацию игрока
func createTestPlayerConfig() []byte {
	return []byte(`---
size: 16

images:
  tank_base: [0, 0]
  tank_base_2: [1, 0]

animations:
  base_tank:
    - image: tank_base
      duration: 100
    - image: tank_base_2
      duration: 100
`)
}

func TestNewTilesetRepository_Success(t *testing.T) {
	// Создаем мок репозитория
	mockFileRepo := NewMockTilesetFileRepository()

	// Добавляем тестовые файлы
	mockFileRepo.AddFile("new/blocks.yml", createTestBlocksConfig())
	mockFileRepo.AddImage("new/blocks", createTestImage(32, 32))

	// Создаем репозиторий
	repo, err := NewTilesetDataRepository(mockFileRepo, "blocks")

	// Проверяем результат
	if err != nil {
		t.Fatalf("NewTilesetRepository вернул ошибку: %v", err)
	}

	if repo == nil {
		t.Fatal("Репозиторий не создан")
	}

	// Проверяем, что изображения закешированы
	if len(repo.imagesCache) == 0 {
		t.Fatal("Изображения не закешированы")
	}

	// Проверяем, что данные анимаций загружены (если есть)
	// В createTestBlocksConfig() анимаций нет, поэтому проверяем только что поле существует
	if repo.animationsData == nil {
		t.Fatal("Поле animationsData не инициализировано")
	}
}

func TestNewTilesetRepository_ConfigNotFound(t *testing.T) {
	// Создаем мок репозитория без конфигурации
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddImage("new/blocks", createTestImage(32, 32))

	// Создаем репозиторий
	_, err := NewTilesetDataRepository(mockFileRepo, "blocks")

	// Проверяем, что получили ошибку
	if err == nil {
		t.Fatal("Ожидалась ошибка для отсутствующей конфигурации")
	}
}

func TestNewTilesetRepository_ImageNotFound(t *testing.T) {
	// Создаем мок репозитория без изображения
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("new/blocks.yml", createTestBlocksConfig())

	// Создаем репозиторий
	_, err := NewTilesetDataRepository(mockFileRepo, "blocks")

	// Проверяем, что получили ошибку
	if err == nil {
		t.Fatal("Ожидалась ошибка для отсутствующего изображения")
	}
}

func TestGetImage_Success(t *testing.T) {
	// Создаем мок репозитория
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("new/blocks.yml", createTestBlocksConfig())
	mockFileRepo.AddImage("new/blocks", createTestImage(32, 32))

	// Создаем репозиторий
	repo, err := NewTilesetDataRepository(mockFileRepo, "blocks")
	if err != nil {
		t.Fatalf("Не удалось создать репозиторий: %v", err)
	}

	// Получаем изображение
	img, err := repo.GetImage("brick")

	// Проверяем результат
	if err != nil {
		t.Fatalf("GetImage вернул ошибку: %v", err)
	}

	if img == nil {
		t.Fatal("Изображение не получено")
	}

	// Проверяем размер изображения (8x8 для размера тайла 8)
	bounds := img.Bounds()
	if bounds.Dx() != 8 || bounds.Dy() != 8 {
		t.Errorf("Ожидался размер 8x8, получен %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGetImage_NotFound(t *testing.T) {
	// Создаем мок репозитория
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("new/blocks.yml", createTestBlocksConfig())
	mockFileRepo.AddImage("new/blocks", createTestImage(32, 32))

	// Создаем репозиторий
	repo, err := NewTilesetDataRepository(mockFileRepo, "blocks")
	if err != nil {
		t.Fatalf("Не удалось создать репозиторий: %v", err)
	}

	// Пытаемся получить несуществующее изображение
	_, err = repo.GetImage("nonexistent")

	// Проверяем, что получили ошибку
	if err == nil {
		t.Fatal("Ожидалась ошибка для несуществующего изображения")
	}

	expectedError := "image 'nonexistent' not found"
	if err.Error() != expectedError {
		t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedError, err.Error())
	}
}

func TestGetAnimationData_Success(t *testing.T) {
	// Создаем мок репозитория
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("new/player.yml", createTestPlayerConfig())
	mockFileRepo.AddImage("new/player", createTestImage(64, 32))

	// Создаем репозиторий
	repo, err := NewTilesetDataRepository(mockFileRepo, "player")
	if err != nil {
		t.Fatalf("Не удалось создать репозиторий: %v", err)
	}

	// Получаем данные анимации
	animationData, err := repo.GetAnimationData("base_tank")

	// Проверяем результат
	if err != nil {
		t.Fatalf("GetAnimationData вернул ошибку: %v", err)
	}

	if len(animationData) == 0 {
		t.Fatal("Данные анимации не получены")
	}

	// Проверяем, что первая анимация имеет правильные данные
	if animationData[0].Image != "tank_base" {
		t.Errorf("Ожидался image 'tank_base', получен '%s'", animationData[0].Image)
	}

	if animationData[0].Duration != 100 {
		t.Errorf("Ожидалась длительность 100, получена %d", animationData[0].Duration)
	}
}

func TestGetAnimationData_NotFound(t *testing.T) {
	// Создаем мок репозитория
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("new/player.yml", createTestPlayerConfig())
	mockFileRepo.AddImage("new/player", createTestImage(64, 32))

	// Создаем репозиторий
	repo, err := NewTilesetDataRepository(mockFileRepo, "player")
	if err != nil {
		t.Fatalf("Не удалось создать репозиторий: %v", err)
	}

	// Пытаемся получить несуществующую анимацию
	_, err = repo.GetAnimationData("nonexistent")

	// Проверяем, что получили ошибку
	if err == nil {
		t.Fatal("Ожидалась ошибка для несуществующей анимации")
	}

	expectedError := "animation 'nonexistent' not found"
	if err.Error() != expectedError {
		t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedError, err.Error())
	}
}

func TestGetAnimationData_EmptyFrames(t *testing.T) {
	// Создаем конфигурацию с пустыми кадрами анимации
	configWithEmptyFrames := []byte(`---
size: 16

images:
  tank_base: [0, 0]

animations:
  base_tank: []
`)

	// Создаем мок репозитория
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("new/player.yml", configWithEmptyFrames)
	mockFileRepo.AddImage("new/player", createTestImage(32, 32))

	// Создаем репозиторий
	repo, err := NewTilesetDataRepository(mockFileRepo, "player")
	if err != nil {
		t.Fatalf("Не удалось создать репозиторий: %v", err)
	}

	// Получаем данные анимации
	animationData, err := repo.GetAnimationData("base_tank")

	// Проверяем результат
	if err != nil {
		t.Fatalf("GetAnimationData вернул ошибку: %v", err)
	}

	if len(animationData) != 0 {
		t.Errorf("Ожидались пустые данные анимации, получено %d кадров", len(animationData))
	}

}

// Интеграционный тест с реальными файлами
func TestTilesetRepository_Integration(t *testing.T) {
	// Пропускаем тест, если нет реальных файлов
	if _, err := os.Stat("assets/new/blocks.yml"); os.IsNotExist(err) {
		t.Skip("Пропуск интеграционного теста: файлы assets/new/blocks.yml не найдены")
	}

	// Создаем реальный файловый репозиторий
	fileRepo := raw.NewFileRepository("assets")

	// Создаем репозиторий тайлсета
	repo, err := NewTilesetDataRepository(fileRepo, "blocks")
	if err != nil {
		t.Fatalf("Не удалось создать репозиторий: %v", err)
	}

	// Тестируем получение изображения
	img, err := repo.GetImage("brick")
	if err != nil {
		t.Fatalf("Не удалось получить изображение brick: %v", err)
	}

	if img == nil {
		t.Fatal("Изображение brick не получено")
	}

	// Проверяем размер
	bounds := img.Bounds()
	if bounds.Dx() != 8 || bounds.Dy() != 8 {
		t.Errorf("Ожидался размер 8x8, получен %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// Тест кэширования изображений
func TestTilesetRepository_Cache(t *testing.T) {
	// Создаем мок репозитория
	mockFileRepo := NewMockTilesetFileRepository()
	mockFileRepo.AddFile("new/blocks.yml", createTestBlocksConfig())
	mockFileRepo.AddImage("new/blocks", createTestImage(32, 32))

	// Создаем репозиторий
	repo, err := NewTilesetDataRepository(mockFileRepo, "blocks")
	if err != nil {
		t.Fatalf("Не удалось создать репозиторий: %v", err)
	}

	// Проверяем, что все изображения предварительно закешированы
	expectedCacheSize := 3 // brick, steel, water из createTestBlocksConfig
	if len(repo.imagesCache) != expectedCacheSize {
		t.Errorf("Ожидалось %d элементов в кэше, получено %d", expectedCacheSize, len(repo.imagesCache))
	}

	// Получаем изображение первый раз
	img1, err := repo.GetImage("brick")
	if err != nil {
		t.Fatalf("Не удалось получить изображение: %v", err)
	}

	// Проверяем, что размер кэша не изменился (предварительное кэширование)
	if len(repo.imagesCache) != expectedCacheSize {
		t.Errorf("Ожидалось %d элементов в кэше, получено %d", expectedCacheSize, len(repo.imagesCache))
	}

	// Получаем то же изображение второй раз
	img2, err := repo.GetImage("brick")
	if err != nil {
		t.Fatalf("Не удалось получить изображение: %v", err)
	}

	// Проверяем, что это тот же объект (из кэша)
	if img1 != img2 {
		t.Error("Изображения должны быть одинаковыми (из кэша)")
	}

	// Проверяем, что размер кэша все еще не изменился
	if len(repo.imagesCache) != expectedCacheSize {
		t.Errorf("Ожидалось %d элементов в кэше, получено %d", expectedCacheSize, len(repo.imagesCache))
	}

	// Получаем другое изображение
	img3, err := repo.GetImage("steel")
	if err != nil {
		t.Fatalf("Не удалось получить изображение steel: %v", err)
	}

	// Проверяем, что размер кэша все еще не изменился
	if len(repo.imagesCache) != expectedCacheSize {
		t.Errorf("Ожидалось %d элементов в кэше, получено %d", expectedCacheSize, len(repo.imagesCache))
	}

	// Проверяем, что это разные объекты
	if img1 == img3 {
		t.Error("Изображения brick и steel должны быть разными")
	}
}
