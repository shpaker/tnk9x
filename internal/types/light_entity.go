package types

import (
	"image/color"
	"math"
)

// MaxLights — предел источников света в кадре; совпадает с размером
// массивов источников в шейдере освещения
const MaxLights = 16

// MaxViewers — предел зрителей (танков игроков); совпадает с размером
// массива зрителей в шейдере освещения
const MaxViewers = 2

// Форма света и поля зрения; значения совпадают с константами
// шейдера освещения, чтобы видимость объектов и картинка не расходились
const (
	// LightConeSoftness — доля угла конуса, освещённая в полную силу
	LightConeSoftness = 0.35
	// LightConeNear — на этой дистанции фара набирает полную силу:
	// у самого танка она не слепит
	LightConeNear = 40.0
	// ViewerAuraRadius — вплотную танк замечает всё с любой стороны
	ViewerAuraRadius = 28.0
	// ViewerFieldCosOuter, ViewerFieldCosInner — край поля зрения:
	// передняя полусфера с мягкой границей от 80° до 100° от оси
	ViewerFieldCosOuter = -0.17
	ViewerFieldCosInner = 0.17
)

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

// IlluminationAt — освещённость точки без учёта препятствий:
// всенаправленный свет спадает квадратично, фара — линейно
// и набирает силу на первых LightConeNear пикселях
func (l LightEntity) IlluminationAt(point Position) float64 {
	dx, dy := point.X-l.Position.X, point.Y-l.Position.Y
	distance := math.Hypot(dx, dy)
	if distance >= l.Radius {
		return 0
	}
	falloff := 1 - distance/l.Radius
	if !l.IsCone() {
		return falloff * falloff * l.Intensity
	}
	spot := 1.0
	if distance >= 1 {
		cosTheta := (dx*l.Direction.X + dy*l.Direction.Y) / distance
		spot = Smoothstep(l.ConeCos, l.ConeInnerCos(), cosTheta)
	}
	return falloff * Smoothstep(0, LightConeNear, distance) * spot * l.Intensity
}

// ViewerEntity — зритель: танк игрока, его глаза и направление взгляда
type ViewerEntity struct {
	Position  Position // Центр танка
	Direction Position // Направление взгляда (фары), единичный вектор
}

// FieldAt — попадает ли точка в поле зрения без учёта препятствий:
// передняя полусфера и вплотную с любой стороны, от 0 до 1
func (v ViewerEntity) FieldAt(point Position) float64 {
	dx, dy := point.X-v.Position.X, point.Y-v.Position.Y
	distance := math.Hypot(dx, dy)
	if distance <= ViewerAuraRadius {
		return 1
	}
	cosTheta := (dx*v.Direction.X + dy*v.Direction.Y) / distance
	return Smoothstep(ViewerFieldCosOuter, ViewerFieldCosInner, cosTheta)
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
	Ripple       float64 // Волнистость: блик на поверхности мерцает рябью
}
