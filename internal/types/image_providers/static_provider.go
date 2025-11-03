package image_providers

// StaticProvider представляет статический провайдер изображений
type StaticProvider struct {
	ImageID string
}

// GetImageID возвращает ID изображения (реализует IImageProvider)
func (sp *StaticProvider) GetImageID() (string, error) {
	return sp.ImageID, nil
}
