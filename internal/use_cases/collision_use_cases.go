package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Untyped on purpose: it must fit both float64 positions and int sizes.
const brickHitDepthPx = 4

// Hits closer than this fraction of the path are simultaneous:
// a bullet on a tile seam hits both tiles.
const simultaneousHit = 1e-9

var _ interfaces.ICollisionUseCases = (*CollisionUseCases)(nil)

type CollisionUseCases struct {
	bulletUseCases           interfaces.IBulletUseCases
	tankActions              interfaces.ITankActionsUseCases
	mapUseCases              interfaces.IMapUseCases
	hqUseCases               interfaces.IHQUseCases
	tankCommonUseCases       interfaces.ITankCommonUseCases
	tankLifecycleUseCases    interfaces.ITankLifecycleUseCases
	bonusUseCases            interfaces.IBonusUseCases
	soundUseCases            interfaces.ISoundUseCases
	visualEffectsUseCases    interfaces.IVisualEffectsUseCases
	tankCollisionService     interfaces.ITankCollisionService
	bulletCollisionService   interfaces.IBulletCollisionService
	spawnCollisionService    interfaces.ISpawnCollisionService
	bonusesRepository        interfaces.IBonusesRepository
	collisionWorldRepository interfaces.ICollisionWorldRepository
}

func NewCollisionUseCases(
	bulletUseCases interfaces.IBulletUseCases,
	tankActions interfaces.ITankActionsUseCases,
	mapUseCases interfaces.IMapUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	tankCollisionService interfaces.ITankCollisionService,
	bulletCollisionService interfaces.IBulletCollisionService,
	spawnCollisionService interfaces.ISpawnCollisionService,
	hqUseCases interfaces.IHQUseCases,
	bonusUseCases interfaces.IBonusUseCases,
	bonusesRepository interfaces.IBonusesRepository,
	collisionWorldRepository interfaces.ICollisionWorldRepository,
	soundUseCases interfaces.ISoundUseCases,
	visualEffectsUseCases interfaces.IVisualEffectsUseCases,
) *CollisionUseCases {
	return &CollisionUseCases{
		bulletUseCases:           bulletUseCases,
		tankActions:              tankActions,
		mapUseCases:              mapUseCases,
		hqUseCases:               hqUseCases,
		tankCommonUseCases:       tankCommonUseCases,
		tankLifecycleUseCases:    tankLifecycleUseCases,
		bonusUseCases:            bonusUseCases,
		soundUseCases:            soundUseCases,
		visualEffectsUseCases:    visualEffectsUseCases,
		tankCollisionService:     tankCollisionService,
		bulletCollisionService:   bulletCollisionService,
		spawnCollisionService:    spawnCollisionService,
		bonusesRepository:        bonusesRepository,
		collisionWorldRepository: collisionWorldRepository,
	}
}

// Bullets move first, then tanks. Each entity travels from its previous-tick
// position, so fast movers cannot skip obstacles.
func (uc *CollisionUseCases) UpdateCollisions() {
	allTanks := uc.tankCommonUseCases.GetAllTanks()
	uc.syncWorld(allTanks)
	uc.moveBullets()
	uc.moveTanks(allTanks)
}

func (uc *CollisionUseCases) IsSpawnerBlocked(
	position types.Position,
	size types.Size,
) bool {
	return uc.spawnCollisionService.IsSpawnerBlocked(
		position,
		size,
		uc.tankCommonUseCases.GetAllTanks(),
	)
}

