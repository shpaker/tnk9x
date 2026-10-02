package processed

import (
	"fmt"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ITextsRepository = (*TextsRepository)(nil)

// TextsRepository отдаёт файлы локалей assets/locales/<язык>.yml
type TextsRepository struct {
	fileRepository interfaces.IFileRepository
}

func NewTextsRepository(
	fileRepository interfaces.IFileRepository,
) *TextsRepository {
	return &TextsRepository{
		fileRepository: fileRepository,
	}
}

// GetLocale реализует ITextsRepository
func (tr *TextsRepository) GetLocale(language types.Language) ([]byte, error) {
	localePath := fmt.Sprintf("locales/%s.yml", language)
	data, err := tr.fileRepository.ReadFile(localePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read locale '%s': %w", language, err)
	}

	return data, nil
}
