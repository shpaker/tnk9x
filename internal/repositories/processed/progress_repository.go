package processed

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Ключи сохранения прогресса: у одиночной игры и игры вдвоём
// раздельное прохождение кампании
const (
	ProgressKeyOnePlayer  = "progress"
	ProgressKeyTwoPlayers = "progress-2p"
)

// progressFormatVersion — версия формата сохранения
const progressFormatVersion = 1

var _ interfaces.IProgressRepository = (*ProgressRepository)(nil)

// ProgressRepository сериализует прогресс кампании в JSON
// поверх пользовательского хранилища
type ProgressRepository struct {
	storage interfaces.IStorageRepository
	key     string
}

// NewProgressRepository — прогресс режима под ключом key
func NewProgressRepository(
	storage interfaces.IStorageRepository,
	key string,
) *ProgressRepository {
	return &ProgressRepository{
		storage: storage,
		key:     key,
	}
}

type progressSchema struct {
	Version int             `json:"version"`
	Stars   map[string]uint `json:"stars"`
}

// GetProgress читает прогресс; отсутствие сохранения — пустой прогресс
func (pr *ProgressRepository) GetProgress() (*types.ProgressEntity, error) {
	data, err := pr.storage.Load(pr.key)
	if err != nil {
		return types.NewProgressEntity(), err
	}
	if len(data) == 0 {
		return types.NewProgressEntity(), nil
	}

	var schema progressSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return types.NewProgressEntity(),
			fmt.Errorf("malformed progress: %w", err)
	}
	if schema.Version != progressFormatVersion {
		return types.NewProgressEntity(), fmt.Errorf(
			"unsupported progress version %d", schema.Version,
		)
	}

	stars := make(map[int]uint, len(schema.Stars))
	for key, value := range schema.Stars {
		level, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		stars[level] = value
	}
	return types.NewProgressEntityFromStars(stars), nil
}

func (pr *ProgressRepository) SaveProgress(
	progress *types.ProgressEntity,
) error {
	schema := progressSchema{
		Version: progressFormatVersion,
		Stars:   make(map[string]uint),
	}
	for level, value := range progress.AllStars() {
		schema.Stars[strconv.Itoa(level)] = value
	}

	data, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	return pr.storage.Save(pr.key, data)
}
