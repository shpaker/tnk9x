package processed

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/interfaces"
)

// ScriptsRepository читает Lua скрипты из файлов
type ScriptsRepository struct {
	fileRepository interfaces.IFileRepository
}

// NewScriptsRepository создает новый репозиторий для работы с Lua скриптами
func NewScriptsRepository(
	fileRepository interfaces.IFileRepository,
) *ScriptsRepository {
	return &ScriptsRepository{
		fileRepository: fileRepository,
	}
}

// GetScript возвращает скрипт по имени
// Читает файл каждый раз при вызове
func (sr *ScriptsRepository) GetScript(name string) (string, error) {
	// Загружаем скрипт из файла
	scriptPath := fmt.Sprintf("scripts/%s.lua", name)
	scriptData, err := sr.fileRepository.ReadFile(scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read script '%s': %w", name, err)
	}

	// Конвертируем в строку
	script := string(scriptData)

	return script, nil
}
