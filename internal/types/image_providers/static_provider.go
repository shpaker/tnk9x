package image_providers

type StaticProvider struct {
	ImageID string
}

func (sp *StaticProvider) GetImageID() (string, error) {
	return sp.ImageID, nil
}
