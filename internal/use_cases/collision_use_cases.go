package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// brickHitDepthPx — глубина скола кирпича за одно попадание обычной
// пули: половина тайла (нетипизированная константа: годится и для
// float64-позиции, и для int-размера)
const brickHitDepthPx = 4

var _ interfaces.ICollisionUseCases = (*CollisionUseCases)(nil)

type CollisionUseCases struct {
	bulletUseCases           interfaces.IBulletUseCases
	tankActions              interfaces.ITankActionsUseCases
	mapUseCases              interfaces.IMapUseCases
	hqUseCases               interfaces.IHQUseCases
	tankCommonUseCases       interfaces.ITankCommonUseCases
	tankLifecycleUseCases    interfaces.ITankLifecycleUseCases
	boundaryCollisionService interfaces.IBoundaryCollisionService
	wallCollisionService     interfaces.IWallCollisionService
	bulletCollisionService   interfaces.IBulletCollisionService
	entitiesCollisionService interfaces.IEntitiesCollisionService
	spawnCollisionService    interfaces.ISpawnCollisionService
	bonusUseCases            interfaces.IBonusUseCases
	bonusesRepository        interfaces.IBonusesRepository
	soundUseCases            interfaces.ISoundUseCases
}

func NewCollisionUseCases(
	bulletUseCases interfaces.IBulletUseCases,
	tankActions interfaces.ITankActionsUseCases,
	mapUseCases interfaces.IMapUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases,
	boundaryCollisionService interfaces.IBoundaryCollisionService,
	wallCollisionService interfaces.IWallCollisionService,
	bulletCollisionService interfaces.IBulletCollisionService,
	entitiesCollisionService interfaces.IEntitiesCollisionService,
	spawnCollisionService interfaces.ISpawnCollisionService,
	hqUseCases interfaces.IHQUseCases,
	bonusUseCases interfaces.IBonusUseCases,
	bonusesRepository interfaces.IBonusesRepository,
	soundUseCases interfaces.ISoundUseCases,
) *CollisionUseCases {
	return &CollisionUseCases{
		bulletUseCases:           bulletUseCases,
		tankActions:              tankActions,
		mapUseCases:              mapUseCases,
		hqUseCases:               hqUseCases,
		tankCommonUseCases:       tankCommonUseCases,
		tankLifecycleUseCases:    tankLifecycleUseCases,
		boundaryCollisionService: boundaryCollisionService,
		wallCollisionService:     wallCollisionService,
		bulletCollisionService:   bulletCollisionService,
		entitiesCollisionService: entitiesCollisionService,
		spawnCollisionService:    spawnCollisionService,
		bonusUseCases:            bonusUseCases,
		bonusesRepository:        bonusesRepository,
		soundUseCases:            soundUseCases,
	}
}

func (uc *CollisionUseCases) UpdateCollisions() {
	allTanks := uc.tankCommonUseCases.GetAllTanks()

	bullets := uc.bulletUseCases.GetBullets()

	hq := uc.hqUseCases.GetHQ()

	bonuses := uc.getBonuses()

	uc.checkBulletsCollisions(bullets, hq)

	uc.checkTanksCollisions(allTanks, bonuses)
}

func (uc *CollisionUseCases) checkTanksCollisions(
	allTanks []*types.TankEntity,
	bonuses []*types.BonusEntity,
) {
	for _, tank := range allTanks {

		if !tank.IsActive() {
			continue
		}
		uc.checkTankBlockCollisions(tank)
		uc.checkTankTankCollision(tank, allTanks)
		uc.checkTankBoundaryCollision(tank)
		uc.checkTankBulletCollisions(tank)
		uc.checkTankBonusCollisions(tank, bonuses)
	}
}

func (uc *CollisionUseCases) checkBulletsCollisions(
	bullets []*types.BulletEntity,
	hq *types.HQEntity,
) {
	for index := 0; index < len(bullets); index++ {
		bullet := bullets[index]
		if bullet == nil {
			continue
		}
		if uc.checkBulletBoundaryCollision(bullet) ||
			uc.checkBulletHQCollision(bullet, hq) ||
			uc.checkBulletWallCollision(bullet) {
			_ = uc.bulletUseCases.RemoveBullet(bullet)
			bullets = uc.bulletUseCases.GetBullets()
			index--
			continue
		}

		for j := index + 1; j < len(bullets); j++ {
			other := bullets[j]
			if uc.checkBulletBulletCollision(bullet, other) {
				_ = uc.bulletUseCases.RemoveBullet(other)
				_ = uc.bulletUseCases.RemoveBullet(bullet)
				bullets = uc.bulletUseCases.GetBullets()
				index--
				break
			}
		}
	}
}

