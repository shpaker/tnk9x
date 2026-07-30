package use_cases

import (
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
					if bullet.IsReinforced() {
						// Усиленная пуля снимает целый тайл глубины —
						// вдвое больше обычной, но не клетку насквозь
						uc.destroyBlockStrip(bullet, block)
					} else {
						// Обычная пуля состругивает слой в полтайла
						// глубиной по всей ширине клетки 16px
						uc.shaveBlockStrip(bullet, block)
					}
					uc.soundUseCases.RequestSound(types.SoundIDBrick, false)
				} else if block.Data.Name == types.Steel {
					if bullet.IsReinforced() {
						// Усиленные пули срезают сталь полосой в тайл
						uc.destroyBlockStrip(bullet, block)
					}
					uc.soundUseCases.RequestSound(types.SoundIDSteel, false)
				}
			}

			return true
		}
	}

	return false
}

// destroyBlockStrip удаляет полосу тайлов шириной в полную клетку 16px
// и глубиной в один тайл перпендикулярно полёту пули
func (uc *CollisionUseCases) destroyBlockStrip(
	bullet *types.BulletEntity,
	block *types.BlockEntity,
) {
	if block.Data == nil {
		return
	}
	blockType := block.Data.Name

	for _, position := range stripPositions(bullet, block) {
		if target := uc.blockOfTypeAt(position, blockType); target != nil {
			_ = uc.mapUseCases.RemoveBlock(target)
		}
	}
}

// shaveBlockStrip состругивает со стороны попадания слой в полтайла
// глубиной по обоим тайлам клетки: минимальная единица разрушения —
// половина тайла, как в оригинале
func (uc *CollisionUseCases) shaveBlockStrip(
	bullet *types.BulletEntity,
	block *types.BlockEntity,
) {
	if block.Data == nil {
		return
	}
	blockType := block.Data.Name

	for _, position := range stripPositions(bullet, block) {
		if target := uc.blockOfTypeAt(position, blockType); target != nil {
			uc.shaveBlock(target, bullet.Direction)
		}
	}
}

// blockShaveDepth — глубина одного среза: половина тайла 8px
const blockShaveDepth = 4

// shaveBlock срезает половину тайла со стороны, в которую летит пуля;
// исчерпанный блок удаляется с карты
func (uc *CollisionUseCases) shaveBlock(
	block *types.BlockEntity,
	direction types.Direction,
) {
	shaveDepth := blockShaveDepth

	switch direction {
	case types.DirectionUp:
		// Пуля летит вверх — срезаем нижнюю часть
		block.Size.Height -= shaveDepth
	case types.DirectionDown:
		// Пуля летит вниз — срезаем верхнюю часть
		block.Position.Y += float64(shaveDepth)
		block.Size.Height -= shaveDepth
	case types.DirectionLeft:
		// Пуля летит влево — срезаем правую часть
		block.Size.Width -= shaveDepth
	case types.DirectionRight:
		// Пуля летит вправо — срезаем левую часть
		block.Position.X += float64(shaveDepth)
		block.Size.Width -= shaveDepth
	}

	if block.Size.Width <= 0 || block.Size.Height <= 0 {
		_ = uc.mapUseCases.RemoveBlock(block)
	}
}

// stripPositions возвращает исходные координаты 8px-тайлов полосы:
// тайл попадания и сосед поперёк полёта со стороны центра пули —
// зона поражения 16px центрируется на пуле, а не на сетке клеток.
// Расчёт идёт от позиции тайла на сетке (Data.Position), поэтому уже
// подрезанные блоки попадают в полосу
func stripPositions(
	bullet *types.BulletEntity,
	block *types.BlockEntity,
) []types.Position {
	blockPosition := block.GetPosition()
	if block.Data != nil {
		blockPosition = block.Data.Position
	}
	tileSize := float64(2 * blockShaveDepth)

	horizontalStrip := bullet.Direction == types.DirectionUp ||
		bullet.Direction == types.DirectionDown

	positions := []types.Position{blockPosition}
	if horizontalStrip {
		bulletCenter := bullet.Position.X + float64(bullet.Size.Width)/2
		tileCenter := blockPosition.X + tileSize/2
		neighborX := blockPosition.X - tileSize
		if bulletCenter >= tileCenter {
			neighborX = blockPosition.X + tileSize
		}
		positions = append(positions, types.Position{
			X: neighborX,
			Y: blockPosition.Y,
		})
	} else {
		bulletCenter := bullet.Position.Y + float64(bullet.Size.Height)/2
		tileCenter := blockPosition.Y + tileSize/2
		neighborY := blockPosition.Y - tileSize
		if bulletCenter >= tileCenter {
			neighborY = blockPosition.Y + tileSize
		}
		positions = append(positions, types.Position{
			X: blockPosition.X,
			Y: neighborY,
		})
	}

	return positions
}

// blockOfTypeAt находит блок указанного типа по исходной позиции тайла
// на сетке: подрезанный блок смещается, но его Data.Position неизменна
func (uc *CollisionUseCases) blockOfTypeAt(
	position types.Position,
	blockType types.BlockType,
) *types.BlockEntity {
	for _, block := range uc.mapUseCases.GetBlocks() {
		if block == nil || block.Data == nil ||
			block.Data.Name != blockType {
			continue
		}
		gridPosition := block.Data.Position
		if gridPosition.X == position.X && gridPosition.Y == position.Y {
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
