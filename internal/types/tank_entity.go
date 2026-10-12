package types

import (
	"fmt"
	"math"

	"github.com/shpaker/koleya"
)

type PlayerTankNum int

const (
	PlayerTankNumPlayer1 PlayerTankNum = 0
	PlayerTankNumPlayer2 PlayerTankNum = 1
)

type TankRole string

const (
	TankRolePlayer1 TankRole = "player1"
	TankRolePlayer2 TankRole = "player2"
	TankRoleEnemy   TankRole = "enemy"
)

type TankState int

const (
	TankStateSpawning TankState = iota
	TankStateActive
	TankStateExploding
	TankStateExploded
)

type TankEntity struct {
	Position     Position
	PrevPosition Position
	Size         Size
	Altitude     Altitude
	Image        IImageProvider
	Direction    Direction
	State        TankState
	id           uint
	role         TankRole
	specs        *SpecsEntity
	withBonus    bool
	blinkCounter int
	blinkFlag    bool
	hitPoints    uint
	shieldTicks  uint
	freezeTicks  uint

	headlightAngle float64
	headlightReady bool
	recoilTicks    uint

	reload float64

	// boatSinking: the boat was shot down on water; the tank can still reach
	// the shore but cannot enter water again.
	boat        bool
	boatSinking bool

	motion koleya.Mover
	drive  koleya.Dir
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
		specs:     nil,
		motion:    koleya.NewMover(koleya.Vec2{}, direction.Dir(), TankLattice),
	}
}

func TankAnimationDirections() []string {
	return []string{"up", "left", "down", "right"}
}

func TankAnimationNameFor(
	role TankRole,
	model uint,
	direction string,
) string {
	return fmt.Sprintf("%s_level%d_tank_%s", role, model, direction)
}

func (t *TankEntity) AnimationName() string {
	if t == nil {
		return TankAnimationNameFor(TankRolePlayer1, 1, "up")
	}

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

func (t *TankEntity) GetPrevPosition() Position {
	return t.PrevPosition
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
	return t.State == TankStateActive && !t.motion.Moving()
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

func (t *TankEntity) GetBlinkProgress() float64 {
	if t == nil {
		return 0
	}
	return float64(t.blinkCounter) / BlinkPhaseTicks
}

func (t *TankEntity) UpdateBlink() {
	if t == nil {
		return
	}
	t.blinkCounter++
	if t.blinkCounter >= BlinkPhaseTicks {
		t.blinkCounter = 0
		t.blinkFlag = !t.blinkFlag
	}
}

func (t *TankEntity) ActivateShield(ticks uint) {
	if t == nil {
		return
	}
	t.shieldTicks = ticks
}

func (t *TankEntity) GetShieldTicks() uint {
	if t == nil {
		return 0
	}
	return t.shieldTicks
}

const ShieldFlickerTicks = 4

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

// Freeze stops the tank for ticks (enemy timer pickup). It lives on the tank,
// so a killed player respawns unfrozen.
func (t *TankEntity) Freeze(ticks uint) {
	if t == nil {
		return
	}
	t.freezeTicks = ticks
}

func (t *TankEntity) IsFrozen() bool {
	if t == nil {
		return false
	}
	return t.freezeTicks > 0
}

func (t *TankEntity) UpdateFreezeCountdown() {
	if t == nil {
		return
	}
	if t.freezeTicks > 0 {
		t.freezeTicks--
	}
}

func (t *TankEntity) GetHitPoints() uint {
	if t == nil {
		return 1
	}
	if t.hitPoints == 0 {
		return 1
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

var _ IBlink = (*TankEntity)(nil)

func (t *TankEntity) GetHeadlightAngle() float64 {
	if t == nil {
		return 0
	}
	if !t.headlightReady {
		return t.Direction.Angle()
	}
	return t.headlightAngle
}

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

const headlightSnapAngle = 0.01

func (t *TankEntity) StartRecoil(ticks uint) {
	if t == nil {
		return
	}
	t.recoilTicks = ticks
}

func (t *TankEntity) TickRecoil() {
	if t == nil || t.recoilTicks == 0 {
		return
	}
	t.recoilTicks--
}

func (t *TankEntity) IsRecoiling() bool {
	return t != nil && t.recoilTicks > 0
}

func (t *TankEntity) IsReloading() bool {
	return t != nil && t.reload > 0
}

func (t *TankEntity) StartReload(seconds float64) {
	if t == nil {
		return
	}
	t.reload = seconds
}

func (t *TankEntity) UpdateReload(dt float64) {
	if t == nil || t.reload <= 0 {
		return
	}
	t.reload = max(0, t.reload-dt)
}

func (t *TankEntity) HasBoat() bool {
	return t != nil && t.boat
}

func (t *TankEntity) SetBoat() {
	if t == nil {
		return
	}
	t.boat = true
	t.boatSinking = false
}

func (t *TankEntity) LoseBoat() {
	if t == nil || !t.boat {
		return
	}
	t.boat = false
	t.boatSinking = true
}

func (t *TankEntity) RemoveBoat() {
	if t == nil {
		return
	}
	t.boat = false
	t.boatSinking = false
}

func (t *TankEntity) IsBoatSinking() bool {
	return t != nil && t.boatSinking
}

func (t *TankEntity) CanSail() bool {
	return t != nil && (t.boat || t.boatSinking)
}
