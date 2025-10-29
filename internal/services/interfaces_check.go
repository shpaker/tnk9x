package services

import "github.com/shpaker/gonflict/internal/interfaces"

// Проверка реализации интерфейсов на этапе компиляции
var (
	_ interfaces.ITankBrakingService       = (*TankBrakingService)(nil)
	_ interfaces.ICoordinateService        = (*CoordinateService)(nil)
	_ interfaces.IBoundaryCollisionService = (*BoundaryCollisionService)(nil)
	_ interfaces.IWallCollisionService     = (*WallCollisionService)(nil)
	_ interfaces.IBulletCollisionService   = (*BulletCollisionService)(nil)
	_ interfaces.IImageService             = (*ImageService)(nil)
	_ interfaces.IAnimationService         = (*AnimationService)(nil)
	_ interfaces.ITileService              = (*TileService)(nil)
)
