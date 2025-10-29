package types

// TileStaticEntity представляет статический тайл с изображением
type TileStaticEntity struct {
	ImageID string
}

// GetImageID возвращает ID изображения тайла (реализует IImageIDGetter)
func (tse *TileStaticEntity) GetImageID() (string, error) {
	return tse.ImageID, nil
}
