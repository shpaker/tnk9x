package stage_select

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/opentype"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// StageSelectRendererAdapter адаптер для рендеринга экрана выбора уровня
type StageSelectRendererAdapter struct {
	selector         *types.StageSelectorEntity
	selectorUseCases *use_cases.StageSelectorUseCases
	fontsRepository  interfaces.IFontsRepository
	fontFace         text.Face
	screenWidth      int
	screenHeight     int
}

// NewStageSelectRendererAdapter создает новый экземпляр StageSelectRendererAdapter
// Загружает шрифт PressStart2P через репозиторий и инициализирует адаптер
func NewStageSelectRendererAdapter(
	selector *types.StageSelectorEntity,
	selectorUseCases *use_cases.StageSelectorUseCases,
	fontsRepository interfaces.IFontsRepository,
	screenWidth int,
	screenHeight int,
) (*StageSelectRendererAdapter, error) {
	// Загружаем шрифт PressStart2P через репозиторий
	fontData, err := fontsRepository.GetFont("PressStart2P")
	if err != nil {
		return nil, fmt.Errorf("failed to get font from repository: %w", err)
	}

	tt, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	fontFace, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size: 32,
		DPI:  72,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %w", err)
	}

	// Создаем text/v2 Face из font.Face
	textFace := text.NewGoXFace(fontFace)

	return &StageSelectRendererAdapter{
		selector:         selector,
		selectorUseCases: selectorUseCases,
		fontsRepository:  fontsRepository,
		fontFace:         textFace,
		screenWidth:      screenWidth,
		screenHeight:     screenHeight,
	}, nil
}

// DrawAll отрисовывает все элементы экрана выбора уровня
func (r *StageSelectRendererAdapter) DrawAll(screen *ebiten.Image) {
	if r.selector == nil || r.selectorUseCases == nil {
		return
	}

	// Получаем реальный размер экрана из screen
	screenBounds := screen.Bounds()
	actualWidth := float64(screenBounds.Dx())
	actualHeight := float64(screenBounds.Dy())

	// Очищаем экран черным фоном для видимости
	screen.Fill(color.Black)

	// Отрисовываем текст выбранного уровня
	stageText := r.selectorUseCases.String(r.selector)

	// Измеряем текст для центрирования используя text/v2 API
	textWidth, textHeight := text.Measure(stageText, r.fontFace, 0)

	// Центрируем текст по горизонтали и вертикали
	// В text/v2, y указывает на baseline, поэтому вычитаем половину высоты текста
	x := (actualWidth - textWidth) / 2
	y := actualHeight/2 - textHeight/2

	// Отрисовываем текст используя text/v2
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, stageText, r.fontFace, op)
}
