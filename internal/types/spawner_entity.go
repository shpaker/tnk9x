package types

type SpawnerEntity struct {
	AnimationGetter IImageIdGetter
	position        Position
	tank            *TankEntity
	duration_ms     uint16
}

func NewSpawnerEntity(
	tank *TankEntity,
) *SpawnerEntity {
	return &SpawnerEntity{
		tank: tank,
	}
}

func (se *SpawnerEntity) GetTank() *TankEntity {
	return se.tank
}

func (se *SpawnerEntity) GetDurationMs() uint16 {
	return se.duration_ms
}
