package processed

import (
	"fmt"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

var _ interfaces.IShadersRepository = (*ShadersRepository)(nil)

type ShadersRepository struct {
	fileRepository interfaces.IFileRepository
}

func NewShadersRepository(
	fileRepository interfaces.IFileRepository,
) *ShadersRepository {
	return &ShadersRepository{
		fileRepository: fileRepository,
	}
}

// GetShader реализует IShadersRepository: исходник Kage-шейдера
// из каталога shaders
func (sr *ShadersRepository) GetShader(name string) ([]byte, error) {
	shaderPath := fmt.Sprintf("shaders/%s.kage", name)
	shaderData, err := sr.fileRepository.ReadFile(shaderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read shader '%s': %w", name, err)
	}

	return shaderData, nil
}
