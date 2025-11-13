package adapters

import (
	"bytes"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"

	"github.com/shpaker/tnk25/internal/interfaces"
	"github.com/shpaker/tnk25/internal/types"
)

type StageSoundAdapter struct {
	soundsRepository interfaces.ISoundsRepository
	audioContext     *audio.Context
	players          map[types.SoundID]*audio.Player
	loopPlayers      map[types.SoundID]*audio.Player
	soundStreams     map[types.SoundID]*vorbis.Stream
}

func NewStageSoundAdapter(
	soundsRepository interfaces.ISoundsRepository,
	audioContext *audio.Context,
) (*StageSoundAdapter, error) {
	if audioContext == nil {
		return nil, fmt.Errorf("audio context is required")
	}

	return &StageSoundAdapter{
		soundsRepository: soundsRepository,
		audioContext:     audioContext,
		players:          make(map[types.SoundID]*audio.Player),
		loopPlayers:      make(map[types.SoundID]*audio.Player),
		soundStreams:     make(map[types.SoundID]*vorbis.Stream),
	}, nil
}

func (a *StageSoundAdapter) getOrCreateStream(
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

func (a *StageSoundAdapter) Play(soundID types.SoundID) error {
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
	player.Play()

	return nil
}

func (a *StageSoundAdapter) PlayLoop(soundID types.SoundID) error {
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
	player.Play()

	return nil
}

func (a *StageSoundAdapter) Stop(soundID types.SoundID) error {
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

func (a *StageSoundAdapter) StopAll() {
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

func (a *StageSoundAdapter) Update() error {
	// Удаляем завершившиеся проигрыватели
	for soundID, player := range a.players {
		if !player.IsPlaying() {
			_ = player.Close()
			delete(a.players, soundID)
		}
	}

	return nil
}
