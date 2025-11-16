package processed

import (
	"fmt"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

type ScriptsRepository struct {
	fileRepository interfaces.IFileRepository
}

func NewScriptsRepository(
	fileRepository interfaces.IFileRepository,
) *ScriptsRepository {
	return &ScriptsRepository{
		fileRepository: fileRepository,
	}
}

func (sr *ScriptsRepository) GetScript(name string) (string, error) {
	scriptPath := fmt.Sprintf("scripts/%s.lua", name)
	scriptData, err := sr.fileRepository.ReadFile(scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read script '%s': %w", name, err)
	}

	script := string(scriptData)

	return script, nil
}
