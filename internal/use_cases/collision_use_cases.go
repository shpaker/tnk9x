package use_cases

import (
	"github.com/shpaker/tnk25/internal/interfaces"
	"github.com/shpaker/tnk25/internal/types"
	"github.com/shpaker/tnk25/internal/types/session_entities"
)

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
	bonusUseCases            *BonusUseCases
	bonusesRepository        interfaces.IBonusesRepository
	soundUseCases            *SoundUseCases
	stageSession             *session_entities.StageSessionEntity
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
	hqUseCases interfaces.IHQUseCases,
	bonusUseCases *BonusUseCases,
	bonusesRepository interfaces.IBonusesRepository,
	soundUseCases *SoundUseCases,
	stageSession *session_entities.StageSessionEntity,
) *CollisionUseCases {
	uc := &CollisionUseCases{
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
		bonusUseCases:            bonusUseCases,
		bonusesRepository:        bonusesRepository,
		soundUseCases:            soundUseCases,
		stageSession:             stageSession,
	}

	return uc
}

func (uc *CollisionUseCases) UpdateCollisions() {
	allTanks := uc.tankCommonUseCases.GetAllTanks()

	bullets := uc.bulletUseCases.GetBullets()

	hq := uc.hqUseCases.GetHQ()

	mapBlocks := uc.mapUseCases.GetBlocks()

	bonuses := uc.getBonuses()

	uc.checkBulletsCollisions(bullets, hq, mapBlocks)

	bullets = uc.bulletUseCases.GetBullets()
	uc.checkTanksCollisions(allTanks, bullets, mapBlocks, bonuses)
}

func (uc *CollisionUseCases) checkTanksCollisions(
	allTanks []*types.TankEntity,
	bullets []*types.BulletEntity,
	mapBlocks types.MapBlocks,
	bonuses []*types.BonusEntity,
) {
	for _, tank := range allTanks {

		if !tank.IsActive() {
			continue
		}
		uc.checkTankBoundaryCollision(tank)
		uc.checkTankBlockCollisions(tank, mapBlocks)
		uc.checkTankTankCollision(tank, allTanks)
		uc.checkTankBulletCollisions(tank, bullets)
		uc.checkTankBonusCollisions(tank, bonuses)
	}
}

