package processed

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ICampaignRepository = (*CampaignRepository)(nil)

// CampaignRepository читает кампании из levels/<name>.tnkcamp
type CampaignRepository struct {
	fileRepository interfaces.IFileRepository
}

func NewCampaignRepository(
	fileRepository interfaces.IFileRepository,
) *CampaignRepository {
	return &CampaignRepository{
		fileRepository: fileRepository,
	}
}

// GetCampaign читает и проверяет кампанию: уровни не повторяются,
// unlock_after ссылается на уровень из предыдущих пачек
func (cr *CampaignRepository) GetCampaign(
	name string,
) (*types.CampaignEntity, error) {
	fileName := "levels/" + name + ".tnkcamp"
	data, err := cr.fileRepository.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read campaign %q: %w", name, err)
	}

	campaign, err := parseCampaign(string(data))
	if err != nil {
		return nil, fmt.Errorf("campaign %q: %w", name, err)
	}
	return campaign, nil
}

func parseCampaign(data string) (*types.CampaignEntity, error) {
	lines := strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n")

	campaignName := ""
	var packs []types.PackSpec
	section := ""

	for index, raw := range lines {
		number := index + 1
		line := stripComment(strings.TrimSpace(raw))
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") {
			switch line {
			case "[campaign]":
				if section != "" {
					return nil, fmt.Errorf(
						"line %d: [campaign] must be the first section",
						number,
					)
				}
				section = "campaign"
			case "[pack]":
				section = "pack"
				packs = append(packs, types.PackSpec{})
			default:
				return nil, fmt.Errorf(
					"line %d: unknown section %s", number, line,
				)
			}
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: expected key = value", number)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch section {
		case "campaign":
			if key != "name" {
				return nil, fmt.Errorf(
					"line %d: unknown key %q in [campaign]", number, key,
				)
			}
			campaignName = value
		case "pack":
			if err := applyPackKey(
				&packs[len(packs)-1], key, value, number,
			); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf(
				"line %d: content before the first section", number,
			)
		}
	}

	if err := validatePacks(packs); err != nil {
		return nil, err
	}

	return types.NewCampaignEntity(campaignName, packs), nil
}

func applyPackKey(
	pack *types.PackSpec,
	key, value string,
	number int,
) error {
	switch key {
	case "name":
		pack.Name = value
	case "levels":
		levels, err := parseLevelList(value)
		if err != nil {
			return fmt.Errorf("line %d: %w", number, err)
		}
		pack.Levels = levels
	case "order":
		switch value {
		case "sequential":
			pack.Order = types.PackOrderSequential
		case "any":
			pack.Order = types.PackOrderAny
		default:
			return fmt.Errorf(
				"line %d: order must be sequential or any", number,
			)
		}
	case "unlock_stars":
		stars, err := parseUint(value, number, key)
		if err != nil {
			return err
		}
		pack.UnlockStars = stars
	case "unlock_after":
		level, err := strconv.Atoi(value)
		if err != nil || level < 1 {
			return fmt.Errorf(
				"line %d: unlock_after must be a level number", number,
			)
		}
		pack.UnlockAfter = level
	default:
		return fmt.Errorf("line %d: unknown key %q in [pack]", number, key)
	}
	return nil
}

// parseLevelList разбирает список уровней: «1-5», «1,2,7», «1-3,8»
func parseLevelList(value string) ([]int, error) {
	var levels []int
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		from, to, isRange := strings.Cut(part, "-")
		first, err := strconv.Atoi(strings.TrimSpace(from))
		if err != nil || first < 1 {
			return nil, fmt.Errorf("malformed levels %q", value)
		}
		last := first
		if isRange {
			last, err = strconv.Atoi(strings.TrimSpace(to))
			if err != nil || last < first {
				return nil, fmt.Errorf("malformed levels %q", value)
			}
		}
		for level := first; level <= last; level++ {
			levels = append(levels, level)
		}
	}
	return levels, nil
}

func validatePacks(packs []types.PackSpec) error {
	if len(packs) == 0 {
		return fmt.Errorf("no [pack] sections")
	}

	seen := make(map[int]bool)
	for index, pack := range packs {
		if len(pack.Levels) == 0 {
			return fmt.Errorf("pack %d has no levels", index+1)
		}
		if pack.UnlockAfter != 0 && !seen[pack.UnlockAfter] {
			return fmt.Errorf(
				"pack %d: unlock_after %d is not in a previous pack",
				index+1, pack.UnlockAfter,
			)
		}
		for _, level := range pack.Levels {
			if seen[level] {
				return fmt.Errorf(
					"pack %d: level %d is listed twice", index+1, level,
				)
			}
			seen[level] = true
		}
	}
	return nil
}
