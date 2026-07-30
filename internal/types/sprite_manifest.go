package types

// SpriteManifest перечисляет спрайты и анимации, на которые ссылается
// код; используется для fail-fast проверки ассетов при старте
type SpriteManifest struct {
	Images     map[TilesetType][]string
	Animations map[TilesetType][]string
}

// Merge возвращает объединение двух манифестов
func (m SpriteManifest) Merge(other SpriteManifest) SpriteManifest {
	return SpriteManifest{
		Images:     mergeSpriteIDs(m.Images, other.Images),
		Animations: mergeSpriteIDs(m.Animations, other.Animations),
	}
}

func mergeSpriteIDs(
	first map[TilesetType][]string,
	second map[TilesetType][]string,
) map[TilesetType][]string {
	merged := make(map[TilesetType][]string, len(first)+len(second))
	for tilesetType, ids := range first {
		merged[tilesetType] = append(merged[tilesetType], ids...)
	}
	for tilesetType, ids := range second {
		merged[tilesetType] = append(merged[tilesetType], ids...)
	}
	return merged
}
