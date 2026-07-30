package stage

import (
	"bytes"
	"fmt"
	"io"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ISoundPlayerAdapter = (*SoundAdapter)(nil)

// SoundAdapter проигрывает звуки поверх единственного audio.Context.
// Все звуки декодируются в PCM один раз при создании адаптера; каждый
// плеер читает собственный bytes.Reader поверх неизменяемого среза,
// поэтому плееры не разделяют состояние декодера между собой
type SoundAdapter struct {
	audioContext *audio.Context
	volume       float64
	pcm          map[types.SoundID][]byte
	players      map[types.SoundID]*audio.Player
	loopPlayers  map[types.SoundID]*audio.Player
}

func NewSoundAdapter(
	soundsRepository interfaces.ISoundsRepository,
	audioContext *audio.Context,
	volume float64,
) (*SoundAdapter, error) {
	if audioContext == nil {
		return nil, fmt.Errorf("audio context is required")
	}

	// Валидация громкости: должна быть от 0.0 до 1.0
	if volume < 0.0 {
		volume = 0.0
	} else if volume > 1.0 {
		volume = 1.0
	}

	adapter := &SoundAdapter{
		audioContext: audioContext,
		volume:       volume,
		pcm:          make(map[types.SoundID][]byte),
		players:      make(map[types.SoundID]*audio.Player),
		loopPlayers:  make(map[types.SoundID]*audio.Player),
	}

	// Fail-fast: все звуки декодируются при старте, ошибки данных
	// проявляются до открытия окна, а не посреди игры
	for _, soundID := range types.AllSoundIDs() {
		soundData, err := soundsRepository.GetSound(string(soundID))
		if err != nil {
			return nil, fmt.Errorf("failed to get sound '%s': %w", soundID, err)
		}

		pcm, err := decodePCM(audioContext.SampleRate(), soundData)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to decode sound '%s': %w",
				soundID,
				err,
			)
		}
		adapter.pcm[soundID] = pcm
	}

	return adapter, nil
}

// decodePCM разворачивает ogg/vorbis в 16-битный PCM целиком в память
func decodePCM(sampleRate int, oggData []byte) ([]byte, error) {
	stream, err := vorbis.DecodeWithSampleRate(
		sampleRate,
		bytes.NewReader(oggData),
	)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(stream)
}

func (a *SoundAdapter) Play(soundID types.SoundID) error {
	pcm, exists := a.pcm[soundID]
	if !exists {
		return fmt.Errorf("unknown sound '%s'", soundID)
	}

	// Закрываем предыдущий проигрыватель до создания нового
	if oldPlayer, exists := a.players[soundID]; exists {
		_ = oldPlayer.Close()
		delete(a.players, soundID)
	}

	player := a.audioContext.NewPlayerFromBytes(pcm)
	player.SetVolume(a.volume)
	player.Play()
	a.players[soundID] = player

	return nil
}

func (a *SoundAdapter) PlayLoop(soundID types.SoundID) error {
	// Идемпотентность: уже играющий луп не перезапускается
	if player, exists := a.loopPlayers[soundID]; exists {
		if player.IsPlaying() {
			return nil
		}
		_ = player.Close()
		delete(a.loopPlayers, soundID)
	}

	pcm, exists := a.pcm[soundID]
	if !exists {
		return fmt.Errorf("unknown sound '%s'", soundID)
	}

	// Свой reader на каждый плеер — общего изменяемого состояния нет
	loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm)))

	player, err := a.audioContext.NewPlayer(loop)
	if err != nil {
		return fmt.Errorf(
			"failed to create loop player for '%s': %w",
			soundID,
			err,
		)
	}

	player.SetVolume(a.volume)
	player.Play()
	a.loopPlayers[soundID] = player

	return nil
}

func (a *SoundAdapter) Stop(soundID types.SoundID) {
	if player, exists := a.players[soundID]; exists {
		_ = player.Close()
		delete(a.players, soundID)
	}

	if player, exists := a.loopPlayers[soundID]; exists {
		_ = player.Close()
		delete(a.loopPlayers, soundID)
	}
}

func (a *SoundAdapter) StopAll() {
	for _, player := range a.players {
		if player != nil {
			_ = player.Close()
		}
	}
	a.players = make(map[types.SoundID]*audio.Player)

	for _, player := range a.loopPlayers {
		if player != nil {
			_ = player.Close()
		}
	}
	a.loopPlayers = make(map[types.SoundID]*audio.Player)
}

func (a *SoundAdapter) Update() {
	// Удаляем завершившиеся проигрыватели из обоих реестров
	for soundID, player := range a.players {
		if !player.IsPlaying() {
			_ = player.Close()
			delete(a.players, soundID)
		}
	}

	for soundID, player := range a.loopPlayers {
		if !player.IsPlaying() {
			_ = player.Close()
			delete(a.loopPlayers, soundID)
		}
	}
}
