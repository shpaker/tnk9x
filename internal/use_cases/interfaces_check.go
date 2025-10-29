package use_cases

import "github.com/shpaker/gonflict/internal/interfaces"

// Проверка реализации интерфейсов на этапе компиляции
var (
	_ interfaces.ITankUseCasesRef   = (*TankUseCases)(nil)
	_ interfaces.IBulletUseCases    = (*BulletUseCases)(nil)
	_ interfaces.IMapUseCases       = (*MapUseCases)(nil)
	_ interfaces.ICollisionUseCases = (*CollisionUseCases)(nil)
	_ interfaces.ITilesUseCases     = (*TilesUseCases)(nil)
)
