package types

import (
	"fmt"
	"math"
)

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
	id            uint     // Порядковый номер танка на уровне (для памяти AI)
	role          TankRole
	specs         *SpecsEntity // Спецификации танка
	withBonus     bool
	blinkCounter  int  // Счетчик тиков для мигания
	blinkFlag     bool // Флаг видимости
	hitPoints     uint // Количество попаданий до уничтожения (для тяжёлых танков)
	shieldTicks   uint // Оставшиеся тики неуязвимости от каски

	// Графические эффекты: направление фары (радианы, плавно
	// доворачивается за стволом) и оставшиеся тики отдачи выстрела
	headlightAngle float64
	headlightReady bool
	recoilTicks    uint

	// Видимость танка для игроков от 0 до 1: плавно следует за зоной
	// зрения, в которой он находится; до первого расчёта — полная
	visibility      float64
	visibilityReady bool
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

func (t *TankEntity) GetID() uint {
	if t == nil {
		return 0
	}
	return t.id
}

func (t *TankEntity) SetID(id uint) {
	if t == nil {
		return
	}
	t.id = id
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

// GetShieldTicks возвращает оставшиеся тики неуязвимости
func (t *TankEntity) GetShieldTicks() uint {
	if t == nil {
		return 0
	}
	return t.shieldTicks
}

// ShieldFlickerTicks — сколько тиков держится одна фаза мерцания щита
const ShieldFlickerTicks = 4

// GetShieldPhase возвращает фазу мерцания силового поля, 0 или 1:
// по ней чередуются кадры поля и яркость его свечения
func (t *TankEntity) GetShieldPhase() int {
	return int(t.GetShieldTicks()/ShieldFlickerTicks) % 2
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

// Графические эффекты

// GetHeadlightAngle возвращает направление фары в радианах
// (atan2 в экранных координатах, ось Y вниз)
func (t *TankEntity) GetHeadlightAngle() float64 {
	if t == nil {
		return 0
	}
	if !t.headlightReady {
		return t.Direction.Angle()
	}
	return t.headlightAngle
}

// TurnHeadlight доворачивает фару к направлению ствола на долю factor
// оставшегося угла по кратчайшему пути; первый вызов ставит фару сразу
func (t *TankEntity) TurnHeadlight(factor float64) {
	if t == nil {
		return
	}
	target := t.Direction.Angle()
	if !t.headlightReady {
		t.headlightAngle = target
		t.headlightReady = true
		return
	}
	diff := math.Remainder(target-t.headlightAngle, 2*math.Pi)
	if math.Abs(diff) < headlightSnapAngle {
		t.headlightAngle = target
		return
	}
	t.headlightAngle = math.Remainder(t.headlightAngle+diff*factor, 2*math.Pi)
}

// headlightSnapAngle — остаток доворота, при котором фара встаёт точно
const headlightSnapAngle = 0.01

// StartRecoil запускает отдачу выстрела на ticks тиков
func (t *TankEntity) StartRecoil(ticks uint) {
	if t == nil {
		return
	}
	t.recoilTicks = ticks
}

// TickRecoil отсчитывает тик отдачи
func (t *TankEntity) TickRecoil() {
	if t == nil || t.recoilTicks == 0 {
		return
	}
	t.recoilTicks--
}

// IsRecoiling сообщает, откатывается ли танк после выстрела
func (t *TankEntity) IsRecoiling() bool {
	return t != nil && t.recoilTicks > 0
}

// GetVisibility возвращает видимость танка для игроков от 0 до 1
func (t *TankEntity) GetVisibility() float64 {
	if t == nil || !t.visibilityReady {
		return 1
	}
	return t.visibility
}

// FadeVisibility приближает видимость к target: появление — на долю
// rise оставшегося, исчезновение — на долю fall; первый вызов ставит сразу
func (t *TankEntity) FadeVisibility(target, rise, fall float64) {
	if t == nil {
		return
	}
	if !t.visibilityReady {
		t.visibility = target
		t.visibilityReady = true
		return
	}
	rate := fall
	if target > t.visibility {
		rate = rise
	}
	t.visibility += (target - t.visibility) * rate
}
