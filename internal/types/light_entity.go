package types

import (
	"image/color"
	"math"
)

// MaxLights — предел источников света в кадре; совпадает с размером
// массивов источников и границей цикла в шейдере освещения
const MaxLights = 24

// LightConeSoftness — доля угла конуса, освещённая в полную силу
const LightConeSoftness = 0.35

// LightEntity — источник света в координатах игрового поля:
// всенаправленный или конус (фара)
type LightEntity struct {
	Position  Position // Центр источника
	Radius    float64  // Дальность освещения в пикселях
	Color     color.NRGBA
	Intensity float64  // Яркость в центре; больше 1 — пересвет для bloom
	Direction Position // Ось конуса, единичный вектор
	ConeCos   float64  // Косинус половины угла конуса; 0 — всенаправленный
}

// IsCone сообщает, светит ли источник конусом
func (l LightEntity) IsCone() bool {
	return l.ConeCos > 0
}

// ConeInnerCos — косинус края конуса, до которого свет в полную силу
func (l LightEntity) ConeInnerCos() float64 {
	return math.Cos(math.Acos(l.ConeCos) * LightConeSoftness)
}

// Smoothstep — плавный переход от 0 при edge0 к 1 при edge1,
// как в шейдерах; края можно передавать в обратном порядке
func Smoothstep(edge0, edge1, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-edge0)/(edge1-edge0)))
	return t * t * (3 - 2*t)
}

// SurfaceMaterial — свойства поверхности блока для освещения
type SurfaceMaterial struct {
	Opacity      float64 // 1 — непрозрачная стена, меньше — частичное затенение
	Reflectivity float64 // Сила блика от источников света
	Sparkle      float64 // Блик только на светлых пикселях спрайта
	// Приглушение рассеянного света источников: 0 — как у пола,
	// 1 — поверхность освещена только общим светом
	Dimming float64
}
