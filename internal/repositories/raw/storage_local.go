//go:build js

package raw

import (
	"errors"
	"syscall/js"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

var _ interfaces.IStorageRepository = (*StorageRepository)(nil)

// StorageRepository хранит пользовательские данные в localStorage
// браузера под префиксом приложения
type StorageRepository struct {
	prefix string
}

// NewStorageRepository создаёт хранилище с префиксом ключей prefix
func NewStorageRepository(prefix string) *StorageRepository {
	return &StorageRepository{
		prefix: prefix,
	}
}

// DefaultStorageDir — префикс ключей localStorage
func DefaultStorageDir(appName string) (string, error) {
	return appName, nil
}

func (sr *StorageRepository) Load(key string) (data []byte, err error) {
	storage, err := localStorage()
	if err != nil {
		return nil, err
	}
	defer recoverJSError(&err)

	value := storage.Call("getItem", sr.prefix+"."+key)
	if value.IsNull() || value.IsUndefined() {
		return nil, nil
	}
	return []byte(value.String()), nil
}

func (sr *StorageRepository) Save(key string, data []byte) (err error) {
	storage, err := localStorage()
	if err != nil {
		return err
	}
	defer recoverJSError(&err)

	storage.Call("setItem", sr.prefix+"."+key, string(data))
	return nil
}

// localStorage недоступен в приватных режимах и при запрете
// хранения данных сайтом — это ошибка, а не паника
func localStorage() (storage js.Value, err error) {
	defer recoverJSError(&err)

	storage = js.Global().Get("localStorage")
	if storage.IsNull() || storage.IsUndefined() {
		return js.Value{}, errors.New("localStorage is unavailable")
	}
	return storage, nil
}

func recoverJSError(err *error) {
	if recovered := recover(); recovered != nil {
		*err = errors.New("localStorage access failed")
	}
}
