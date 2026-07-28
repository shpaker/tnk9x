package processed

import (
	"fmt"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

var _ interfaces.ISoundsRepository = (*SoundsRepository)(nil)

type SoundsRepository struct {
	fileRepository interfaces.IFileRepository
	cache          map[string][]byte
}

func NewSoundsRepository(
	fileRepository interfaces.IFileRepository,
) *SoundsRepository {
	return &SoundsRepository{
		fileRepository: fileRepository,
		cache:          make(map[string][]byte),
	}
}

func (sr *SoundsRepository) GetSound(name string) ([]byte, error) {
	if cached, exists := sr.cache[name]; exists {
		return cached, nil
	}

	soundPath := fmt.Sprintf("sounds/%s.ogg", name)
	data, err := sr.fileRepository.ReadFile(soundPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sound '%s': %w", name, err)
	}

	sr.cache[name] = data
	return data, nil
}
