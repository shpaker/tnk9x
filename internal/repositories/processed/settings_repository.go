package processed

import (
	"encoding/json"
	"fmt"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// settingsKey — ключ сохранения настроек в хранилище
const settingsKey = "settings"

// settingsFormatVersion — версия формата сохранения
const settingsFormatVersion = 1

var _ interfaces.ISettingsRepository = (*SettingsRepository)(nil)

// SettingsRepository сериализует пользовательские настройки в JSON
// поверх пользовательского хранилища
type SettingsRepository struct {
	storage interfaces.IStorageRepository
}

func NewSettingsRepository(
	storage interfaces.IStorageRepository,
) *SettingsRepository {
	return &SettingsRepository{
		storage: storage,
	}
}

// settingsSchema: поля-указатели — отсутствующий ключ оставляет
// значение по умолчанию
type settingsSchema struct {
	Version    int   `json:"version"`
	Effects    *bool `json:"effects,omitempty"`
	Fullscreen *bool `json:"fullscreen,omitempty"`
	Volume     *int  `json:"volume,omitempty"`
	Players    *uint `json:"players,omitempty"`
}

// GetSettings читает настройки; отсутствие сохранения — дефолты
func (sr *SettingsRepository) GetSettings() (*types.SettingsEntity, error) {
	settings := types.NewSettingsEntity()

	data, err := sr.storage.Load(settingsKey)
	if err != nil {
		return settings, err
	}
	if len(data) == 0 {
		return settings, nil
	}

	var schema settingsSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return settings, fmt.Errorf("malformed settings: %w", err)
	}
	if schema.Version != settingsFormatVersion {
		return settings, fmt.Errorf(
			"unsupported settings version %d", schema.Version,
		)
	}

	if schema.Effects != nil {
		settings.SetEffectsEnabled(*schema.Effects)
	}
	if schema.Fullscreen != nil {
		settings.SetFullscreen(*schema.Fullscreen)
	}
	if schema.Volume != nil {
		settings.SetVolumeLevel(*schema.Volume)
	}
	if schema.Players != nil {
		settings.SetPlayers(*schema.Players)
	}
	return settings, nil
}

func (sr *SettingsRepository) SaveSettings(
	settings *types.SettingsEntity,
) error {
	effects := settings.IsEffectsEnabled()
	fullscreen := settings.IsFullscreen()
	volume := settings.GetVolumeLevel()
	players := settings.GetPlayers()

	data, err := json.Marshal(settingsSchema{
		Version:    settingsFormatVersion,
		Effects:    &effects,
		Fullscreen: &fullscreen,
		Volume:     &volume,
		Players:    &players,
	})
	if err != nil {
		return err
	}
	return sr.storage.Save(settingsKey, data)
}
