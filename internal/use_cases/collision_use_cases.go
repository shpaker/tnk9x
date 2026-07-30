package use_cases

import (
	"math"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

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
			uc.bulletUseCases.SpawnImpact(bullet)
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

	// Пули врагов проходят сквозь друг друга; гасятся только пары
	// с участием пули игрока
	firstOwner := first.GetOwner()
	secondOwner := second.GetOwner()
	if firstOwner != nil && secondOwner != nil &&
		firstOwner.IsEnemy() && secondOwner.IsEnemy() {
		return false
	}

	return uc.entitiesCollisionService.CheckColliders(first, second)
}

// friendlyFireFreezeTicks — заморозка союзника после дружественного
// попадания (~3 секунды при 60 TPS)
const friendlyFireFreezeTicks = 180

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
			uc.bulletUseCases.SpawnImpact(bullet)
			_ = uc.bulletUseCases.RemoveBullet(bullet)
			if tank.IsActive() {
				if !tank.IsEnemy() {
					// Щит поглощает пулю без урона
					if tank.HasShield() {
						return
					}

					// Дружественный огонь: союзник временно замирает,
					// урона нет, как в оригинале
					if owner := bullet.GetOwner(); owner != nil &&
						!owner.IsEnemy() {
						tank.SetFrozenTicks(friendlyFireFreezeTicks)
						uc.tankActions.Stop(tank, true)
						return
					}

					// Попадание врага — гибель танка независимо от
					// уровня; жизни спишутся при респавне
					uc.soundUseCases.RequestSound(
						types.SoundIDExplosion,
						false,
					)
					_ = uc.tankLifecycleUseCases.Explode(tank)
				} else {
					// Первое попадание по мигающему танку роняет бонус
					// на поле, мигание прекращается
					if tank.GetWithBonus() {
						tank.SetWithBonus(false)
						if uc.bonusUseCases != nil {
							uc.bonusUseCases.SpawnBonusOnField()
						}
					}

					// Для врагов проверяем здоровье
					// Тяжёлый танк (уровень 3) требует несколько попаданий
					currentHitPoints := tank.GetHitPoints()

					// Если здоровье уже 1 или меньше, взрываем сразу
					if currentHitPoints <= 1 {
						// Запоминаем автора добивающего выстрела
						// для начисления очков
						if owner := bullet.GetOwner(); owner != nil &&
							!owner.IsEnemy() {
							tank.SetDestroyedBy(owner.GetRole())
						}
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
	mapBlocks := uc.mapUseCases.GetBlocks()
	for _, block := range mapBlocks {
		// вода не останавливает снаряды
		if block.Data != nil && block.Data.Name == types.Water {
			continue
		}

		if uc.bulletCollisionService.CheckBulletBlockCollision(
			bullet,
			block,
		) {

			if block.Data != nil {
				if block.Data.Name == types.Brick {
					// Выстрел срезает полосу шириной в клетку 16px;
					// усиленная пуля пробивает клетку насквозь
					uc.destroyBlockStrip(
						bullet,
						block,
						bullet.IsReinforced(),
					)
					uc.soundUseCases.RequestSound(types.SoundIDBrick, false)
				} else if block.Data.Name == types.Steel {
					if bullet.IsReinforced() {
						// Усиленные пули срезают сталь той же полосой
						uc.destroyBlockStrip(bullet, block, false)
					}
					uc.soundUseCases.RequestSound(types.SoundIDSteel, false)
				}
			}

			return true
		}
	}

	return false
}

// destroyBlockStrip удаляет блоки полосы шириной в полную клетку 16px
// перпендикулярно полёту пули; fullCell дополнительно снимает вторую
// половину клетки по оси полёта (сквозное пробитие)
func (uc *CollisionUseCases) destroyBlockStrip(
	bullet *types.BulletEntity,
	block *types.BlockEntity,
	fullCell bool,
) {
	if block.Data == nil {
		return
	}
	blockType := block.Data.Name

	for _, position := range stripPositions(bullet, block, fullCell) {
		if target := uc.blockOfTypeAt(position, blockType); target != nil {
			_ = uc.mapUseCases.RemoveBlock(target)
		}
	}
}

// stripPositions возвращает координаты 8px-тайлов, снимаемых выстрелом:
// пара тайлов клетки поперёк полёта, при fullCell — вся клетка 16x16
func stripPositions(
	bullet *types.BulletEntity,
	block *types.BlockEntity,
	fullCell bool,
) []types.Position {
	blockPosition := block.GetPosition()
	tileSize := float64(block.GetSize().Width)
	cellSize := tileSize * 2

	// Вторая координата тайла внутри клетки 16px по указанной оси
	cellNeighbor := func(coordinate float64) float64 {
		cellStart := math.Floor(coordinate/cellSize) * cellSize
		if coordinate == cellStart {
			return cellStart + tileSize
		}
		return cellStart
	}

	horizontalStrip := bullet.Direction == types.DirectionUp ||
		bullet.Direction == types.DirectionDown

	positions := []types.Position{blockPosition}
	if horizontalStrip {
		positions = append(positions, types.Position{
			X: cellNeighbor(blockPosition.X),
			Y: blockPosition.Y,
		})
	} else {
		positions = append(positions, types.Position{
			X: blockPosition.X,
			Y: cellNeighbor(blockPosition.Y),
		})
	}

	if !fullCell {
		return positions
	}

	// Вторая половина клетки по оси полёта
	depth := len(positions)
	for i := 0; i < depth; i++ {
		position := positions[i]
		if horizontalStrip {
			position.Y = cellNeighbor(position.Y)
		} else {
			position.X = cellNeighbor(position.X)
		}
		positions = append(positions, position)
	}

	return positions
}

// blockOfTypeAt находит блок указанного типа в точке карты
func (uc *CollisionUseCases) blockOfTypeAt(
	position types.Position,
	blockType types.BlockType,
) *types.BlockEntity {
	for _, block := range uc.mapUseCases.GetBlocks() {
		if block == nil || block.Data == nil ||
			block.Data.Name != blockType {
			continue
		}
		blockPosition := block.GetPosition()
		if blockPosition.X == position.X && blockPosition.Y == position.Y {
			return block
		}
	}
	return nil
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
