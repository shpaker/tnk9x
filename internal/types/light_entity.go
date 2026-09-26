package types

import "image/color"

// MaxLights — предел источников света в кадре; совпадает с размером
// массивов источников в шейдере освещения
const MaxLights = 16

// LightEntity — точечный источник света в координатах игрового поля
type LightEntity struct {
	Position  Position // Центр источника
	Radius    float64  // Дальность освещения в пикселях
	Color     color.NRGBA
	Intensity float64 // Яркость в центре; больше 1 — пересвет для bloom
}

// SurfaceMaterial — свойства поверхности блока для освещения
type SurfaceMaterial struct {
	Opacity      float64 // 1 — непрозрачная стена, меньше — частичное затенение
	Reflectivity float64 // Сила блика от источников света
	Ripple       float64 // Волнистость: блик на поверхности мерцает рябью
}
