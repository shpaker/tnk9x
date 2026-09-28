package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Параметры освещения и постобработки
const (
	// lightingAmbient — освещённость поля без источников: периферия
	// заметно темнее обзора фары, но не пропадает
	lightingAmbient = 0.33
	lightingHaze    = 0.16 // видимость света на чёрном полу
	bloomThreshold  = 0.65 // яркость, с которой начинается свечение
	bloomStrength   = 0.9
	scanlineDepth   = 0.3
	// scanlinesMinScale — меньший масштаб не вмещает сканлайн
	// в логический пиксель, линии превращаются в муар
	scanlinesMinScale = 3
	// Ламповый телевизор: доля яркости послесвечения за кадр, сила
	// апертурной маски, расхождение лучей R и B у края в логических
	// пикселях, сила зерна, сила мерцания и бегущей полосы
	phosphorDecay = 0.6
	maskStrength  = 0.3
	convergence   = 0.45
	grainStrength = 0.08
	noiseStrength = 0.025
	// distortionPerPixel — срыв строк на пиксель тряски кадра: полный
	// срыв при тряске в 4 пикселя
	distortionPerPixel = 0.25
	// visionFadeSteps — на сколько единиц 8-битного канала за кадр
	// гаснет память недавно увиденного: 1 — за 255 кадров, ~4 с
	visionFadeSteps = 1
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
	visionShader   *ebiten.Shader
	lightingShader *ebiten.Shader
	bloomShader    *ebiten.Shader
	phosphorShader *ebiten.Shader
	crtShader      *ebiten.Shader

	// Буферы размера логического экрана, пересоздаются при его смене
	scene       *ebiten.Image
	mask        *ebiten.Image
	bloomBuffer *ebiten.Image
	bloom       *ebiten.Image

	// Зрение с памятью увиденного: прошлый и текущий кадр, меняются
	// местами каждый кадр; сбрасываются на новом уровне
	visionPrevious *ebiten.Image
	visionCurrent  *ebiten.Image

	// Послесвечение люминофора: прошлый и текущий итоговый кадр,
	// меняются местами каждый кадр; сбрасываются при включении эффектов
	phosphorPrevious *ebiten.Image
	phosphorCurrent  *ebiten.Image
	// effectsWereEnabled — эффекты были включены в прошлом кадре
	effectsWereEnabled bool

	// distortion — срыв строк от тряски кадра уровня, от 0 до 1;
	// гаснет после каждого итогового кадра, вне уровня нулевой
	distortion float64

	// Белый пиксель для заливки маски материалов
	pixel *ebiten.Image

	// Uniform-массивы источников, переиспользуются между кадрами
	lightsUniform      []float32
	lightColorsUniform []float32
	lightConesUniform  []float32
	viewersUniform     []float32

	frames int
	// finalFrames — счётчик итоговых кадров для помех: CRT работает
	// и в меню, где проход освещения не идёт
	finalFrames int
}

func NewEffectsRendererAdapter(
	shadersRepository interfaces.IShadersRepository,
) (*EffectsRendererAdapter, error) {
	visionShader, err := loadShader(shadersRepository, "vision")
	if err != nil {
		return nil, err
	}
	lightingShader, err := loadShader(shadersRepository, "lighting")
	if err != nil {
		return nil, err
	}
	bloomShader, err := loadShader(shadersRepository, "bloom")
	if err != nil {
		return nil, err
	}
	phosphorShader, err := loadShader(shadersRepository, "phosphor")
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
		visionShader:       visionShader,
		lightingShader:     lightingShader,
		bloomShader:        bloomShader,
		phosphorShader:     phosphorShader,
		crtShader:          crtShader,
		pixel:              pixel,
		lightsUniform:      make([]float32, types.MaxLights*4),
		lightColorsUniform: make([]float32, types.MaxLights*4),
		lightConesUniform:  make([]float32, types.MaxLights*4),
		viewersUniform:     make([]float32, types.MaxViewers*4),
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
	a.visionPrevious = ensureImage(a.visionPrevious, size)
	a.visionCurrent = ensureImage(a.visionCurrent, size)
	a.scene.Clear()
	return a.scene
}

// ResetVisionMemory забывает увиденное — на новом уровне игрок
// начинает с нетронутой картой
func (a *EffectsRendererAdapter) ResetVisionMemory() {
	if a.visionPrevious != nil {
		a.visionPrevious.Clear()
	}
	if a.visionCurrent != nil {
		a.visionCurrent.Clear()
	}
}

