package types

// TileStaticEntity представляет статический тайл с изображением
type TileStaticEntity struct {
	ImageId string
}

// GetImageId возвращает ID изображения тайла
func (tse *TileStaticEntity) GetImageId() (string, error) {
	return tse.ImageId, nil
}
