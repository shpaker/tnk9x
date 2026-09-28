//go:build !js

package raw

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

var _ interfaces.IStorageRepository = (*StorageRepository)(nil)

// StorageRepository хранит пользовательские данные файлами
// в каталоге приложения (десктоп)
type StorageRepository struct {
	dir string
}

// NewStorageRepository создаёт хранилище в каталоге dir
func NewStorageRepository(dir string) *StorageRepository {
	return &StorageRepository{
		dir: dir,
	}
}

// DefaultStorageDir — каталог сохранений в пользовательском конфиге ОС
func DefaultStorageDir(appName string) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName), nil
}

func (sr *StorageRepository) Load(key string) ([]byte, error) {
	data, err := os.ReadFile(sr.path(key))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return data, err
}

// Save пишет данные через временный файл, чтобы сбой записи
// не оставил сохранение повреждённым
func (sr *StorageRepository) Save(key string, data []byte) error {
	if err := os.MkdirAll(sr.dir, 0o755); err != nil {
		return err
	}
	tmp := sr.path(key) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, sr.path(key))
}

func (sr *StorageRepository) path(key string) string {
	return filepath.Join(sr.dir, key+".json")
}
