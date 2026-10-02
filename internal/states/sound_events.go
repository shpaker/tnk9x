package states

import (
	"log"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// playSoundEvents — единственная точка контакта сцены со звуковым
// адаптером: применяет накопленные за кадр события в порядке добавления
func playSoundEvents(
	soundUseCases interfaces.ISoundUseCases,
	soundPlayerAdapter interfaces.ISoundPlayerAdapter,
) {
	for _, event := range soundUseCases.GetEvents() {
		applySoundEvent(soundPlayerAdapter, event)
	}
	soundPlayerAdapter.Update()
}

func applySoundEvent(
	soundPlayerAdapter interfaces.ISoundPlayerAdapter,
	event types.SoundEntity,
) {
	var err error
	switch event.Action {
	case types.SoundActionPlay:
		err = soundPlayerAdapter.Play(event.SoundID)
	case types.SoundActionPlayLoop:
		err = soundPlayerAdapter.PlayLoop(event.SoundID)
	case types.SoundActionStop:
		soundPlayerAdapter.Stop(event.SoundID)
	case types.SoundActionStopAll:
		soundPlayerAdapter.StopAll()
	}
	// Ошибки воспроизведения не фатальны: логируем и продолжаем
	if err != nil {
		log.Printf("sound %q: %v", event.SoundID, err)
	}
}