func (uc *CollisionUseCases) checkBulletBulletCollision(
	first *types.BulletEntity,
	second *types.BulletEntity,
) bool {
	if first == nil || second == nil {
		return false
	}

	return uc.entitiesCollisionService.CheckColliders(first, second)
}

func (uc *CollisionUseCases) checkTankBulletCollisions(
	tank *types.TankEntity,
) {
	bullets := uc.bulletUseCases.GetBullets()
	for _, bullet := range bullets {
		if bullet == nil {
			continue
		}
		if uc.bulletCollisionService.CheckBulletTankCollision(
			bullet,
			tank,
		) {
			_ = uc.bulletUseCases.RemoveBullet(bullet)
			if tank.IsActive() {
				// Щит от каски поглощает попадание без урона
				if tank.HasShield() {
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
						// Танк ещё жив - воспроизводим звук попадания
						uc.soundUseCases.RequestSound(
							types.SoundIDExplosion,
							false,
						)
						// Можно добавить визуальный эффект (мигание) для тяжёлого танка
					}
				}
			}
			return
		}
	}
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

func (uc *CollisionUseCases) checkTankBlockCollisions(
	tank *types.TankEntity,
) {
	mapBlocks := uc.mapUseCases.GetBlocks()
	for _, block := range mapBlocks {
		if uc.wallCollisionService.CheckTankWallCollision(tank, block) {

			correctedPos, err := uc.entitiesCollisionService.ResolveCollisionPosition(
				tank,
				block,
				tank.Direction,
			)
			if err != nil {
				// Блок сбоку/сзади: перекрытие не от текущего движения,
				// оставляем возможность выехать из него
				continue
			}

			tank.Position = correctedPos
			uc.tankActions.Stop(tank, true)
		}
	}
}

func (uc *CollisionUseCases) checkTankTankCollision(
	tank *types.TankEntity,
	allTanks []*types.TankEntity,
) {
	for _, otherTank := range allTanks {

		if tank == otherTank || otherTank == nil || otherTank.IsDestroyed() {
			continue
		}

		if !uc.entitiesCollisionService.CheckColliders(tank, otherTank) {
			continue
		}

		// Откатываем только танк, который в этом тике ехал в сторону другого:
		// его движение «не произошло». Другой танк не трогаем — если он тоже
		// ехал навстречу, откатится в собственной итерации. Перекрытие, не
		// вызванное текущим движением (например, спавн поверх другого танка),
		// не откатывается, чтобы из него можно было выехать.
		if !uc.tankMovedTowards(tank, otherTank) {
			continue
		}

		switch tank.Direction {
		case types.DirectionUp, types.DirectionDown:
			tank.Position.Y = tank.PrevPosition.Y
		case types.DirectionLeft, types.DirectionRight:
			tank.Position.X = tank.PrevPosition.X
		}
		uc.tankActions.Stop(tank, true)
	}
}

// tankMovedTowards сообщает, сместился ли танк в текущем тике по своей оси
// движения в сторону другого танка
func (uc *CollisionUseCases) tankMovedTowards(
	tank *types.TankEntity,
	otherTank *types.TankEntity,
) bool {
	switch tank.Direction {
	case types.DirectionUp:
		return tank.Position.Y < tank.PrevPosition.Y &&
			otherTank.Position.Y < tank.Position.Y
	case types.DirectionDown:
		return tank.Position.Y > tank.PrevPosition.Y &&
			otherTank.Position.Y > tank.Position.Y
	case types.DirectionLeft:
		return tank.Position.X < tank.PrevPosition.X &&
			otherTank.Position.X < tank.Position.X
	case types.DirectionRight:
		return tank.Position.X > tank.PrevPosition.X &&
			otherTank.Position.X > tank.Position.X
	}
	return false
}

