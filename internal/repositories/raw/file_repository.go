package raw

import (
	"bytes"
	"image"
	_ "image/png"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	imageExt  = ".png"
	configExt = ".yml"
)

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

func ReadConfig[T any](repo *FileRepository, name string) (T, error) {
	var config T
	data, err := repo.ReadFile(name + configExt)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}
