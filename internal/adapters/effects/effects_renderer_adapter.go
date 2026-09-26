package effects

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Параметры освещения и постобработки
const (
	lightingAmbient = 0.55 // освещённость поля без источников
	lightingHaze    = 0.16 // видимость света на чёрном полу
	bloomThreshold  = 0.65 // яркость, с которой начинается свечение
	bloomStrength   = 0.9
	scanlineDepth   = 0.3
	// scanlinesMinScale — меньший масштаб не вмещает сканлайн
	// в логический пиксель, линии превращаются в муар
	scanlinesMinScale = 3
)

// Surface — прямоугольник поверхности на экране и её материал
type Surface struct {
	Rect     image.Rectangle
	Material types.SurfaceMaterial
}

// EffectsRendererAdapter рисует графические эффекты: свет с тенями
// на логическом экране, bloom и CRT в проходе итогового экрана;
// живёт всё время работы приложения
type EffectsRendererAdapter struct {
	// Шейдеры
	lightingShader *ebiten.Shader
	bloomShader    *ebiten.Shader
	crtShader      *ebiten.Shader

	// Буферы размера логического экрана, пересоздаются при его смене
	scene       *ebiten.Image
	mask        *ebiten.Image
	bloomBuffer *ebiten.Image
	bloom       *ebiten.Image

	// Белый пиксель для заливки маски материалов
	pixel *ebiten.Image

	// Uniform-массивы источников, переиспользуются между кадрами
	lightsUniform      []float32
	lightColorsUniform []float32

	frames int
}

func NewEffectsRendererAdapter(
	shadersRepository interfaces.IShadersRepository,
) (*EffectsRendererAdapter, error) {
	lightingShader, err := loadShader(shadersRepository, "lighting")
	if err != nil {
		return nil, err
	}
	bloomShader, err := loadShader(shadersRepository, "bloom")
	if err != nil {
		return nil, err
	}
	crtShader, err := loadShader(shadersRepository, "crt")
	if err != nil {
		return nil, err
	}

	pixel := ebiten.NewImage(1, 1)
	pixel.Fill(color.White)

	return &EffectsRendererAdapter{
		lightingShader:     lightingShader,
		bloomShader:        bloomShader,
		crtShader:          crtShader,
		pixel:              pixel,
		lightsUniform:      make([]float32, types.MaxLights*4),
		lightColorsUniform: make([]float32, types.MaxLights*4),
	}, nil
}

// loadShader компилирует шейдер из репозитория; ошибка компиляции
// останавливает запуск до открытия окна
func loadShader(
	shadersRepository interfaces.IShadersRepository,
	name string,
) (*ebiten.Shader, error) {
	source, err := shadersRepository.GetShader(name)
	if err != nil {
		return nil, err
	}
	shader, err := ebiten.NewShader(source)
	if err != nil {
		return nil, fmt.Errorf("failed to compile shader '%s': %w", name, err)
	}
	return shader, nil
}

// Освещение

// BeginScene возвращает очищенный буфер сцены размера экрана: в него
// рисуется поле перед проходом освещения DrawLighting
func (a *EffectsRendererAdapter) BeginScene(size image.Point) *ebiten.Image {
	a.scene = ensureImage(a.scene, size)
	a.mask = ensureImage(a.mask, size)
	a.scene.Clear()
	return a.scene
}

