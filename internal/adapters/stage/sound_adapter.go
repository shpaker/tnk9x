package stage

import (
	"bytes"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ISoundPlayerAdapter = (*SoundAdapter)(nil)

type SoundAdapter struct {
	soundsRepository interfaces.ISoundsRepository
	audioContext     *audio.Context
	volume           float64
	players          map[types.SoundID]*audio.Player
	loopPlayers      map[types.SoundID]*audio.Player
	soundStreams     map[types.SoundID]*vorbis.Stream
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

	return &SoundAdapter{
		soundsRepository: soundsRepository,
		audioContext:     audioContext,
		volume:           volume,
		players:          make(map[types.SoundID]*audio.Player),
		loopPlayers:      make(map[types.SoundID]*audio.Player),
		soundStreams:     make(map[types.SoundID]*vorbis.Stream),
	}, nil
}

func (a *SoundAdapter) getOrCreateStream(
	soundID types.SoundID,
) (*vorbis.Stream, error) {
	if stream, exists := a.soundStreams[soundID]; exists {
		return stream, nil
	}

	soundData, err := a.soundsRepository.GetSound(string(soundID))
	if err != nil {
		return nil, fmt.Errorf("failed to get sound '%s': %w", soundID, err)
	}

	stream, err := vorbis.DecodeWithoutResampling(bytes.NewReader(soundData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode sound '%s': %w", soundID, err)
	}

	a.soundStreams[soundID] = stream
	return stream, nil
}

func (a *SoundAdapter) Play(soundID types.SoundID) error {
	// Получаем данные звука и декодируем заново для каждого воспроизведения
	soundData, err := a.soundsRepository.GetSound(string(soundID))
	if err != nil {
		return fmt.Errorf("failed to get sound '%s': %w", soundID, err)
	}

	stream, err := vorbis.DecodeWithoutResampling(bytes.NewReader(soundData))
	if err != nil {
		return fmt.Errorf("failed to decode sound '%s': %w", soundID, err)
	}

	player, err := a.audioContext.NewPlayer(stream)
	if err != nil {
		return fmt.Errorf("failed to create player for '%s': %w", soundID, err)
	}

	// Останавливаем предыдущий проигрыватель, если он есть
	if oldPlayer, exists := a.players[soundID]; exists {
		_ = oldPlayer.Close()
	}

	a.players[soundID] = player
	player.SetVolume(a.volume)
	player.Play()

	return nil
}

func (a *SoundAdapter) PlayLoop(soundID types.SoundID) error {
	// Используем кэшированный поток для зацикливания
	stream, err := a.getOrCreateStream(soundID)
	if err != nil {
		return err
	}

	// Создаем бесконечный поток для зацикливания
	infiniteStream := audio.NewInfiniteLoop(stream, stream.Length())

	player, err := a.audioContext.NewPlayer(infiniteStream)
	if err != nil {
		return fmt.Errorf(
			"failed to create loop player for '%s': %w",
			soundID,
			err,
		)
	}

	// Останавливаем предыдущий зацикленный проигрыватель, если он есть
	if oldPlayer, exists := a.loopPlayers[soundID]; exists {
		_ = oldPlayer.Close()
	}

	a.loopPlayers[soundID] = player
	player.SetVolume(a.volume)
	player.Play()

	return nil
}

func (a *SoundAdapter) Stop(soundID types.SoundID) error {
	if player, exists := a.players[soundID]; exists {
		_ = player.Close()
		delete(a.players, soundID)
	}

	if player, exists := a.loopPlayers[soundID]; exists {
		_ = player.Close()
		delete(a.loopPlayers, soundID)
	}

	return nil
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

func (a *SoundAdapter) Update() error {
	// Удаляем завершившиеся проигрыватели
	for soundID, player := range a.players {
		if !player.IsPlaying() {
			_ = player.Close()
			delete(a.players, soundID)
		}
	}

	return nil
}
