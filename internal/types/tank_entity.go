package types

// TankState представляет состояние танка
type TankState int

const (
	TankStateSpawning  TankState = iota // Танк спавнится
	TankStateMoving                     // Танк движется
	TankStateStopped                    // Танк остановлен
	TankStateBraking                    // Танк тормозит (доезжает до кратного 4)
	TankStateExploding                  // Танк взрывается
	TankStateExploded                   // Танк взорвался
)

// TankEntity представляет танк (игрока или врага)
type TankEntity struct {
	Position      Position
	Speed         float64
	Direction     Direction
	State         TankState // Состояние танка
	SpawnedAt     float64   // Время спавна танка
	Altitude      Altitude
	NextDirection *Direction // Следующее направление (используется во время Stops)
	Size          Size       // Размер танка
}

// GetSize возвращает размер танка
func (t *TankEntity) GetSize() Size {
	return t.Size
}

// GetPosition возвращает позицию танка в мире
func (t *TankEntity) GetPosition() Position {
	return t.Position
}

// GetAltitude возвращает высоту танка
func (t *TankEntity) GetAltitude() Altitude {
	// Если танк взрывается, показываем выше всего
	if t.State == TankStateExploding {
		return AIR
	}
	return t.Altitude
}

// IsActive возвращает true если танк активен (не спавнится, не взрывается и не взорвался)
func (t *TankEntity) IsActive() bool {
	return t.State != TankStateSpawning &&
		t.State != TankStateExploding &&
		t.State != TankStateExploded
}
