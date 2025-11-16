package use_cases

import (
	"fmt"

	"golang.org/x/image/font/opentype"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

type FontUseCases struct {
	fontsRepository interfaces.IFontsRepository
	baseFont        *opentype.Font
}

func NewFontUseCases(
	fontsRepository interfaces.IFontsRepository,
) *FontUseCases {
	return &FontUseCases{
		fontsRepository: fontsRepository,
	}
}

func (uc *FontUseCases) GetFont() (*opentype.Font, error) {
	if uc == nil || uc.fontsRepository == nil {
		return nil, fmt.Errorf("fonts repository is not initialized")
	}

	if uc.baseFont != nil {
		return uc.baseFont, nil
	}

	fontData, err := uc.fontsRepository.GetFont("PressStart2P")
	if err != nil {
		return nil, fmt.Errorf("failed to load font data: %w", err)
	}

	parsedFont, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	uc.baseFont = parsedFont
	return uc.baseFont, nil
}
