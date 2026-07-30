package processed

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

const wavesFileName = "levels/waves.yml"

// wavesEnemiesPerStage — размер волны каждого этапа, как в NES
const wavesEnemiesPerStage = 20

// waveTierLevels отображает имена типов врагов из waves.yml
// на уровни спецификаций
var waveTierLevels = map[string]uint{
	"basic": 0,
	"fast":  1,
	"power": 2,
	"armor": 3,
}

type wavesFileSchema struct {
	Stages map[int][]waveGroupSchema `yaml:"stages"`
}

type waveGroupSchema struct {
	Tier  string `yaml:"tier"`
	Count uint   `yaml:"count"`
}

var _ interfaces.IWavesRepository = (*WavesRepository)(nil)

// WavesRepository читает состав вражеских волн по этапам из
// assets/levels/waves.yml; файл разбирается один раз и кэшируется
type WavesRepository struct {
	fileRepository interfaces.IFileRepository

	stages map[int][]uint
}

func NewWavesRepository(
	fileRepository interfaces.IFileRepository,
) *WavesRepository {
	return &WavesRepository{fileRepository: fileRepository}
}

func (r *WavesRepository) load() error {
	if r.stages != nil {
		return nil
	}

	data, err := r.fileRepository.ReadFile(wavesFileName)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", wavesFileName, err)
	}

	var schema wavesFileSchema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse %s: %w", wavesFileName, err)
	}

	stages := make(map[int][]uint, len(schema.Stages))
	for stage, groups := range schema.Stages {
		tiers, err := flattenWaveGroups(stage, groups)
		if err != nil {
			return err
		}
		stages[stage] = tiers
	}

	r.stages = stages
	return nil
}

// flattenWaveGroups разворачивает группы (тип, количество) в
// упорядоченный список уровней врагов и проверяет размер волны
func flattenWaveGroups(
	stage int,
	groups []waveGroupSchema,
) ([]uint, error) {
	tiers := make([]uint, 0, wavesEnemiesPerStage)
	for _, group := range groups {
		level, known := waveTierLevels[group.Tier]
		if !known {
			return nil, fmt.Errorf(
				"stage %d: unknown enemy tier %q",
				stage,
				group.Tier,
			)
		}
		for i := uint(0); i < group.Count; i++ {
			tiers = append(tiers, level)
		}
	}

	if len(tiers) != wavesEnemiesPerStage {
		return nil, fmt.Errorf(
			"stage %d: wave has %d enemies, expected %d",
			stage,
			len(tiers),
			wavesEnemiesPerStage,
		)
	}

	return tiers, nil
}

func (r *WavesRepository) GetWave(stage int) (types.StageWave, error) {
	if err := r.load(); err != nil {
		return types.StageWave{}, err
	}

	tiers, exists := r.stages[stage]
	if !exists {
		return types.StageWave{}, fmt.Errorf(
			"wave for stage %d not found in %s",
			stage,
			wavesFileName,
		)
	}

	return types.StageWave{Tiers: tiers}, nil
}

// GetStages возвращает отсортированные номера этапов с волнами
func (r *WavesRepository) GetStages() ([]int, error) {
	if err := r.load(); err != nil {
		return nil, err
	}

	stages := make([]int, 0, len(r.stages))
	for stage := range r.stages {
		stages = append(stages, stage)
	}
	sort.Ints(stages)
	return stages, nil
}
