package types

import "fmt"

// PlayerTankNum представляет номер игрока в массиве
type PlayerTankNum int

const (
	PlayerTankNumPlayer1 PlayerTankNum = 0 // Игрок 1
	PlayerTankNumPlayer2 PlayerTankNum = 1 // Игрок 2
)

// TankRole представляет роль танка (игрок 1, игрок 2 или враг)
type TankRole string

const (
	TankRolePlayer1 TankRole = "player1" // Игрок 1
	TankRolePlayer2 TankRole = "player2" // Игрок 2
	TankRoleEnemy   TankRole = "enemy"   // Враг
)

// TankModelName представляет название модели танка
type TankModelName string

const (
	TankModelNameRegular TankModelName = "regular" // Обычный танк
)

// TankModel представляет модель танка
type TankModel struct {
	name TankModelName
}

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
	Size          Size
	Altitude      Altitude
	Image         IImageProvider
	Speed         float64
	Direction     Direction
	State         TankState
	NextDirection *Direction
	role          TankRole
	model         TankModel
}

// NewDefaultTankEntity создает TankEntity с базовыми значениями
func NewDefaultTankEntity(role TankRole, direction Direction) TankEntity {
	return TankEntity{
		Position: Position{},
		Size: Size{
			Width:  16,
			Height: 16,
		},
		Altitude:  SURFACE,
		Speed:     0,
		Direction: direction,
		State:     TankStateSpawning,
		role:      role,
		model: TankModel{
			name: TankModelNameRegular,
		},
	}
}

// IsEnemy возвращает true, если танк является врагом
func (t *TankEntity) IsEnemy() bool {
	if t == nil {
		return false
	}
	return t.role == TankRoleEnemy
}

// GetRole возвращает роль танка
func (t *TankEntity) GetRole() TankRole {
	if t == nil {
		return TankRolePlayer1
	}
	return t.role
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

func (t *TankEntity) IsDestroyed() bool {
	return t.State == TankStateExploding || t.State == TankStateExploded
}

// IsStopped возвращает true если танк остановлен (по скорости или по состоянию)
func (t *TankEntity) IsStopped() bool {
	return t.Speed == 0 || t.State == TankStateStopped
}

// GetTankAnimationName возвращает имя анимации танка в зависимости от направления и типа
func (t *TankEntity) GetTankAnimationName() string {
	if t == nil {
		return "player1_regular_tank_up"
	}

	modelName := string(t.model.name)
	if modelName == "" {
		modelName = "regular"
	}

	roleStr := string(t.role)
	if roleStr == "" {
		roleStr = "player1"
	}

	prefix := roleStr + "_" + modelName

	var direction string
	switch t.Direction {
	case DirectionUp:
		direction = "up"
	case DirectionDown:
		direction = "down"
	case DirectionLeft:
		direction = "left"
	case DirectionRight:
		direction = "right"
	default:
		direction = "up"
	}

	return fmt.Sprintf("%s_tank_%s", prefix, direction)
}

// PlayerTankNumToRole преобразует PlayerTankNum в TankRole
func PlayerTankNumToRole(num PlayerTankNum) TankRole {
	switch num {
	case PlayerTankNumPlayer1:
		return TankRolePlayer1
	case PlayerTankNumPlayer2:
		return TankRolePlayer2
	default:
		return TankRolePlayer1
	}
}

// RoleToPlayerTankNum преобразует TankRole в PlayerTankNum
func RoleToPlayerTankNum(role TankRole) PlayerTankNum {
	switch role {
	case TankRolePlayer1:
		return PlayerTankNumPlayer1
	case TankRolePlayer2:
		return PlayerTankNumPlayer2
	default:
		return PlayerTankNumPlayer1
	}
}