// DrawLighting переносит сцену на экран с освещением: поверхности
// задают маску материалов, источники — свет и тени внутри field
func (a *EffectsRendererAdapter) DrawLighting(
	screen *ebiten.Image,
	field image.Rectangle,
	surfaces []Surface,
	lights []types.LightEntity,
) {
	a.frames++
	a.drawMask(surfaces)

	count := min(len(lights), types.MaxLights)
	for i := range count {
		light := lights[i]
		a.lightsUniform[i*4] = float32(light.Position.X)
		a.lightsUniform[i*4+1] = float32(light.Position.Y)
		a.lightsUniform[i*4+2] = float32(light.Radius)
		a.lightsUniform[i*4+3] = float32(light.Intensity)
		a.lightColorsUniform[i*4] = float32(light.Color.R) / 0xff
		a.lightColorsUniform[i*4+1] = float32(light.Color.G) / 0xff
		a.lightColorsUniform[i*4+2] = float32(light.Color.B) / 0xff
		a.lightColorsUniform[i*4+3] = 1
	}

	size := a.scene.Bounds().Size()
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = a.scene
	op.Images[1] = a.mask
	op.Uniforms = map[string]any{
		"Lights":      a.lightsUniform,
		"LightColors": a.lightColorsUniform,
		"LightCount":  count,
		"Ambient":     float32(lightingAmbient),
		"Haze":        float32(lightingHaze),
		"FieldRect": []float32{
			float32(field.Min.X),
			float32(field.Min.Y),
			float32(field.Max.X),
			float32(field.Max.Y),
		},
		"Time": float32(a.frames),
	}
	screen.DrawRectShader(size.X, size.Y, a.lightingShader, op)
}

// drawMask заливает маску материалов: R — непрозрачность,
// G — отражательная способность, B — волнистость
func (a *EffectsRendererAdapter) drawMask(surfaces []Surface) {
	a.mask.Clear()
	for _, surface := range surfaces {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(
			float64(surface.Rect.Dx()),
			float64(surface.Rect.Dy()),
		)
		op.GeoM.Translate(
			float64(surface.Rect.Min.X),
			float64(surface.Rect.Min.Y),
		)
		op.ColorScale.Scale(
			float32(surface.Material.Opacity),
			float32(surface.Material.Reflectivity),
			float32(surface.Material.Ripple),
			1,
		)
		a.mask.DrawImage(a.pixel, op)
	}
}

// Итоговый экран

// DrawFinal масштабирует логический экран целым множителем scale
// в позицию (x, y); с эффектами добавляет bloom и CRT
func (a *EffectsRendererAdapter) DrawFinal(
	screen ebiten.FinalScreen,
	offscreen *ebiten.Image,
	x int,
	y int,
	scale int,
	enabled bool,
) {
	if !enabled {
		op := &ebiten.DrawImageOptions{} // Filter по умолчанию — Nearest
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(offscreen, op)
		return
	}

	size := offscreen.Bounds().Size()
	a.drawBloom(offscreen, size)

	depth := float32(0)
	if scale >= scanlinesMinScale {
		depth = scanlineDepth
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = offscreen
	op.Images[1] = a.bloom
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(x), float64(y))
	op.Uniforms = map[string]any{
		"BloomStrength": float32(bloomStrength),
		"ScanlineDepth": depth,
	}
	screen.DrawRectShader(size.X, size.Y, a.crtShader, op)
}

// drawBloom выделяет яркие участки кадра и размывает их
// двумя проходами: по горизонтали, затем по вертикали
func (a *EffectsRendererAdapter) drawBloom(
	offscreen *ebiten.Image,
	size image.Point,
) {
	a.bloomBuffer = ensureImage(a.bloomBuffer, size)
	a.bloom = ensureImage(a.bloom, size)

	horizontal := &ebiten.DrawRectShaderOptions{Blend: ebiten.BlendCopy}
	horizontal.Images[0] = offscreen
	horizontal.Uniforms = map[string]any{
		"Direction": []float32{1, 0},
		"Threshold": float32(bloomThreshold),
	}
	a.bloomBuffer.DrawRectShader(size.X, size.Y, a.bloomShader, horizontal)

	vertical := &ebiten.DrawRectShaderOptions{Blend: ebiten.BlendCopy}
	vertical.Images[0] = a.bloomBuffer
	vertical.Uniforms = map[string]any{
		"Direction": []float32{0, 1},
		"Threshold": float32(0),
	}
	a.bloom.DrawRectShader(size.X, size.Y, a.bloomShader, vertical)
}

// ensureImage возвращает изображение нужного размера, пересоздавая
// его только при смене размера
func ensureImage(img *ebiten.Image, size image.Point) *ebiten.Image {
	if img != nil && img.Bounds().Size() == size {
		return img
	}
	if img != nil {
		img.Deallocate()
	}
	return ebiten.NewImage(size.X, size.Y)
}
