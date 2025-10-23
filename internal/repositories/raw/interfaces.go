package raw

import "image"

// IFileRepository определяет интерфейс для работы с файлами
type IFileRepository interface {
	// ReadFile читает файл и возвращает его содержимое в виде байтов
	ReadFile(name string) ([]byte, error)

	// ReadImage читает изображение (добавляет расширение .png автоматически)
	ReadImage(name string) (image.Image, error)
}
