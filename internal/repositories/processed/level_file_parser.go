package processed

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shpaker/tnk9x/internal/types"
)

// Секции файла уровня
const (
	levelSectionLevel = "level"
	levelSectionWaves = "waves"
	levelSectionMap   = "map"
)

// Границы лимита одновременно активных врагов
const (
	minLevelMaxActive = 1
	maxLevelMaxActive = 10
)

// ticksPerSecond — тиков игрового цикла в секунде
const ticksPerSecond = 60

// waveTankLevels — буквы танков волны в уровни врагов;
// строчная буква — носитель бонуса
var waveTankLevels = map[rune]uint{
	'B': types.EnemyLevelBasic,
	'F': types.EnemyLevelFast,
	'P': types.EnemyLevelPower,
	'A': types.EnemyLevelArmor,
}

// levelFile — разобранный файл уровня до построения карты
type levelFile struct {
	name            string
	maxActive       uint
	time3StarTicks  uint
	waves           []types.WaveSpec
	explicitBonuses bool
	mapLines        []string
}

// sourceLine — строка файла с номером для сообщений об ошибках
type sourceLine struct {
	number int
	text   string
}

// parseLevelFile разбирает файл уровня: секционный формат
// ([level], [waves], [map]) или легаси — только сетка карты
func parseLevelFile(
	data string,
	defaults types.LevelDefaults,
) (levelFile, error) {
	result := levelFile{
		maxActive:      defaults.MaxActive,
		time3StarTicks: defaults.Time3StarTicks,
	}

	lines := strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n")
	if !isSectionedLevel(lines) {
		result.waves = defaults.Waves
		result.mapLines = nonEmptyLines(lines)
		return result, nil
	}

	sections, err := splitSections(lines)
	if err != nil {
		return result, err
	}

	if err := applyLevelSection(
		&result, sections[levelSectionLevel],
	); err != nil {
		return result, err
	}

	wavesLines, hasWaves := sections[levelSectionWaves]
	if hasWaves {
		waves, explicit, err := parseWaves(wavesLines, defaults.DelayTicks)
		if err != nil {
			return result, err
		}
		result.waves = waves
		result.explicitBonuses = explicit
	} else {
		result.waves = defaults.Waves
	}

	for _, line := range sections[levelSectionMap] {
		result.mapLines = append(result.mapLines, line.text)
	}
	if len(result.mapLines) == 0 {
		return result, fmt.Errorf("section [map] is missing or empty")
	}

	return result, nil
}

// isSectionedLevel — в файле есть заголовок секции; строки легаси-сетки
// не начинаются с «[», а комментарий «#» неотличим от ряда кирпичей
func isSectionedLevel(lines []string) bool {
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "[") {
			return true
		}
	}
	return false
}

func nonEmptyLines(lines []string) []string {
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// splitSections раскладывает строки по секциям; в [map] символ #
// — кирпич, поэтому комментарии там не вырезаются
func splitSections(lines []string) (map[string][]sourceLine, error) {
	sections := make(map[string][]sourceLine)
	current := ""

	for index, raw := range lines {
		number := index + 1
		trimmed := strings.TrimSpace(raw)

		if strings.HasPrefix(trimmed, "[") {
			name, err := parseSectionHeader(trimmed, number)
			if err != nil {
				return nil, err
			}
			if _, exists := sections[name]; exists {
				return nil, fmt.Errorf(
					"line %d: duplicate section [%s]", number, name,
				)
			}
			sections[name] = nil
			current = name
			continue
		}

		if current != levelSectionMap {
			trimmed = stripComment(trimmed)
		}
		if trimmed == "" {
			continue
		}
		if current == "" {
			return nil, fmt.Errorf(
				"line %d: content before the first section", number,
			)
		}
		sections[current] = append(
			sections[current],
			sourceLine{number: number, text: trimmed},
		)
	}

	return sections, nil
}

func parseSectionHeader(line string, number int) (string, error) {
	if !strings.HasSuffix(line, "]") {
		return "", fmt.Errorf("line %d: malformed section header", number)
	}
	name := strings.TrimSpace(line[1 : len(line)-1])
	switch name {
	case levelSectionLevel, levelSectionWaves, levelSectionMap:
		return name, nil
	}
	return "", fmt.Errorf("line %d: unknown section [%s]", number, name)
}

func stripComment(line string) string {
	if index := strings.Index(line, "#"); index >= 0 {
		line = line[:index]
	}
	return strings.TrimSpace(line)
}

// applyLevelSection разбирает пары key = value секции [level]
func applyLevelSection(result *levelFile, lines []sourceLine) error {
	for _, line := range lines {
		key, value, ok := strings.Cut(line.text, "=")
		if !ok {
			return fmt.Errorf(
				"line %d: expected key = value", line.number,
			)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "name":
			result.name = value
		case "max_active":
			number, err := parseUint(value, line.number, key)
			if err != nil {
				return err
			}
			if number < minLevelMaxActive || number > maxLevelMaxActive {
				return fmt.Errorf(
					"line %d: max_active must be within %d..%d",
					line.number, minLevelMaxActive, maxLevelMaxActive,
				)
			}
			result.maxActive = number
		case "time_3star":
			seconds, err := parseUint(value, line.number, key)
			if err != nil {
				return err
			}
			result.time3StarTicks = seconds * ticksPerSecond
		default:
			return fmt.Errorf(
				"line %d: unknown key %q in [level]", line.number, key,
			)
		}
	}
	return nil
}

func parseUint(value string, number int, key string) (uint, error) {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf(
			"line %d: %s must be a non-negative integer", number, key,
		)
	}
	return uint(parsed), nil
}

