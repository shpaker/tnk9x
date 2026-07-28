package raw

import (
	"bytes"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

const (
	imageExt  = ".png"
	configExt = ".yml"
)

var _ interfaces.IFileRepository = (*FileRepository)(nil)

type FileRepository struct {
	baseDir string
}

func NewFileRepository(baseDir string) *FileRepository {
	return &FileRepository{
		baseDir: baseDir,
	}
}

func (fr *FileRepository) getPath(name string) (string, error) {
	return filepath.Abs(filepath.Join(fr.baseDir, name))
}

func (fr *FileRepository) ReadFile(name string) ([]byte, error) {
	path, err := fr.getPath(name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
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
	fullPath, err := fr.getPath(dirPath)
	if err != nil {
		return 0, err
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return 0, err
	}

	count := 0

	patternExt := ""
	if strings.HasPrefix(pattern, "*") {
		patternExt = pattern[1:]
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if patternExt != "" {
			if strings.HasSuffix(entry.Name(), patternExt) {
				count++
			}
		} else {

			matched, err := filepath.Match(pattern, entry.Name())
			if err != nil {
				return 0, err
			}
			if matched {
				count++
			}
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
