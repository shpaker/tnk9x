package processed

import (
	"errors"
	"image"
	"strings"
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
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
