package game

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

func slideAll(_, _ types.IEntityCollider) types.CollisionResponse {
	return types.CollisionSlide
}

func TestCollisionWorldRepository(t *testing.T) {
	repo := NewCollisionWorldRepository()
	wall := &types.BorderEntity{
		Position: types.Position{X: 20},
		Size:     types.Size{Width: 4, Height: 16},
	}
	bullet := types.NewBulletEntity(
		types.Position{X: 0},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		nil,
		types.DirectionRight,
		nil,
		nil,
	)

	// Новая движущаяся сущность встаёт на позицию начала тика
	bullet.Position.X = 100
	repo.Sync([]types.IEntityCollider{wall}, []types.IMover{bullet})
	pos, collisions := repo.Move(bullet, bullet.Position, slideAll)
	if pos.X != 16 || len(collisions) != 1 || collisions[0].Other != wall {
		t.Fatalf("пуля в %v, встречи %+v", pos, collisions)
	}

	// Пропавшая стена убирается из мира
	repo.Sync(nil, []types.IMover{bullet})
	if pos, collisions = repo.Move(bullet, types.Position{X: 100}, slideAll); pos.X != 100 ||
		len(collisions) != 0 {
		t.Errorf(
			"стена осталась в мире: пуля в %v, встречи %+v",
			pos,
			collisions,
		)
	}

	// Сущность не из мира встаёт на цель без коллизий
	repo.Remove(bullet)
	if pos, _ = repo.Move(bullet, types.Position{X: 7}, slideAll); pos.X != 7 {
		t.Errorf("пуля не из мира в %v", pos)
	}
}
