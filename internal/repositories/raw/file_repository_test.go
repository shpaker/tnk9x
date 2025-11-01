package raw

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestGetAssetPath(t *testing.T) {
	baseDir := "/tmp/assets"
	assetName := "base.json"

	repo := NewFileRepository(baseDir)
	result, err := repo.getPath(assetName)
	if err != nil {
		t.Errorf("неожиданная ошибка: %v", err)
		return
	}

	expected := fmt.Sprintf("%s/%s", baseDir, assetName)
	if result != expected {
		t.Errorf("ожидался путь %s, получен %s", expected, result)
	}
}

func TestReadFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "assets_test")
	if err != nil {
		t.Fatalf("не удалось создать временную папку: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	testData := []byte("test content")
	testFile := "test.txt"
	filePath := filepath.Join(tempDir, testFile)

	if err := os.WriteFile(filePath, testData, 0o644); err != nil {
		t.Fatalf("не удалось создать тестовый файл: %v", err)
	}

	repo := NewFileRepository(tempDir)
	result, err := repo.ReadFile(testFile)
	if err != nil {
		t.Errorf("неожиданная ошибка: %v", err)
		return
	}

	if string(result) != string(testData) {
		t.Errorf("ожидались данные '%s', получены '%s'", testData, result)
	}
}

func TestReadImage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "assets_test")
	if err != nil {
		t.Fatalf("не удалось создать временную папку: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	testFile := "test"
	filePath := filepath.Join(tempDir, testFile+".png")

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("не удалось создать тестовый файл: %v", err)
	}
	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		t.Fatalf("не удалось записать PNG: %v", err)
	}
	_ = file.Close()

	repo := NewFileRepository(tempDir)
	result, err := repo.ReadImage(testFile)
	if err != nil {
		t.Errorf("неожиданная ошибка: %v", err)
		return
	}

	if result == nil {
		t.Error("ожидалось изображение, получен nil")
	}
}

func TestReadConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "assets_test")
	if err != nil {
		t.Fatalf("не удалось создать временную папку: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	testFile := "config"
	filePath := filepath.Join(tempDir, testFile+".yml")
	testYAML := []byte("name: test\nvalue: 42\n")

	if err := os.WriteFile(filePath, testYAML, 0o644); err != nil {
		t.Fatalf("не удалось создать тестовый файл: %v", err)
	}

	type TestConfig struct {
		Name  string `yaml:"name"`
		Value int    `yaml:"value"`
	}

	repo := NewFileRepository(tempDir)
	config, err := ReadConfig[TestConfig](repo, testFile)
	if err != nil {
		t.Errorf("неожиданная ошибка: %v", err)
		return
	}

	if config.Name != "test" {
		t.Errorf("ожидалось name='test', получено '%s'", config.Name)
	}

	if config.Value != 42 {
		t.Errorf("ожидалось value=42, получено %d", config.Value)
	}
}
