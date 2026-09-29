package types

// EnemyAIDecision — решение AI-скрипта для вражеского танка
type EnemyAIDecision struct {
	Direction Direction // Направление, в которое повернуть танк
	Move      bool      // Ехать после поворота (false — стоять на месте)
	Shoot     bool      // Выстрелить
}

// AITankInfo — сведения о танке, доступные AI-скрипту
type AITankInfo struct {
	ID         uint
	Position   Position
	Size       Size
	Direction  Direction
	Level      uint // Уровень танка (0-3): у врагов Basic, Fast, Power, Armor
	HitPoints  uint
	WithBonus  bool
	Reinforced bool // Пули танка ломают бетон
	Boat       bool // У танка лодка: он ходит по воде
}

// AIBonusInfo — сведения о бонусе на поле
type AIBonusInfo struct {
	Position Position
	Size     Size
	Type     BonusType
}

// AIBulletInfo — сведения о летящей пуле
type AIBulletInfo struct {
	Position  Position
	Size      Size
	Direction Direction
	FromEnemy bool // Пуля выпущена вражеским танком
}

// AIHQInfo — сведения о штабе
type AIHQInfo struct {
	Position Position
	Size     Size
	Intact   bool
}

// EnemyAIContext — снимок мира, который видит AI-скрипт за один вызов
type EnemyAIContext struct {
	Self        AITankInfo
	Players     []AITankInfo // Активные танки игроков
	Enemies     []AITankInfo // Активные союзники без самого танка
	Bullets     []AIBulletInfo
	Bonuses     []AIBonusInfo // Бонусы, за которыми может ехать враг
	HQ          AIHQInfo
	StageNumber uint
	Tick        int
	Grid        *NavGridEntity
}

// NavOptions — параметры поиска пути
type NavOptions struct {
	TankSize      int  // Размер танка в пикселях; шаг сетки пути — половина
	BrickCost     int  // Дополнительная стоимость шага сквозь кирпич
	SteelPassable bool // Бетон считается кирпичом (танк его пробивает)
	WaterPassable bool // Вода проходима (у танка лодка)
}

// PathStep — первый шаг найденного пути
type PathStep struct {
	Direction Direction
	Length    int // Число шагов до цели; 0 — танк уже у цели
}

// RayHitKind — что встретил луч линии огня
type RayHitKind string

const (
	RayHitEdge   RayHitKind = "edge"
	RayHitBrick  RayHitKind = "brick"
	RayHitSteel  RayHitKind = "steel"
	RayHitPlayer RayHitKind = "player"
	RayHitEnemy  RayHitKind = "enemy"
	RayHitHQ     RayHitKind = "hq"
)

// RayTarget — объект, который может перекрыть линию огня
type RayTarget struct {
	Kind     RayHitKind
	Position Position
	Size     Size
}

// RayHit — первое препятствие на линии огня и расстояние до него в пикселях
type RayHit struct {
	Kind     RayHitKind
	Distance int
}