func (uc *CollisionUseCases) syncWorld(allTanks []*types.TankEntity) {
	var static []types.IEntityCollider
	for _, block := range uc.mapUseCases.GetBlocks() {
		static = append(static, block)
	}
	for _, border := range uc.mapUseCases.GetBorders() {
		static = append(static, border)
	}
	if hq := uc.hqUseCases.GetHQ(); hq != nil {
		static = append(static, hq)
	}
	for _, bonus := range uc.bonusesRepository.GetAllBonuses() {
		if bonus != nil {
			static = append(static, bonus)
		}
	}

	var movers []types.IMover
	for _, tank := range allTanks {
		if tank != nil {
			movers = append(movers, tank)
		}
	}
	for _, bullet := range uc.bulletUseCases.GetBullets() {
		if bullet != nil {
			movers = append(movers, bullet)
		}
	}
	uc.collisionWorldRepository.Sync(static, movers)
}

func (uc *CollisionUseCases) moveBullets() {
	bullets := append(
		[]*types.BulletEntity(nil),
		uc.bulletUseCases.GetBullets()...)
	gone := make(map[*types.BulletEntity]bool)
	for _, bullet := range bullets {
		if bullet == nil || gone[bullet] {
			continue
		}
		_, collisions := uc.collisionWorldRepository.Move(
			bullet,
			bullet.Position,
			func(_, other types.IEntityCollider) types.CollisionResponse {
				return uc.bulletCollisionService.Response(bullet, other)
			},
		)
		hits := firstHits(collisions)
		if len(hits) == 0 {
			continue
		}
		bullet.Position = hits[0].Touch
		for _, removed := range uc.hitByBullet(bullet, hits) {
			gone[removed] = true
			_ = uc.bulletUseCases.RemoveBullet(removed)
			uc.collisionWorldRepository.Remove(removed)
		}
	}
}

func firstHits(collisions []types.CollisionEntity) []types.CollisionEntity {
	if len(collisions) == 0 {
		return nil
	}
	first := collisions[0].Time
	for _, c := range collisions[1:] {
		first = min(first, c.Time)
	}
	var hits []types.CollisionEntity
	for _, c := range collisions {
		if c.Time <= first+simultaneousHit {
			hits = append(hits, c)
		}
	}
	return hits
}

// Walls and the map edge stop a bullet before the HQ: a brick and the eagle
// hit at the same time do not blow up the HQ.
func (uc *CollisionUseCases) hitByBullet(
	bullet *types.BulletEntity,
	hits []types.CollisionEntity,
) []*types.BulletEntity {
	var (
		blocks []*types.BlockEntity
		border bool
		hq     *types.HQEntity
		tank   *types.TankEntity
		other  *types.BulletEntity
	)
	for _, hit := range hits {
		switch entity := hit.Other.(type) {
		case *types.BlockEntity:
			blocks = append(blocks, entity)
		case *types.BorderEntity:
			border = true
		case *types.HQEntity:
			hq = entity
		case *types.TankEntity:
			if tank == nil {
				tank = entity
			}
		case *types.BulletEntity:
			if other == nil {
				other = entity
			}
		}
	}

	switch {
	case len(blocks) > 0 || border:
		uc.hitBlocks(bullet, blocks)
	case hq != nil:
		if !hq.IsDestroyed() {
			_ = uc.hqUseCases.Explode(hq)
		}
	case tank != nil:
		uc.hitTank(tank, bullet)
	case other != nil:
		uc.requestBulletClash(bullet, other)
		return []*types.BulletEntity{bullet, other}
	}
	return []*types.BulletEntity{bullet}
}

func (uc *CollisionUseCases) hitBlocks(
	bullet *types.BulletEntity,
	blocks []*types.BlockEntity,
) {
	hitBrick, hitSteel := false, false
	for _, block := range blocks {
		switch block.Data.Name {
		case types.Brick:
			hitBrick = true
			if bullet.IsReinforced() {
				uc.requestDebris(block, block.Position, block.GetSize(), bullet)
				uc.removeBlock(block)
			} else {
				uc.shaveBrickBlock(bullet, block)
			}
		case types.Steel:
			hitSteel = true
			if bullet.IsReinforced() {
				uc.requestDebris(block, block.Position, block.GetSize(), bullet)
				uc.removeBlock(block)
			}
		}
	}

	// One sound per hit, not per damaged block.
	if hitBrick {
		uc.soundUseCases.RequestSound(types.SoundIDBrick, false)
		uc.requestBulletEffect(types.VisualEventBrickHit, bullet)
	}
	if hitSteel {
		uc.soundUseCases.RequestSound(types.SoundIDSteel, false)
		uc.requestBulletEffect(types.VisualEventSteelHit, bullet)
	}
}

