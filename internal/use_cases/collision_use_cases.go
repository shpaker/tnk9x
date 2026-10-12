package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// brickHitDepthPx — глубина скола кирпича за одно попадание обычной
// пули: половина тайла (нетипизированная константа: годится и для
// float64-позиции, и для int-размера)
const brickHitDepthPx = 4

// simultaneousHit — встречи, разделённые меньшей долей пути, считаются
// одновременными: пуля на стыке тайлов бьёт оба
const simultaneousHit = 1e-9

var _ interfaces.ICollisionUseCases = (*CollisionUseCases)(nil)

// CollisionUseCases проводит пули и танки через мир коллизий и
// применяет последствия встреч: сколы, попадания, упоры, подбор бонусов
type CollisionUseCases struct {
	// Use Cases
	bulletUseCases        interfaces.IBulletUseCases
	tankActions           interfaces.ITankActionsUseCases
	mapUseCases           interfaces.IMapUseCases
	hqUseCases            interfaces.IHQUseCases
	tankCommonUseCases    interfaces.ITankCommonUseCases
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases
	bonusUseCases         interfaces.IBonusUseCases
	soundUseCases         interfaces.ISoundUseCases
	visualEffectsUseCases interfaces.IVisualEffectsUseCases
	// Services
	tankCollisionService   interfaces.ITankCollisionService
	bulletCollisionService interfaces.IBulletCollisionService
	spawnCollisionService  interfaces.ISpawnCollisionService
	// Repositories
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

// UpdateCollisions реализует ICollisionUseCases: сначала летят пули,
// затем едут танки. Каждая сущность идёт от места, где закончила
// прошлый тик, к текущей позиции и встречает всё на пути, как бы
// далеко ни ушла за тик
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

// syncWorld сверяет мир коллизий с уровнем
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

// Пули

// moveBullets проводит каждую пулю к её позиции: пуля гаснет о первое,
// что встретила
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

// firstHits — встречи, случившиеся раньше всех остальных
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

// hitByBullet применяет одновременные встречи пули и возвращает пули,
// которые гаснут. Стена и край карты гасят пулю раньше штаба: пуля,
// задевшая одновременно кирпич и орла, не взрывает штаб
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

// hitBlocks скалывает и сносит задетые пулей блоки; пуля на стыке
// тайлов задевает несколько
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
				// Усиленные пули сносят кирпич целиком
				uc.requestDebris(block, block.Position, block.GetSize(), bullet)
				uc.removeBlock(block)
			} else {
				uc.shaveBrickBlock(bullet, block)
			}
		case types.Steel:
			hitSteel = true
			if bullet.IsReinforced() {
				// Усиленные пули могут ломать стальные блоки
				uc.requestDebris(block, block.Position, block.GetSize(), bullet)
				uc.removeBlock(block)
			}
		}
	}

	// Один звук на попадание, а не на каждый задетый блок
	if hitBrick {
		uc.soundUseCases.RequestSound(types.SoundIDBrick, false)
		uc.requestBulletEffect(types.VisualEventBrickHit, bullet)
	}
	if hitSteel {
		uc.soundUseCases.RequestSound(types.SoundIDSteel, false)
		uc.requestBulletEffect(types.VisualEventSteelHit, bullet)
	}
}

// removeBlock убирает блок с карты и из мира коллизий
func (uc *CollisionUseCases) removeBlock(block *types.BlockEntity) {
	_ = uc.mapUseCases.RemoveBlock(block)
	uc.collisionWorldRepository.Remove(block)
}

// hitTank применяет попадание пули в танк
func (uc *CollisionUseCases) hitTank(
	tank *types.TankEntity,
	bullet *types.BulletEntity,
) {
	if !tank.IsActive() {
		return
	}
	// Щит от каски поглощает попадание без урона
	if tank.HasShield() {
		uc.requestBulletEffect(types.VisualEventShieldHit, bullet)
		return
	}
	// Лодка поглощает попадание: танк
	// теряет лодку, но не уровень и не жизнь
	if tank.HasBoat() {
		tank.LoseBoat()
		uc.requestBulletEffect(types.VisualEventShieldHit, bullet)
		uc.soundUseCases.RequestSound(types.SoundIDSteel, false)
		return
	}
	// Для игроков понижаем уровень вместо взрыва
	if !tank.IsEnemy() {
		currentLevel := uint(0)
		if tank.GetSpecs() != nil {
			currentLevel = tank.GetSpecs().GetLevel()
		}

		if currentLevel > 0 {
			// Понижаем уровень
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
			// Уровень уже минимальный - взрываем танк
			// Жизни будут уменьшены при респавне в TryRespawnPlayersTanks()
			uc.soundUseCases.RequestSound(
				types.SoundIDExplosion,
				false,
			)
			_ = uc.tankLifecycleUseCases.Explode(tank)
		}
	} else {
		// Для врагов проверяем здоровье
		// Тяжёлый танк (уровень 3) требует несколько попаданий
		currentHitPoints := tank.GetHitPoints()

		// Если здоровье уже 1 или меньше, взрываем сразу
		if currentHitPoints <= 1 {
			uc.soundUseCases.RequestSound(
				types.SoundIDExplosion,
				false,
			)
			_ = uc.tankLifecycleUseCases.Explode(tank)
		} else {
			// Уменьшаем здоровье
			tank.DecrementHitPoints()
			uc.requestBulletEffect(
				types.VisualEventEnemyHit,
				bullet,
			)
			// Танк ещё жив - воспроизводим звук попадания
			uc.soundUseCases.RequestSound(
				types.SoundIDExplosion,
				false,
			)
			// Можно добавить визуальный эффект (мигание) для тяжёлого танка
		}
	}
}

// Танки

// moveTanks проводит каждый активный танк к его позиции: стена, край
// карты, целый штаб и другие танки останавливают его вплотную, бонусы
// и пули встречаются на ходу
func (uc *CollisionUseCases) moveTanks(allTanks []*types.TankEntity) {
	for _, tank := range allTanks {
		if tank == nil || !tank.IsActive() {
			continue
		}
		// Танк без лодки выплыл на берег: в воду больше не заехать
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
				// Одно попадание за тик
				hitByBullet = true
				_ = uc.bulletUseCases.RemoveBullet(other)
				uc.collisionWorldRepository.Remove(other)
				uc.hitTank(tank, other)
			case *types.BonusEntity:
				if tookBonus || !tank.IsActive() {
					continue
				}
				// Один бонус за тик; Apply убирает бонус
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
			// У края карты докатывание прерывается: узел за краем
			// недостижим. Едущий в край танк продолжает ехать —
			// клавиша зажата
			tank.Halt()
		}
	}
}

// Блоки

// shaveBrickBlock срезает с кирпича слой brickHitDepthPx на всю
// ширину остатка от грани, в которую попала пуля; остаток не толще
// слоя уничтожается целиком
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

	// Срезанный слой — у грани, в которую вошла пуля: Up бьёт в нижнюю,
	// Left — в правую, Right и Down — в левую и верхнюю
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

	// Right бьёт в левую грань, Down — в верхнюю: у них вместе с
	// размером смещается и позиция остатка
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

// Эффекты

// requestBulletEffect запрашивает эффект попадания в точке пули
// с направлением её полёта
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

// requestBulletClash запрашивает эффект столкновения пуль
// в точке между их центрами
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

// requestDebris запрашивает обломки разрушенной области стены:
// они вылетают из самой области, а не из точки попадания
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