func (uc *CollisionUseCases) checkTankBoundaryCollision(
	tank *types.TankEntity,
) {
	if uc.boundaryCollisionService.CheckLeftBoundaryCollision(tank) {
		uc.tankActions.SetMinXPosition(tank)
	}
	if uc.boundaryCollisionService.CheckRightBoundaryCollision(tank) {
		uc.tankActions.SetMaxXPosition(tank)
	}
	if uc.boundaryCollisionService.CheckTopBoundaryCollision(tank) {
		uc.tankActions.SetMinYPosition(tank)
	}
	if uc.boundaryCollisionService.CheckBottomBoundaryCollision(tank) {
		uc.tankActions.SetMaxYPosition(tank)
	}
}

func (uc *CollisionUseCases) checkBulletBoundaryCollision(
	bullet *types.BulletEntity,
) bool {
	if uc.boundaryCollisionService.CheckLeftBoundaryCollision(bullet) ||
		uc.boundaryCollisionService.CheckRightBoundaryCollision(bullet) ||
		uc.boundaryCollisionService.CheckTopBoundaryCollision(bullet) ||
		uc.boundaryCollisionService.CheckBottomBoundaryCollision(bullet) {
		return true
	}
	return false
}

func (uc *CollisionUseCases) checkBulletWallCollision(
	bullet *types.BulletEntity,
) bool {
	// Пуля на стыке тайлов накрывает несколько блоков — сначала
	// собираем все пересечённые, потом меняем карту: RemoveBlock
	// уплотняет срез блоков, менять его во время обхода нельзя
	var hitBlocks []*types.BlockEntity
	for _, block := range uc.mapUseCases.GetBlocks() {
		// вода не останавливает снаряды
		if block.Data != nil && block.Data.Name == types.Water {
			continue
		}

		if uc.bulletCollisionService.CheckBulletBlockCollision(
			bullet,
			block,
		) {
			hitBlocks = append(hitBlocks, block)
		}
	}

	hitBrick, hitSteel := false, false
	for _, block := range hitBlocks {
		if block.Data == nil {
			continue
		}

		switch block.Data.Name {
		case types.Brick:
			hitBrick = true
			if bullet.IsReinforced() {
				// Усиленные пули сносят кирпич целиком
				_ = uc.mapUseCases.RemoveBlock(block)
			} else {
				uc.shaveBrickBlock(bullet, block)
			}
		case types.Steel:
			hitSteel = true
			if bullet.IsReinforced() {
				// Усиленные пули могут ломать стальные блоки
				_ = uc.mapUseCases.RemoveBlock(block)
			}
		}
	}

	// Один звук на попадание, а не на каждый задетый блок
	if hitBrick {
		uc.soundUseCases.RequestSound(types.SoundIDBrick, false)
	}
	if hitSteel {
		uc.soundUseCases.RequestSound(types.SoundIDSteel, false)
	}

	return len(hitBlocks) > 0
}

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
		_ = uc.mapUseCases.RemoveBlock(block)
		return
	}

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
}

func (uc *CollisionUseCases) checkBulletHQCollision(
	bullet *types.BulletEntity,
	hq *types.HQEntity,
) bool {
	if uc.bulletCollisionService.CheckBulletHQCollision(bullet, hq) {

		if !hq.IsDestroyed() {
			_ = uc.hqUseCases.Explode(hq)
		}
		return true
	}
	return false
}

func (uc *CollisionUseCases) getBonuses() []*types.BonusEntity {
	return uc.bonusesRepository.GetAllBonuses()
}

func (uc *CollisionUseCases) checkTankBonusCollisions(
	tank *types.TankEntity,
	bonuses []*types.BonusEntity,
) {
	if tank == nil {
		return
	}

	// Вражеские танки не могут подбирать бонусы
	if tank.IsEnemy() {
		return
	}

	for _, bonus := range bonuses {
		if bonus == nil {
			continue
		}

		if uc.entitiesCollisionService.CheckColliders(tank, bonus) {
			// Применяем бонус
			uc.bonusUseCases.Apply(bonus, tank)
			// Бонус будет удален внутри Apply для гранаты и танка
			return
		}
	}
}