// parseWaves разбирает строки секции [waves]; второе значение —
// заданы ли носители бонусов явно (строчными буквами)
func parseWaves(
	lines []sourceLine,
	defaultDelay uint,
) ([]types.WaveSpec, bool, error) {
	if len(lines) == 0 {
		return nil, false, fmt.Errorf("section [waves] is empty")
	}

	waves := make([]types.WaveSpec, 0, len(lines))
	explicit := false
	for _, line := range lines {
		wave, hasBonus, err := ParseWaveLine(line.text, defaultDelay)
		if err != nil {
			return nil, false, fmt.Errorf("line %d: %w", line.number, err)
		}
		explicit = explicit || hasBonus
		waves = append(waves, wave)
	}
	return waves, explicit, nil
}

// ParseWaveLine разбирает строку волны «танки [delay] [start]»;
// второе значение — есть ли в волне носители бонусов
func ParseWaveLine(
	line string,
	defaultDelay uint,
) (types.WaveSpec, bool, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 || len(fields) > 3 {
		return types.WaveSpec{}, false, fmt.Errorf(
			"wave must be: tanks [delay] [start]",
		)
	}

	wave := types.WaveSpec{DelayTicks: defaultDelay}
	hasBonus := false
	for _, letter := range fields[0] {
		upper := []rune(strings.ToUpper(string(letter)))[0]
		level, ok := waveTankLevels[upper]
		if !ok {
			return types.WaveSpec{}, false, fmt.Errorf(
				"unknown tank letter %q (use B, F, P, A)", letter,
			)
		}
		bonus := letter != upper
		hasBonus = hasBonus || bonus
		wave.Tanks = append(wave.Tanks, types.WaveTank{
			Level:    level,
			HasBonus: bonus,
		})
	}

	if len(fields) > 1 {
		delay, err := strconv.ParseUint(fields[1], 10, 32)
		if err != nil {
			return types.WaveSpec{}, false, fmt.Errorf(
				"delay must be a non-negative integer, got %q", fields[1],
			)
		}
		wave.DelayTicks = uint(delay)
	}

	if len(fields) > 2 {
		start, err := parseWaveStart(fields[2])
		if err != nil {
			return types.WaveSpec{}, false, err
		}
		wave.Start = start
	}

	return wave, hasBonus, nil
}

func parseWaveStart(value string) (types.WaveStart, error) {
	switch {
	case value == "now":
		return types.WaveStart{Kind: types.WaveStartNow}, nil
	case value == "clear":
		return types.WaveStart{Kind: types.WaveStartClear}, nil
	case strings.HasPrefix(value, "left<="):
		left, err := strconv.ParseUint(
			strings.TrimPrefix(value, "left<="), 10, 32,
		)
		if err != nil {
			return types.WaveStart{}, fmt.Errorf(
				"malformed start %q, expected left<=N", value,
			)
		}
		return types.WaveStart{
			Kind: types.WaveStartLeft,
			Left: uint(left),
		}, nil
	}
	return types.WaveStart{}, fmt.Errorf(
		"unknown start %q (use now, clear or left<=N)", value,
	)
}

// ParseLevelDefaults собирает дефолты уровня из конфигурации:
// волны задаются строками в синтаксисе секции [waves]
func ParseLevelDefaults(
	maxActive uint,
	time3StarSeconds uint,
	delayTicks uint,
	waveLines []string,
) (types.LevelDefaults, error) {
	defaults := types.LevelDefaults{
		MaxActive:      maxActive,
		Time3StarTicks: time3StarSeconds * ticksPerSecond,
		DelayTicks:     delayTicks,
	}
	if maxActive < minLevelMaxActive || maxActive > maxLevelMaxActive {
		return defaults, fmt.Errorf(
			"default max_active must be within %d..%d",
			minLevelMaxActive, maxLevelMaxActive,
		)
	}
	if len(waveLines) == 0 {
		return defaults, fmt.Errorf("default waves are empty")
	}
	for index, line := range waveLines {
		wave, _, err := ParseWaveLine(line, delayTicks)
		if err != nil {
			return defaults, fmt.Errorf("default wave %d: %w", index+1, err)
		}
		defaults.Waves = append(defaults.Waves, wave)
	}
	return defaults, nil
}