func (uc *CollisionUseCases) removeBlock(block *types.BlockEntity) {
	_ = uc.mapUseCases.RemoveBlock(block)
	uc.collisionWorldRepository.Remove(block)
}

func (uc *CollisionUseCases) hitTank(
	tank *types.TankEntity,
	bullet *types.BulletEntity,
) {
	if !tank.IsActive() {
		return
	}
	if tank.HasShield() {
		uc.requestBulletEffect(types.VisualEventShieldHit, bullet)
		return
	}
	if tank.HasBoat() {
		tank.LoseBoat()
		uc.requestBulletEffect(types.VisualEventShieldHit, bullet)
		uc.soundUseCases.RequestSound(types.SoundIDSteel, false)
		return
	}
	if !tank.IsEnemy() {
		currentLevel := uint(0)
		if tank.GetSpecs() != nil {
			currentLevel = tank.GetSpecs().GetLevel()
		}

		if currentLevel > 0 {
			uc.tankCommonUseCases.LevelDown(tank)
			uc.requestBulletEffect(
				types.VisualEventPlayerHit,
				bullet,
			)
			uc.soundUseCases.RequestSound(
				types.SoundIDExplosion,
				false,
			)
		} else {
			// Lives are decremented on respawn in TryRespawnPlayersTanks().
			uc.soundUseCases.RequestSound(
				types.SoundIDExplosion,
				false,
			)
			_ = uc.tankLifecycleUseCases.Explode(tank)
		}
	} else {
		currentHitPoints := tank.GetHitPoints()

		if currentHitPoints <= 1 {
			uc.soundUseCases.RequestSound(
				types.SoundIDExplosion,
				false,
			)
			_ = uc.tankLifecycleUseCases.Explode(tank)
		} else {
			tank.DecrementHitPoints()
			uc.requestBulletEffect(
				types.VisualEventEnemyHit,
				bullet,
			)
			uc.soundUseCases.RequestSound(
				types.SoundIDExplosion,
				false,
			)
		}
	}
}

func (uc *CollisionUseCases) moveTanks(allTanks []*types.TankEntity) {
	for _, tank := range allTanks {
		if tank == nil || !tank.IsActive() {
			continue
		}
		// A boatless tank that left the water cannot re-enter it.
		if tank.IsBoatSinking() &&
			!uc.mapUseCases.IsWaterUnder(tank.Position, tank.Size) {
			tank.RemoveBoat()
		}

		pos, collisions := uc.collisionWorldRepository.Move(
			tank,
			tank.Position,
			func(_, other types.IEntityCollider) types.CollisionResponse {
				return uc.tankCollisionService.Response(tank, other)
			},
		)
		tank.Position = pos

		blocked, atBorder, hitByBullet, tookBonus := false, false, false, false
		for _, collision := range collisions {
			switch other := collision.Other.(type) {
			case *types.BorderEntity:
				atBorder = true
			case *types.BulletEntity:
				if hitByBullet || uc.bulletCollisionService.Response(other, tank) == types.CollisionIgnore {
					continue
				}
				hitByBullet = true
				_ = uc.bulletUseCases.RemoveBullet(other)
				uc.collisionWorldRepository.Remove(other)
				uc.hitTank(tank, other)
			case *types.BonusEntity:
				if tookBonus || !tank.IsActive() {
					continue
				}
				// Apply removes the bonus, so take only one per tick.
				tookBonus = true
				uc.bonusUseCases.Apply(other, tank)
				uc.collisionWorldRepository.Remove(other)
			default:
				if collision.Response != types.CollisionCross {
					blocked = true
				}
			}
		}
		switch {
		case blocked:
			uc.tankActions.Stop(tank, true)
		case atBorder && tank.IsDocking():
			// At the map edge docking is aborted: the node beyond it is unreachable.
			// A tank driving into the edge keeps going because the key is held.
			tank.Halt()
		}
	}
}

