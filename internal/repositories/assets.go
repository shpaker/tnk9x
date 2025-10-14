package repositories

import (
	"bytes"
	"image"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/shpaker/gonflict/internal/interfaces"
	"gopkg.in/yaml.v3"
)

const (
	imageExt  = ".png"
	configExt = ".yml"
)

type AssetsRepository struct {
	baseDir string
}

func NewAssetsService(baseDir string) *AssetsRepository {
	return &AssetsRepository{
		baseDir: baseDir,
	}
}

func (self *AssetsRepository) getAssetPath(name string) (string, error) {
	return filepath.Abs(filepath.Join(self.baseDir, name))
}

func (self *AssetsRepository) ReadAsset(name string) ([]byte, error) {
	path, error := self.getAssetPath(name)
	if error != nil {
		return nil, error
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (self *AssetsRepository) ReadImage(name string) (image.Image, error) {
	data, err := self.ReadAsset(name + imageExt)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return img, nil
}

func ReadConfig[T any](repo interfaces.IAssetsRepository, name string) (T, error) {
	var config T
	data, err := repo.ReadAsset(name + configExt)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}