func (uc *CollisionUseCases) checkBulletsCollisions(
	bullets []*types.BulletEntity,
	hq *types.HQEntity,
	mapBlocks types.MapBlocks,
) {
	for index := 0; index < len(bullets); index++ {
		bullet := bullets[index]
		if bullet == nil {
			continue
		}
		if uc.checkBulletBoundaryCollision(bullet) ||
			uc.checkBulletHQCollision(bullet, hq) ||
			uc.checkBulletWallCollision(bullet, mapBlocks) {
			_ = uc.bulletUseCases.RemoveBullet(index)
			bullets = uc.bulletUseCases.GetBullets()
			index--
			continue
		}

		for j := index + 1; j < len(bullets); j++ {
			other := bullets[j]
			if uc.checkBulletBulletCollision(bullet, other) {
				_ = uc.bulletUseCases.RemoveBullet(j)
				_ = uc.bulletUseCases.RemoveBullet(index)
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
	bullets []*types.BulletEntity,
) {
	for index, bullet := range bullets {
		if bullet == nil {
			continue
		}
		if uc.bulletCollisionService.CheckBulletTankCollision(
			bullet,
			tank,
		) {
			_ = uc.bulletUseCases.RemoveBullet(index)
			if tank.IsActive() {
				// Для игроков понижаем уровень вместо взрыва
				if !tank.IsEnemy() {
					currentLevel := uint(0)
					if tank.GetSpecs() != nil {
						currentLevel = tank.GetSpecs().GetLevel()
					}

					if currentLevel > 0 {
						// Понижаем уровень
						if uc.tankCommonUseCases != nil {
							uc.tankCommonUseCases.LevelDown(tank)
						}
						if uc.soundUseCases != nil {
							uc.soundUseCases.RequestSound(
								types.SoundIDExplosion,
								false,
							)
						}
					} else {
						// Уровень уже минимальный - взрываем и уменьшаем жизни
						if uc.soundUseCases != nil {
							uc.soundUseCases.RequestSound(types.SoundIDExplosion, false)
						}
						_ = uc.tankLifecycleUseCases.Explode(tank)

						// Уменьшаем жизни игрока
						if uc.stageSession != nil {
							playerNum := types.RoleToPlayerTankNum(tank.GetRole())
							uc.stageSession.DecrementPlayerLives(playerNum)
						}
					}
				} else {
					// Для врагов взрываем сразу (старая логика)
					if uc.soundUseCases != nil {
						uc.soundUseCases.RequestSound(types.SoundIDExplosion, false)
					}
					_ = uc.tankLifecycleUseCases.Explode(tank)
				}
			}
			return
		}
	}
}

func (uc *CollisionUseCases) HasTankCollision(
	candidate *types.TankEntity,
) bool {
	if candidate == nil {
		return false
	}

	allTanks := uc.tankCommonUseCases.GetAllTanks()
	for _, otherTank := range allTanks {
		if otherTank == nil {
			continue
		}

		if uc.entitiesCollisionService.CheckColliders(candidate, otherTank) {
			return true
		}
	}

	return false
}

func (uc *CollisionUseCases) IsSpawnerBlocked(
	position types.Position,
	size types.Size,
) bool {
	if uc.entitiesCollisionService == nil ||
		uc.tankCommonUseCases == nil ||
		size.Width == 0 ||
		size.Height == 0 {
		return false
	}

	candidate := types.NewDefaultTankEntity(
		types.TankRoleEnemy,
		types.DirectionUp,
	)
	candidate.Size = size
	candidate.Position = types.Position{
		X: position.X * float64(size.Width),
		Y: position.Y * float64(size.Height),
	}

	for _, otherTank := range uc.tankCommonUseCases.GetAllTanks() {
		if otherTank == nil {
			continue
		}

		if uc.entitiesCollisionService.CheckColliders(&candidate, otherTank) {
			return true
		}
	}

	return false
}

func (uc *CollisionUseCases) checkTankBlockCollisions(
	tank *types.TankEntity,
	mapBlocks types.MapBlocks,
) {
	for _, block := range mapBlocks {
		if uc.wallCollisionService.CheckTankWallCollision(tank, block) {

			correctedPos, err := uc.entitiesCollisionService.ResolveCollisionPosition(
				tank,
				block,
				tank.Direction,
			)
			if err != nil {
				continue
			}

			tank.Position = correctedPos
			uc.tankActions.Stop(tank, true)
			return
		}
	}
}

func (uc *CollisionUseCases) checkTankTankCollision(
	tank *types.TankEntity,
	allTanks []*types.TankEntity,
) {
	for _, otherTank := range allTanks {

		if tank == otherTank || otherTank.IsDestroyed() {
			continue
		}

		if uc.entitiesCollisionService.CheckColliders(tank, otherTank) {

			correctedPos, err := uc.entitiesCollisionService.ResolveCollisionPosition(
				tank,
				otherTank,
				tank.Direction,
			)
			if err != nil {
				continue
			}

			tank.Position = correctedPos
			uc.tankActions.Stop(tank, true)
			uc.tankActions.Stop(otherTank, true)
			return
		}
	}
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
	mapBlocks types.MapBlocks,
) bool {
	for _, block := range mapBlocks {
		if uc.bulletCollisionService.CheckBulletBlockCollision(
			bullet,
			block,
		) {

			if block.Data != nil {
				if block.Data.Name == types.Brick && uc.soundUseCases != nil {
					_ = uc.mapUseCases.RemoveBlock(block)
					uc.soundUseCases.RequestSound(types.SoundIDBrick, false)
				} else if block.Data.Name == types.Steel {
					if bullet.IsReinforced() && uc.mapUseCases != nil {
						// Усиленные пули могут ломать стальные блоки
						_ = uc.mapUseCases.RemoveBlock(block)
						if uc.soundUseCases != nil {
							uc.soundUseCases.RequestSound(types.SoundIDSteel, false)
						}
					} else if uc.soundUseCases != nil {
						// Обычные пули только отскакивают от стальных блоков
						uc.soundUseCases.RequestSound(types.SoundIDSteel, false)
					}
				}
			}

			return true
		}
	}

	return false
}

func (uc *CollisionUseCases) checkBulletHQCollision(
	bullet *types.BulletEntity,
	hq *types.HQEntity,
) bool {
	if uc.bulletCollisionService.CheckBulletHQCollision(bullet, hq) {

		if uc.hqUseCases != nil && !hq.IsDestroyed() {
			_ = uc.hqUseCases.Explode(hq)
		}
		return true
	}
	return false
}

func (uc *CollisionUseCases) getBonuses() []*types.BonusEntity {
	if uc.bonusesRepository == nil {
		return nil
	}
	return uc.bonusesRepository.GetAllBonuses()
}

func (uc *CollisionUseCases) checkTankBonusCollisions(
	tank *types.TankEntity,
	bonuses []*types.BonusEntity,
) {
	if uc.bonusUseCases == nil || tank == nil {
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
