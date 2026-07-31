package types

import "fmt"

type PlayerTankNum int

const (
	PlayerTankNumPlayer1 PlayerTankNum = 0
	PlayerTankNumPlayer2 PlayerTankNum = 1
)

type TankRole string

const (
	TankRolePlayer1 TankRole = "player1" // Игрок 1
	TankRolePlayer2 TankRole = "player2" // Игрок 2
	TankRoleEnemy   TankRole = "enemy"   // Враг
)

type TankState int

const (
	TankStateSpawning TankState = iota
	TankStateMoving
	TankStateStopped
	TankStateBraking
	TankStateExploding
	TankStateExploded
)

type TankEntity struct {
	Position      Position
	PrevPosition  Position // Позиция до движения в текущем тике (для отката коллизий)
	Size          Size
	Altitude      Altitude
	Image         IImageProvider
	Direction     Direction
	State         TankState
	NextDirection *Direction
	SlideTarget   *float64 // Зафиксированная цель скольжения на льду (nil — обычное торможение)
	role          TankRole
	specs         *SpecsEntity // Спецификации танка
	withBonus     bool
	blinkCounter  int  // Счетчик тиков для мигания
	blinkFlag     bool // Флаг видимости
	hitPoints     uint // Количество попаданий до уничтожения (для тяжёлых танков)
	shieldTicks   uint // Оставшиеся тики неуязвимости от каски
}

func NewDefaultTankEntity(role TankRole, direction Direction) TankEntity {
	return TankEntity{
		Position: Position{},
		Size: Size{
			Width:  16,
			Height: 16,
		},
		Altitude:  SURFACE,
		Direction: direction,
		State:     TankStateSpawning,
		role:      role,
		specs:     nil, // Будет установлено при создании танка
	}
}

// TankAnimationDirections перечисляет имена направлений в
// идентификаторах анимаций танков
func TankAnimationDirections() []string {
	return []string{"up", "left", "down", "right"}
}

// TankAnimationNameFor возвращает идентификатор анимации танка для роли,
// номера модели (1-4) и имени направления — единственное место, где
// задан формат имени
func TankAnimationNameFor(
	role TankRole,
	model uint,
	direction string,
) string {
	return fmt.Sprintf("%s_level%d_tank_%s", role, model, direction)
}

// AnimationName возвращает имя анимации танка, производное от уровня
// спецификаций, роли и направления; для nil-танка — имя по умолчанию.
func (t *TankEntity) AnimationName() string {
	if t == nil {
		return TankAnimationNameFor(TankRolePlayer1, 1, "up")
	}

	// Получаем уровень танка из спецификаций
	tankLevel := uint(0)
	if t.GetSpecs() != nil {
		tankLevel = t.GetSpecs().GetLevel()
	}
	if tankLevel > 3 {
		tankLevel = 3
	}

	role := t.GetRole()
	if role == "" {
		role = TankRolePlayer1
	}

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

	return TankAnimationNameFor(role, tankLevel+1, direction)
}

func (t *TankEntity) GetSpecs() *SpecsEntity {
	if t == nil {
		return nil
	}
	return t.specs
}

func (t *TankEntity) SetSpecs(specs *SpecsEntity) {
	if t == nil {
		return
	}
	t.specs = specs
}

func (t *TankEntity) IsEnemy() bool {
	if t == nil {
		return false
	}
	return t.role == TankRoleEnemy
}

func (t *TankEntity) GetRole() TankRole {
	if t == nil {
		return TankRolePlayer1
	}
	return t.role
}

func (t *TankEntity) GetSize() Size {
	return t.Size
}

func (t *TankEntity) GetPosition() Position {
	return t.Position
}

func (t *TankEntity) GetAltitude() Altitude {
	if t.State == TankStateExploding {
		return AIR
	}
	return t.Altitude
}

func (t *TankEntity) IsActive() bool {
	return t.State != TankStateSpawning &&
		t.State != TankStateExploding &&
		t.State != TankStateExploded
}

func (t *TankEntity) IsDestroyed() bool {
	return t.State == TankStateExploding || t.State == TankStateExploded
}

func (t *TankEntity) IsStopped() bool {
	return t.State == TankStateStopped
}

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

func (t *TankEntity) GetWithBonus() bool {
	if t == nil {
		return false
	}
	return t.withBonus
}

func (t *TankEntity) SetWithBonus(withBonus bool) {
	if t == nil {
		return
	}
	t.withBonus = withBonus
}

func (t *TankEntity) GetBlinkFlag() bool {
	if t == nil {
		return false
	}
	return t.blinkFlag
}

func (t *TankEntity) UpdateBlink() {
	if t == nil {
		return
	}
	t.blinkCounter++
	if t.blinkCounter >= 10 {
		t.blinkCounter = 0
		t.blinkFlag = !t.blinkFlag
	}
}

// ActivateShield включает неуязвимость танка на заданное число тиков
func (t *TankEntity) ActivateShield(ticks uint) {
	if t == nil {
		return
	}
	t.shieldTicks = ticks
}

func (t *TankEntity) HasShield() bool {
	if t == nil {
		return false
	}
	return t.shieldTicks > 0
}

func (t *TankEntity) UpdateShieldCountdown() {
	if t == nil {
		return
	}
	if t.shieldTicks > 0 {
		t.shieldTicks--
	}
}

func (t *TankEntity) GetHitPoints() uint {
	if t == nil {
		return 1
	}
	if t.hitPoints == 0 {
		return 1 // По умолчанию 1 попадание
	}
	return t.hitPoints
}

func (t *TankEntity) SetHitPoints(hitPoints uint) {
	if t == nil {
		return
	}
	t.hitPoints = hitPoints
}

func (t *TankEntity) DecrementHitPoints() {
	if t == nil {
		return
	}
	if t.hitPoints > 0 {
		t.hitPoints--
	}
}

// Реализация интерфейса IBlink
var _ IBlink = (*TankEntity)(nil)
