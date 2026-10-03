package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/kinescope"
	"github.com/shpaker/kinescope/ebitengine"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Параметры освещения и телевизора
const (
	// lightingAmbient — освещённость поля без источников: всё поле
	// видно, но свет фар и выстрелов заметно ярче
	lightingAmbient = 0.4
	lightingHaze    = 0.16 // видимость света на чёрном полу
	// distortionPerPixel — срыв строк на пиксель тряски кадра: полный
	// срыв при тряске в 4 пикселя
	distortionPerPixel = 0.25
	// shakeSignal — сигнал телевизора о тряске кадра уровня
	shakeSignal = "shake"
)

// Surface — прямоугольник поверхности на экране и её материал
type Surface struct {
	Rect     image.Rectangle
	Material types.SurfaceMaterial
}

// EffectsRendererAdapter рисует графические эффекты: свет с тенями
// на логическом экране и ламповый телевизор (kinescope) в проходе
// итогового экрана; живёт всё время работы приложения
type EffectsRendererAdapter struct {
	// Свет
	lightingShader *ebiten.Shader

	// Буферы размера логического экрана, пересоздаются при его смене
	scene *ebiten.Image
	mask  *ebiten.Image

	// Телевизор: вид «Горизонта», сигнал тряски и его отрисовка
	tv       *kinescope.TV
	shake    *kinescope.Level
	renderer *ebitengine.Renderer
	// effectsWereEnabled — эффекты были включены в прошлом кадре
	effectsWereEnabled bool

	// Белый пиксель для заливки маски материалов
	pixel *ebiten.Image

	// Uniform-массивы источников, переиспользуются между кадрами
	lightsUniform      []float32
	lightColorsUniform []float32
	lightConesUniform  []float32
}

func NewEffectsRendererAdapter(
	shadersRepository interfaces.IShadersRepository,
) (*EffectsRendererAdapter, error) {
	lightingShader, err := loadShader(shadersRepository, "lighting")
	if err != nil {
		return nil, err
	}
	tv, shake, err := newTV()
	if err != nil {
		return nil, err
	}
	renderer, err := ebitengine.NewRenderer()
	if err != nil {
		return nil, err
	}
	// Шейдер телевизора собирается сразу: ошибка останавливает
	// запуск до открытия окна
	if err := renderer.Prepare(tv); err != nil {
		return nil, err
	}

	pixel := ebiten.NewImage(1, 1)
	pixel.Fill(color.White)

	return &EffectsRendererAdapter{
		lightingShader:     lightingShader,
		tv:                 tv,
		shake:              shake,
		renderer:           renderer,
		pixel:              pixel,
		lightsUniform:      make([]float32, types.MaxLights*4),
		lightColorsUniform: make([]float32, types.MaxLights*4),
		lightConesUniform:  make([]float32, types.MaxLights*4),
	}, nil
}

// newTV собирает телевизор «Горизонт»: тряска кадра уровня срывает
// строки
func newTV() (*kinescope.TV, *kinescope.Level, error) {
	setup := kinescope.Gorizont()
	setup.Sources = map[string]kinescope.Source{shakeSignal: kinescope.Signal{}}
	setup.Drives = []kinescope.Drive{
		{From: shakeSignal, To: kinescope.TearStrength, Weight: 1},
	}
	tv, err := kinescope.NewTV(setup)
	if err != nil {
		return nil, nil, err
	}
	shake, err := tv.Signal(shakeSignal)
	if err != nil {
		return nil, nil, err
	}
	return tv, shake, nil
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
// задают маску материалов, источники — свет и тени внутри field;
// shake сдвигает весь кадр при тряске
func (a *EffectsRendererAdapter) DrawLighting(
	screen *ebiten.Image,
	field image.Rectangle,
	surfaces []Surface,
	lights []types.LightEntity,
	shake image.Point,
) {
	a.shake.Set(float32(min(
		1,
		math.Hypot(float64(shake.X), float64(shake.Y))*distortionPerPixel,
	)))
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
		a.fillCone(i, light)
	}

	fieldRect := []float32{
		float32(field.Min.X),
		float32(field.Min.Y),
		float32(field.Max.X),
		float32(field.Max.Y),
	}

	size := a.scene.Bounds().Size()
	op := &ebiten.DrawRectShaderOptions{}
	op.GeoM.Translate(float64(shake.X), float64(shake.Y))
	op.Images[0] = a.scene
	op.Images[1] = a.mask
	op.Uniforms = map[string]any{
		"Lights":      a.lightsUniform,
		"LightColors": a.lightColorsUniform,
		"LightCones":  a.lightConesUniform,
		"LightCount":  count,
		"Ambient":     float32(lightingAmbient),
		"Haze":        float32(lightingHaze),
		"FieldRect":   fieldRect,
	}
	screen.DrawRectShader(size.X, size.Y, a.lightingShader, op)
}

// fillCone заполняет конус источника i: ось и косинусы внешнего
// и внутреннего края; всенаправленный источник конусом не ограничен
func (a *EffectsRendererAdapter) fillCone(i int, light types.LightEntity) {
	if !light.IsCone() {
		a.lightConesUniform[i*4] = 0
		a.lightConesUniform[i*4+1] = 0
		a.lightConesUniform[i*4+2] = -2
		a.lightConesUniform[i*4+3] = -1
		return
	}
	a.lightConesUniform[i*4] = float32(light.Direction.X)
	a.lightConesUniform[i*4+1] = float32(light.Direction.Y)
	a.lightConesUniform[i*4+2] = float32(light.ConeCos)
	a.lightConesUniform[i*4+3] = float32(light.ConeInnerCos())
}

// drawMask заливает маску материалов: R — непрозрачность,
// G — отражательная способность, B — блики только на светлых
// пикселях спрайта, A — доля рассеянного света источников.
// Каналы независимы, а не цвет с альфой: пол без поверхностей
// освещён полностью, поверхность копируется в маску без смешивания
func (a *EffectsRendererAdapter) drawMask(surfaces []Surface) {
	a.mask.Fill(color.Black)
	for _, surface := range surfaces {
		op := &ebiten.DrawImageOptions{Blend: ebiten.BlendCopy}
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
			float32(surface.Material.Sparkle),
			float32(1-surface.Material.Dimming),
		)
		a.mask.DrawImage(a.pixel, op)
	}
}

// Итоговый экран

// DrawFinal масштабирует логический экран целым множителем scale
// в позицию (x, y); с эффектами показывает его через телевизор
func (a *EffectsRendererAdapter) DrawFinal(
	screen ebiten.FinalScreen,
	offscreen *ebiten.Image,
	x int,
	y int,
	scale int,
	enabled bool,
) {
	wereEnabled := a.effectsWereEnabled
	a.effectsWereEnabled = enabled
	if enabled {
		// Включённые заново эффекты начинают без послесвечения:
		// старый кадр не проступает призраком
		if !wereEnabled {
			a.tv.Reset()
		}
		// Часы телевизора идут по итоговым кадрам с эффектами: помехи
		// работают и в меню, где проход освещения не идёт
		a.tv.Update(1 / float64(ebiten.TPS()))
		err := a.renderer.Draw(screen, offscreen, a.tv, x, y, scale)
		// Срыв строк гаснет после каждого итогового кадра: вне уровня
		// тряски нет
		a.shake.Set(0)
		if err == nil {
			return
		}
	}

	op := &ebiten.DrawImageOptions{} // Filter по умолчанию — Nearest
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(offscreen, op)
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