// DrawLighting переносит сцену на экран с освещением: поверхности
// задают маску материалов, источники — свет и тени внутри field;
// viewers — танки игроков: свет виден только там, куда они смотрят;
// shake сдвигает весь кадр при тряске
func (a *EffectsRendererAdapter) DrawLighting(
	screen *ebiten.Image,
	field image.Rectangle,
	surfaces []Surface,
	lights []types.LightEntity,
	viewers []types.ViewerEntity,
	shake image.Point,
) {
	a.frames++
	a.distortion = min(
		1,
		math.Hypot(float64(shake.X), float64(shake.Y))*distortionPerPixel,
	)
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
	a.drawVision(fieldRect, viewers)

	size := a.scene.Bounds().Size()
	op := &ebiten.DrawRectShaderOptions{}
	op.GeoM.Translate(float64(shake.X), float64(shake.Y))
	op.Images[0] = a.scene
	op.Images[1] = a.mask
	op.Images[2] = a.visionCurrent
	op.Uniforms = map[string]any{
		"Lights":      a.lightsUniform,
		"LightColors": a.lightColorsUniform,
		"LightCones":  a.lightConesUniform,
		"LightCount":  count,
		"Ambient":     float32(lightingAmbient),
		"Haze":        float32(lightingHaze),
		"FieldRect":   fieldRect,
		"Time":        float32(a.frames),
	}
	screen.DrawRectShader(size.X, size.Y, a.lightingShader, op)
}

// drawVision считает, что игроки видят сейчас, и обновляет память
// увиденного: прошлый кадр — вход, текущий — результат
func (a *EffectsRendererAdapter) drawVision(
	fieldRect []float32,
	viewers []types.ViewerEntity,
) {
	viewerCount := min(len(viewers), types.MaxViewers)
	for i := range viewerCount {
		a.viewersUniform[i*4] = float32(viewers[i].Position.X)
		a.viewersUniform[i*4+1] = float32(viewers[i].Position.Y)
		a.viewersUniform[i*4+2] = float32(viewers[i].Direction.X)
		a.viewersUniform[i*4+3] = float32(viewers[i].Direction.Y)
	}

	a.visionPrevious, a.visionCurrent = a.visionCurrent, a.visionPrevious
	size := a.mask.Bounds().Size()
	op := &ebiten.DrawRectShaderOptions{Blend: ebiten.BlendCopy}
	op.Images[0] = a.visionPrevious
	op.Images[1] = a.mask
	op.Uniforms = map[string]any{
		"Viewers":     a.viewersUniform,
		"ViewerCount": viewerCount,
		"FieldRect":   fieldRect,
		"Fade":        float32(visionFadeSteps) / 0xff,
	}
	a.visionCurrent.DrawRectShader(size.X, size.Y, a.visionShader, op)
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
	wereEnabled := a.effectsWereEnabled
	a.effectsWereEnabled = enabled
	if !enabled {
		op := &ebiten.DrawImageOptions{} // Filter по умолчанию — Nearest
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(offscreen, op)
		return
	}

	a.finalFrames++
	size := offscreen.Bounds().Size()
	a.drawBloom(offscreen, size)
	a.drawPhosphor(offscreen, size, !wereEnabled)

	// Мелкий масштаб не вмещает сканлайн и триаду маски
	// в логический пиксель: без них нет муара
	depth, mask := float32(0), float32(0)
	if scale >= scanlinesMinScale {
		depth, mask = scanlineDepth, maskStrength
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = a.phosphorCurrent
	op.Images[1] = a.bloom
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(x), float64(y))
	op.Uniforms = map[string]any{
		"BloomStrength": float32(bloomStrength),
		"ScanlineDepth": depth,
		"MaskStrength":  mask,
		"Convergence":   float32(convergence),
		"Grain":         float32(grainStrength),
		"Noise":         float32(noiseStrength),
		"Distortion":    float32(a.distortion),
		"Time":          float32(a.finalFrames),
	}
	screen.DrawRectShader(size.X, size.Y, a.crtShader, op)
	a.distortion = 0
}

// drawPhosphor накладывает кадр на гаснущую историю прошлых кадров;
// reset стирает историю — старый кадр не проступает призраком
func (a *EffectsRendererAdapter) drawPhosphor(
	offscreen *ebiten.Image,
	size image.Point,
	reset bool,
) {
	a.phosphorPrevious = ensureImage(a.phosphorPrevious, size)
	a.phosphorCurrent = ensureImage(a.phosphorCurrent, size)
	a.phosphorPrevious, a.phosphorCurrent = a.phosphorCurrent, a.phosphorPrevious
	if reset {
		a.phosphorPrevious.Clear()
	}

	op := &ebiten.DrawRectShaderOptions{Blend: ebiten.BlendCopy}
	op.Images[0] = offscreen
	op.Images[1] = a.phosphorPrevious
	op.Uniforms = map[string]any{
		"Decay": float32(phosphorDecay),
	}
	a.phosphorCurrent.DrawRectShader(size.X, size.Y, a.phosphorShader, op)
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
