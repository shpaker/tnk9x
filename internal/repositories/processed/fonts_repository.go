package processed

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/interfaces"
)

// FontsRepository читает шрифты из файлов
type FontsRepository struct {
	fileRepository interfaces.IFileRepository
}

// NewFontsRepository создает новый репозиторий для работы со шрифтами
func NewFontsRepository(
	fileRepository interfaces.IFileRepository,
) *FontsRepository {
	return &FontsRepository{
		fileRepository: fileRepository,
	}
}

// GetFont возвращает данные шрифта по имени (без расширения .ttf)
// Читает файл каждый раз при вызове
// Ищет шрифт в папке fonts/
func (fr *FontsRepository) GetFont(name string) ([]byte, error) {
	// Загружаем шрифт из файла
	// Путь: fonts/имя.ttf (относительно базовой директории assets)
	fontPath := fmt.Sprintf("fonts/%s.ttf", name)
	fontData, err := fr.fileRepository.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read font '%s': %w", name, err)
	}

	return fontData, nil
}
