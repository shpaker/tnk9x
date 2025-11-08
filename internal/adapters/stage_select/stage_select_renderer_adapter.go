package stage_select

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// StageSelectRendererAdapter адаптер для рендеринга экрана выбора уровня
type StageSelectRendererAdapter struct {
	selector         *types.StageSelectorEntity
	selectorUseCases *use_cases.StageSelectorUseCases
	fontFace         text.Face
	screenWidth      int
	screenHeight     int
}

// NewStageSelectRendererAdapter создает новый экземпляр StageSelectRendererAdapter
func NewStageSelectRendererAdapter(
	selector *types.StageSelectorEntity,
	selectorUseCases *use_cases.StageSelectorUseCases,
	fontFace text.Face,
	screenWidth int,
	screenHeight int,
) (*StageSelectRendererAdapter, error) {
	if fontFace == nil {
		return nil, fmt.Errorf("font face is nil")
	}

	return &StageSelectRendererAdapter{
		selector:         selector,
		selectorUseCases: selectorUseCases,
		fontFace:         fontFace,
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
