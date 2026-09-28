package types

// Уровни врагов в волнах: совпадают с уровнями спецификаций танков
const (
	EnemyLevelBasic uint = 0
	EnemyLevelFast  uint = 1
	EnemyLevelPower uint = 2
	EnemyLevelArmor uint = 3
)

// EnemyLevelsCount — число типов вражеских танков
const EnemyLevelsCount = 4

// WaveStartKind — условие старта волны относительно предыдущих
type WaveStartKind int

const (
	// WaveStartNow — сразу после выхода последнего танка прошлой волны
	WaveStartNow WaveStartKind = iota
	// WaveStartLeft — когда живых врагов прошлых волн не больше Left
	WaveStartLeft
	// WaveStartClear — когда все враги прошлых волн уничтожены
	WaveStartClear
)

// WaveStart — условие старта волны
type WaveStart struct {
	Kind WaveStartKind
	Left uint
}

// WaveTank — танк волны: уровень врага и признак носителя бонуса
type WaveTank struct {
	Level    uint
	HasBonus bool
}

// WaveSpec — волна врагов: танки в порядке выхода, пауза между
// спаунами в тиках и условие старта
type WaveSpec struct {
	Tanks      []WaveTank
	DelayTicks uint
	Start      WaveStart
}

// LevelDefaults — параметры уровня для файлов без секций
// и для незаданных ключей
type LevelDefaults struct {
	MaxActive      uint
	Time3StarTicks uint
	// DelayTicks — пауза между спаунами для волн без явного delay
	DelayTicks uint
	Waves      []WaveSpec
}

// LevelEntity — уровень: карта и сценарий врагов
type LevelEntity struct {
	number         int
	name           string
	maxActive      uint
	time3StarTicks uint
	waves          []WaveSpec
	// explicitBonuses — носители бонусов заданы в файле строчными
	// буквами; иначе действует классическая нумерация
	explicitBonuses bool
	mapEntity       *MapEntity
}

func NewLevelEntity(
	number int,
	name string,
	maxActive uint,
	time3StarTicks uint,
	waves []WaveSpec,
	explicitBonuses bool,
	mapEntity *MapEntity,
) *LevelEntity {
	return &LevelEntity{
		number:          number,
		name:            name,
		maxActive:       maxActive,
		time3StarTicks:  time3StarTicks,
		waves:           waves,
		explicitBonuses: explicitBonuses,
		mapEntity:       mapEntity,
	}
}

func (l *LevelEntity) GetNumber() int {
	return l.number
}

func (l *LevelEntity) GetName() string {
	return l.name
}

func (l *LevelEntity) GetMaxActive() uint {
	return l.maxActive
}

func (l *LevelEntity) GetTime3StarTicks() uint {
	return l.time3StarTicks
}

func (l *LevelEntity) GetWaves() []WaveSpec {
	return l.waves
}

func (l *LevelEntity) HasExplicitBonuses() bool {
	return l.explicitBonuses
}

func (l *LevelEntity) GetMap() *MapEntity {
	return l.mapEntity
}

// GetTotalEnemies — число врагов уровня: сумма танков всех волн
func (l *LevelEntity) GetTotalEnemies() uint {
	total := uint(0)
	for _, wave := range l.waves {
		total += uint(len(wave.Tanks))
	}
	return total
}

// GetEnemyCounts — число врагов каждого типа по всем волнам
func (l *LevelEntity) GetEnemyCounts() [EnemyLevelsCount]uint {
	var counts [EnemyLevelsCount]uint
	for _, wave := range l.waves {
		for _, tank := range wave.Tanks {
			if tank.Level < EnemyLevelsCount {
				counts[tank.Level]++
			}
		}
	}
	return counts
}
