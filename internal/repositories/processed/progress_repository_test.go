package processed

import (
	"errors"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

// memoryStorage — хранилище в памяти для тестов
type memoryStorage struct {
	data map[string][]byte
	err  error
}

func (s *memoryStorage) Load(key string) ([]byte, error) {
	return s.data[key], s.err
}

func (s *memoryStorage) Save(key string, data []byte) error {
	if s.err != nil {
		return s.err
	}
	s.data[key] = data
	return nil
}

func TestProgressRepository_RoundTrip(t *testing.T) {
	storage := &memoryStorage{data: map[string][]byte{}}
	repository := NewProgressRepository(storage, ProgressKeyOnePlayer)

	progress := types.NewProgressEntity()
	progress.SetStars(1, 3)
	progress.SetStars(2, 1)
	if err := repository.SaveProgress(progress); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := repository.GetProgress()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.GetStars(1) != 3 || loaded.GetStars(2) != 1 ||
		loaded.TotalStars() != 4 {
		t.Errorf("loaded stars %v", loaded.AllStars())
	}
}

func TestProgressRepository_Empty(t *testing.T) {
	repository := NewProgressRepository(
		&memoryStorage{data: map[string][]byte{}},
		ProgressKeyOnePlayer,
	)
	progress, err := repository.GetProgress()
	if err != nil || progress.TotalStars() != 0 {
		t.Errorf("empty storage: %v, stars %d", err, progress.TotalStars())
	}
}

// Повреждённое сохранение — ошибка, но игра получает пустой прогресс
func TestProgressRepository_Broken(t *testing.T) {
	tests := map[string][]byte{
		"malformed json":  []byte("{"),
		"unknown version": []byte(`{"version":99,"stars":{"1":3}}`),
	}
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			repository := NewProgressRepository(
				&memoryStorage{
					data: map[string][]byte{ProgressKeyOnePlayer: data},
				},
				ProgressKeyOnePlayer,
			)
			progress, err := repository.GetProgress()
			if err == nil {
				t.Error("expected an error")
			}
			if progress == nil || progress.TotalStars() != 0 {
				t.Error("expected an empty progress")
			}
		})
	}
}

// Значения вне диапазона обрезаются, мусорные ключи пропускаются
func TestProgressRepository_Sanitizes(t *testing.T) {
	data := []byte(`{"version":1,"stars":{"1":9,"x":2,"3":0}}`)
	repository := NewProgressRepository(
		&memoryStorage{data: map[string][]byte{ProgressKeyOnePlayer: data}},
		ProgressKeyOnePlayer,
	)
	progress, err := repository.GetProgress()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if progress.GetStars(1) != types.MaxLevelStars ||
		progress.IsCompleted(3) || progress.TotalStars() != 3 {
		t.Errorf("stars %v", progress.AllStars())
	}
}

func TestProgressRepository_StorageError(t *testing.T) {
	storage := &memoryStorage{
		data: map[string][]byte{},
		err:  errors.New("denied"),
	}
	repository := NewProgressRepository(storage, ProgressKeyOnePlayer)
	if _, err := repository.GetProgress(); err == nil {
		t.Error("load must report the storage error")
	}
	if err := repository.SaveProgress(types.NewProgressEntity()); err == nil {
		t.Error("save must report the storage error")
	}
}

// У одиночной игры и игры вдвоём раздельный прогресс в одном хранилище
func TestProgressRepository_SeparateModes(t *testing.T) {
	storage := &memoryStorage{data: map[string][]byte{}}
	single := NewProgressRepository(storage, ProgressKeyOnePlayer)
	duo := NewProgressRepository(storage, ProgressKeyTwoPlayers)

	progress := types.NewProgressEntity()
	progress.SetStars(1, 3)
	if err := duo.SaveProgress(progress); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := single.GetProgress()
	if err != nil || loaded.TotalStars() != 0 {
		t.Errorf(
			"прогресс 2P не должен попадать в 1P: %v, %d",
			err,
			loaded.TotalStars(),
		)
	}
	loaded, _ = duo.GetProgress()
	if loaded.GetStars(1) != 3 {
		t.Errorf("прогресс 2P: %v", loaded.AllStars())
	}
}
