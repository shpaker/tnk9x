//go:build js

package raw

import (
	"errors"
	"syscall/js"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

var _ interfaces.IStorageRepository = (*StorageRepository)(nil)

// bridgeName — глобальный мост площадки браузера; контракт описан
// в web/README.md
const bridgeName = "tnk9xPlatform"

// StorageRepository хранит пользовательские данные в хранилище
// площадки браузера (мост window.tnk9xPlatform.storage) под префиксом
// приложения: на обычной странице это localStorage, у портала — его
// облачные сохранения
type StorageRepository struct {
	prefix string
}

// NewStorageRepository создаёт хранилище с префиксом ключей prefix
func NewStorageRepository(prefix string) *StorageRepository {
	return &StorageRepository{
		prefix: prefix,
	}
}

// DefaultStorageDir — префикс ключей хранилища
func DefaultStorageDir(appName string) (string, error) {
	return appName, nil
}

func (sr *StorageRepository) Load(key string) (data []byte, err error) {
	storage, err := platformStorage()
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
	storage, err := platformStorage()
	if err != nil {
		return err
	}
	defer recoverJSError(&err)

	if !storage.Call("setItem", sr.prefix+"."+key, string(data)).Truthy() {
		return errors.New("platform storage rejected the write")
	}
	return nil
}

// platformStorage — хранилище моста; без моста страница собрана
// неверно — это ошибка, а не паника
func platformStorage() (storage js.Value, err error) {
	defer recoverJSError(&err)

	storage = js.Global().Get(bridgeName).Get("storage")
	if storage.IsNull() || storage.IsUndefined() {
		return js.Value{}, errors.New("platform storage is unavailable")
	}
	return storage, nil
}

func recoverJSError(err *error) {
	if recovered := recover(); recovered != nil {
		*err = errors.New("platform storage access failed")
	}
}
