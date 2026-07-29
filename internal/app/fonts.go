package app

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/opentype"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

// buildTextFace загружает базовый шрифт и создаёт face титульного размера,
// общий для всех рендер-адаптеров
func buildTextFace(
	fontsRepository interfaces.IFontsRepository,
	titleFontSize uint,
) (text.Face, error) {
	if titleFontSize == 0 {
		titleFontSize = 32
	}

	fontData, err := fontsRepository.GetFont("PressStart2P")
	if err != nil {
		return nil, fmt.Errorf("failed to load font data: %w", err)
	}

	parsedFont, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	fontFace, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
		Size: float64(titleFontSize),
		DPI:  72,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %w", err)
	}

	return text.NewGoXFace(fontFace), nil
}
