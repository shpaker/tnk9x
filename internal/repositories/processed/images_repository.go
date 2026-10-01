package processed

import (
	"fmt"
	"image"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

var _ interfaces.IImagesRepository = (*ImagesRepository)(nil)

// ImagesRepository — картинки экранов (логотипы) из images/<name>.png
type ImagesRepository struct {
	fileRepository interfaces.IFileRepository
}

func NewImagesRepository(
	fileRepository interfaces.IFileRepository,
) *ImagesRepository {
	return &ImagesRepository{
		fileRepository: fileRepository,
	}
}

func (ir *ImagesRepository) GetImage(name string) (image.Image, error) {
	img, err := ir.fileRepository.ReadImage("images/" + name)
	if err != nil {
		return nil, fmt.Errorf("failed to read image '%s': %w", name, err)
	}
	return img, nil
}
