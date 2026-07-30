package processed

import (
	"errors"
	"image"
	"strings"
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

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
	mockFileRepo := NewMockFileRepository()

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

	mockTilesetRegistry := &testutil.FakeTilesetRegistry{}

	mapsService := NewMapsDataRepository(mockFileRepo, mockTilesetRegistry)

	tileBaseSize := 8
	mapEntity, err := mapsService.GetLevel(1, tileBaseSize)
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

// Вода анимируется и остаётся на SURFACE, лёд уходит на GROUND
func TestGetLevel_WaterAndIceBlocks(t *testing.T) {
	mockFileRepo := NewMockFileRepository()
	mockFileRepo.AddFile("levels/1.bcmap", []byte("~-#.\n...."))

	mapsService := NewMapsDataRepository(
		mockFileRepo,
		&testutil.FakeTilesetRegistry{},
	)

	mapEntity, err := mapsService.GetLevel(1, 8)
	if err != nil {
		t.Fatalf("GetLevel вернул ошибку: %v", err)
	}

	byName := map[types.BlockType]*types.BlockEntity{}
	for _, block := range mapEntity.GetBlocks() {
		byName[block.Data.Name] = block
	}

	water, ok := byName[types.Water]
	if !ok {
		t.Fatal("вода не распарсилась")
	}
	if water.Altitude != types.SURFACE {
		t.Errorf("вода: altitude %v, ожидался SURFACE", water.Altitude)
	}
	anim, ok := water.Image.(*image_providers.AnimationProvider)
	if !ok {
		t.Fatalf("вода: провайдер %T, ожидался AnimationProvider", water.Image)
	}
	if !anim.IsAnimating {
		t.Error("анимация воды не запущена")
	}

	ice, ok := byName[types.Ice]
	if !ok {
		t.Fatal("лёд ('-') не распарсился")
	}
	if ice.Altitude != types.GROUND {
		t.Errorf("лёд: altitude %v, ожидался GROUND", ice.Altitude)
	}

	brick, ok := byName[types.Brick]
	if !ok {
		t.Fatal("кирпич не распарсился")
	}
	if _, isStatic := brick.Image.(*image_providers.StaticProvider); !isStatic {
		t.Errorf("кирпич: провайдер %T, ожидался StaticProvider", brick.Image)
	}
}

// '=' больше не маппится на лёд
func TestGetLevel_UnknownEqualsChar(t *testing.T) {
	mockFileRepo := NewMockFileRepository()
	mockFileRepo.AddFile("levels/1.bcmap", []byte("=.\n.."))

	mapsService := NewMapsDataRepository(
		mockFileRepo,
		&testutil.FakeTilesetRegistry{},
	)

	if _, err := mapsService.GetLevel(1, 8); err == nil {
		t.Fatal("ожидалась ошибка неизвестного символа '='")
	}
}

func TestGetLevel_InvalidSize(t *testing.T) {
	mockFileRepo := NewMockFileRepository()

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

	mockTilesetRegistry := &testutil.FakeTilesetRegistry{}

	mapsService := NewMapsDataRepository(mockFileRepo, mockTilesetRegistry)

	tileBaseSize := 8
	_, err := mapsService.GetLevel(1, tileBaseSize)

	if err == nil {
		t.Fatal("Ожидалась ошибка для неправильного размера")
	}
}
