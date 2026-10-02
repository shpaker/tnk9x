//go:build !js

package platform

import (
	"os"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var (
	_ interfaces.IPlatformAdapter = (*PlatformAdapter)(nil)
	_ interfaces.IRewardAdapter   = (*PlatformAdapter)(nil)
)

// PlatformAdapter — игра без площадки (десктоп): игру никто
// не приостанавливает, рекламы нет, сообщать о готовности
// и геймплее некому; язык — язык ОС
type PlatformAdapter struct{}

func NewPlatformAdapter() *PlatformAdapter {
	return &PlatformAdapter{}
}

// Update реализует IPlatformAdapter
func (a *PlatformAdapter) Update() {}

// IsSuspended реализует IPlatformAdapter
func (a *PlatformAdapter) IsSuspended() bool { return false }

// IsJustSuspended реализует IPlatformAdapter
func (a *PlatformAdapter) IsJustSuspended() bool { return false }

// Ready реализует IPlatformAdapter
func (a *PlatformAdapter) Ready() {}

// SetGameplayActive реализует IPlatformAdapter
func (a *PlatformAdapter) SetGameplayActive(bool) {}

// RequestIntermission реализует IPlatformAdapter
func (a *PlatformAdapter) RequestIntermission() {}

// GetLanguage реализует IPlatformAdapter: переменные локали
// (их задают терминал и запуск вида LANG=ru_RU.UTF-8), иначе язык
// интерфейса ОС
func (a *PlatformAdapter) GetLanguage() string {
	if language := languageFromEnv(os.Getenv); language != "" {
		return language
	}
	return osLanguage()
}

// IsRewardAvailable реализует IRewardAdapter
func (a *PlatformAdapter) IsRewardAvailable() bool { return false }

// RequestReward реализует IRewardAdapter
func (a *PlatformAdapter) RequestReward() {}

// PollReward реализует IRewardAdapter
func (a *PlatformAdapter) PollReward() types.RewardStatus {
	return types.RewardStatusNone
}