func (uc *CollisionUseCases) shaveBrickBlock(
	bullet *types.BulletEntity,
	block *types.BlockEntity,
) {
	size := block.GetSize()

	horizontal := bullet.Direction == types.DirectionLeft ||
		bullet.Direction == types.DirectionRight

	depth := size.Height
	if horizontal {
		depth = size.Width
	}
	if depth <= brickHitDepthPx {
		uc.requestDebris(block, block.Position, size, bullet)
		uc.removeBlock(block)
		return
	}

	// The shaved slab is at the face the bullet entered: Up hits the bottom,
	// Left the right, Right and Down the left and top.
	slab := block.Position
	slabSize := types.Size{Width: size.Width, Height: brickHitDepthPx}
	if horizontal {
		slabSize = types.Size{Width: brickHitDepthPx, Height: size.Height}
	}
	switch bullet.Direction {
	case types.DirectionUp:
		slab.Y += float64(size.Height - brickHitDepthPx)
	case types.DirectionLeft:
		slab.X += float64(size.Width - brickHitDepthPx)
	}
	uc.requestDebris(block, slab, slabSize, bullet)

	// Right hits the left face and Down the top: the remainder position
	// shifts together with its size.
	switch bullet.Direction {
	case types.DirectionRight:
		block.Position.X += brickHitDepthPx
	case types.DirectionDown:
		block.Position.Y += brickHitDepthPx
	}

	if horizontal {
		block.Size = types.Size{
			Width:  size.Width - brickHitDepthPx,
			Height: size.Height,
		}
	} else {
		block.Size = types.Size{
			Width:  size.Width,
			Height: size.Height - brickHitDepthPx,
		}
	}
	uc.collisionWorldRepository.Update(block)
}

func (uc *CollisionUseCases) requestBulletEffect(
	kind types.VisualEventKind,
	bullet *types.BulletEntity,
) {
	size := bullet.GetSize()
	uc.visualEffectsUseCases.RequestEffect(types.VisualEventEntity{
		Kind: kind,
		Position: types.Position{
			X: bullet.Position.X + float64(size.Width)/2,
			Y: bullet.Position.Y + float64(size.Height)/2,
		},
		Direction: bullet.Direction,
	})
}

func (uc *CollisionUseCases) requestBulletClash(
	first *types.BulletEntity,
	second *types.BulletEntity,
) {
	firstSize, secondSize := first.GetSize(), second.GetSize()
	uc.visualEffectsUseCases.RequestEffect(types.VisualEventEntity{
		Kind: types.VisualEventBulletClash,
		Position: types.Position{
			X: (first.Position.X + float64(firstSize.Width)/2 +
				second.Position.X + float64(secondSize.Width)/2) / 2,
			Y: (first.Position.Y + float64(firstSize.Height)/2 +
				second.Position.Y + float64(secondSize.Height)/2) / 2,
		},
		Direction: first.Direction,
	})
}

// Debris fly from the destroyed area itself, not from the hit point.
func (uc *CollisionUseCases) requestDebris(
	block *types.BlockEntity,
	position types.Position,
	size types.Size,
	bullet *types.BulletEntity,
) {
	uc.visualEffectsUseCases.RequestEffect(types.VisualEventEntity{
		Kind:      types.VisualEventBlockDebris,
		Position:  position,
		Size:      size,
		Direction: bullet.Direction,
		Block:     block,
	})
}
