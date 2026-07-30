package types

// SoundAction — тип звукового события в очереди кадра
type SoundAction uint8

const (
	SoundActionPlay SoundAction = iota
	SoundActionPlayLoop
	SoundActionStop
	SoundActionStopAll
)

// SoundEntity — звуковое событие кадра
type SoundEntity struct {
	Action  SoundAction
	SoundID SoundID // не задан для SoundActionStopAll
}
