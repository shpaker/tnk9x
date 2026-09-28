package types

// VisualEventKind — вид игрового события, порождающего визуальный эффект
type VisualEventKind uint8

const (
	VisualEventShot          VisualEventKind = iota // Выстрел танка
	VisualEventBrickHit                             // Пуля попала в кирпич
	VisualEventSteelHit                             // Пуля попала в сталь
	VisualEventShieldHit                            // Щит поглотил попадание
	VisualEventPlayerHit                            // Игрок потерял уровень
	VisualEventEnemyHit                             // Бронированный враг пережил попадание
	VisualEventTankExplosion                        // Взрыв танка
	VisualEventHQExplosion                          // Взрыв штаба
	VisualEventBlockDebris                          // Кусок стены разрушен
	VisualEventBulletClash                          // Пули столкнулись в полёте
)

// VisualEventEntity — событие кадра для графических эффектов
// в координатах игрового поля
type VisualEventEntity struct {
	Kind      VisualEventKind
	Position  Position     // Точка эффекта; у обломков — угол разрушенной области
	Size      Size         // Разрушенная область стены; у точечных событий нулевая
	Direction Direction    // Направление выстрела или летевшей пули
	Tank      *TankEntity  // Танк-источник; nil, если не танк
	Block     *BlockEntity // Разрушенный блок: спрайт задаёт цвета обломков
}
