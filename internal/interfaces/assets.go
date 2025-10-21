package interfaces

import "image"

// IAssetsRepository определяет интерфейс для работы с ассетами
type IAssetsRepository interface {
	// ReadAsset читает файл и возвращает его содержимое в виде байтов
	ReadAsset(name string) ([]byte, error)

	// ReadImage читает изображение (добавляет расширение .png автоматически)
	ReadImage(name string) (image.Image, error)
}
