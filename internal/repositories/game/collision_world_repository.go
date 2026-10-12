package game

import (
	"github.com/shpaker/obstenku"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// collisionCellSizePx — ячейка пространственного индекса: размер танка.
// Влияет на скорость запросов, не на их результат
const collisionCellSizePx = 16

var _ interfaces.ICollisionWorldRepository = (*CollisionWorldRepository)(nil)

// CollisionWorldRepository хранит прямоугольники сущностей уровня
// в мире obstenku. Движущиеся сущности стоят там, где закончили прошлый
// тик, и Move ведёт их оттуда без пролёта сквозь стены
type CollisionWorldRepository struct {
	world *obstenku.World[types.IEntityCollider]
	// Поколение Sync, в котором сущность видели последней
	seen       map[types.IEntityCollider]uint64
	generation uint64
	// Буфер встреч, переиспользуемый между вызовами Move
	collisions []obstenku.Collision[types.IEntityCollider]
}

func NewCollisionWorldRepository() *CollisionWorldRepository {
	return &CollisionWorldRepository{
		world: obstenku.NewWorld[types.IEntityCollider](collisionCellSizePx),
		seen:  make(map[types.IEntityCollider]uint64),
	}
}

// Sync реализует ICollisionWorldRepository
func (r *CollisionWorldRepository) Sync(
	static []types.IEntityCollider,
	movers []types.IMover,
) {
	r.generation++
	for _, entity := range static {
		r.place(entity, entity.GetPosition())
		r.seen[entity] = r.generation
	}
	for _, mover := range movers {
		if !r.world.Has(mover) {
			r.place(mover, mover.GetPrevPosition())
		}
		r.seen[mover] = r.generation
	}
	for entity, generation := range r.seen {
		if generation != r.generation {
			_ = r.world.Remove(entity)
			delete(r.seen, entity)
		}
	}
}

// Update реализует ICollisionWorldRepository
func (r *CollisionWorldRepository) Update(entity types.IEntityCollider) {
	if r.world.Has(entity) {
		r.place(entity, entity.GetPosition())
	}
}

// Remove реализует ICollisionWorldRepository
func (r *CollisionWorldRepository) Remove(entity types.IEntityCollider) {
	_ = r.world.Remove(entity)
	delete(r.seen, entity)
}

// Move реализует ICollisionWorldRepository. Сущность не из мира
// встаёт на goal без коллизий
func (r *CollisionWorldRepository) Move(
	entity types.IEntityCollider,
	goal types.Position,
	filter types.CollisionFilter,
) (types.Position, []types.CollisionEntity) {
	if !r.world.Has(entity) {
		r.place(entity, goal)
		return goal, nil
	}

	pos, collisions, err := r.world.Move(
		entity,
		obstenku.Vec2{X: goal.X, Y: goal.Y},
		func(mover, other types.IEntityCollider) obstenku.Response {
			return toResponse(filter(mover, other))
		},
		r.collisions[:0],
	)
	r.collisions = collisions
	if err != nil {
		return goal, nil
	}

	result := make([]types.CollisionEntity, len(collisions))
	for i, c := range collisions {
		result[i] = types.CollisionEntity{
			Other:      c.Other,
			Response:   fromResponse(c.Response),
			Time:       c.Time,
			Touch:      types.Position{X: c.Touch.X, Y: c.Touch.Y},
			Overlapped: c.Overlapped,
		}
	}
	return types.Position{X: pos.X, Y: pos.Y}, result
}

// place ставит сущность в мир на position без коллизий
func (r *CollisionWorldRepository) place(
	entity types.IEntityCollider,
	position types.Position,
) {
	rect := types.RectOf(position, entity.GetSize())
	if r.world.Update(entity, rect) != nil {
		_ = r.world.Add(entity, rect)
	}
}

func toResponse(response types.CollisionResponse) obstenku.Response {
	switch response {
	case types.CollisionTouch:
		return obstenku.Touch
	case types.CollisionSlide:
		return obstenku.Slide
	case types.CollisionCross:
		return obstenku.Cross
	}
	return obstenku.Ignore
}

func fromResponse(response obstenku.Response) types.CollisionResponse {
	switch response {
	case obstenku.Touch:
		return types.CollisionTouch
	case obstenku.Slide:
		return types.CollisionSlide
	case obstenku.Cross:
		return types.CollisionCross
	}
	return types.CollisionIgnore
}
