package raw

import (
	"bytes"
	"image"
	_ "image/png"
	"io/fs"
	"path"

	"gopkg.in/yaml.v3"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

const (
	imageExt  = ".png"
	configExt = ".yml"
)

var _ interfaces.IFileRepository = (*FileRepository)(nil)

type FileRepository struct {
	fsys fs.FS
}

// NewFileRepository создаёт репозиторий поверх любой fs.FS:
// os.DirFS для каталога на диске или embed.FS для встроенных ресурсов
func NewFileRepository(fsys fs.FS) *FileRepository {
	return &FileRepository{
		fsys: fsys,
	}
}

func (fr *FileRepository) ReadFile(name string) ([]byte, error) {
	data, err := fs.ReadFile(fr.fsys, name)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (fr *FileRepository) ReadImage(name string) (image.Image, error) {
	data, err := fr.ReadFile(name + imageExt)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return img, nil
}

func (fr *FileRepository) CountFiles(
	dirPath string,
	pattern string,
) (int, error) {
	entries, err := fs.ReadDir(fr.fsys, dirPath)
	if err != nil {
		return 0, err
	}

	count := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matched, err := path.Match(pattern, entry.Name())
		if err != nil {
			return 0, err
		}
		if matched {
			count++
		}
	}

	return count, nil
}

func ReadConfig[T any](repo *FileRepository, name string) (T, error) {
	var config T
	data, err := repo.ReadFile(name + configExt)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}
