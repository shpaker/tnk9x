package processed

import (
	"fmt"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

type FontsRepository struct {
	fileRepository interfaces.IFileRepository
}

func NewFontsRepository(
	fileRepository interfaces.IFileRepository,
) *FontsRepository {
	return &FontsRepository{
		fileRepository: fileRepository,
	}
}

func (fr *FontsRepository) GetFont(name string) ([]byte, error) {
	fontPath := fmt.Sprintf("fonts/%s.ttf", name)
	fontData, err := fr.fileRepository.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read font '%s': %w", name, err)
	}

	return fontData, nil
}
