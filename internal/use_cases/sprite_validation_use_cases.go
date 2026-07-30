package use_cases

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// SpriteValidationUseCases проверяет при старте, что все спрайты
// и анимации из манифеста существуют в загруженных тайлсетах
type SpriteValidationUseCases struct {
	tilesetRegistry interfaces.ITilesetRepositoryRegistry
}

func NewSpriteValidationUseCases(
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
) *SpriteValidationUseCases {
	return &SpriteValidationUseCases{tilesetRegistry: tilesetRegistry}
}

// Validate пробует каждый идентификатор манифеста и собирает все
// несоответствия в одну ошибку — старт падает с полным списком проблем
func (uc *SpriteValidationUseCases) Validate(
	manifest types.SpriteManifest,
) error {
	var problems []string

	forEachID(manifest.Images, func(tilesetType types.TilesetType, id string) {
		if _, err := uc.tilesetRegistry.GetImageData(tilesetType, id); err != nil {
			problems = append(
				problems,
				fmt.Sprintf("image %s/%s: %v", tilesetType, id, err),
			)
		}
	})

	forEachID(
		manifest.Animations,
		func(tilesetType types.TilesetType, id string) {
			_, err := uc.tilesetRegistry.GetAnimationData(tilesetType, id)
			if err != nil {
				problems = append(
					problems,
					fmt.Sprintf("animation %s/%s: %v", tilesetType, id, err),
				)
			}
		},
	)

	if len(problems) > 0 {
		return fmt.Errorf(
			"sprite validation failed: %s",
			strings.Join(problems, "; "),
		)
	}
	return nil
}

// forEachID обходит манифест в детерминированном порядке тайлсетов,
// чтобы сообщение об ошибке было стабильным
func forEachID(
	ids map[types.TilesetType][]string,
	visit func(tilesetType types.TilesetType, id string),
) {
	tilesetTypes := make([]types.TilesetType, 0, len(ids))
	for tilesetType := range ids {
		tilesetTypes = append(tilesetTypes, tilesetType)
	}
	sort.Slice(tilesetTypes, func(i, j int) bool {
		return tilesetTypes[i] < tilesetTypes[j]
	})

	for _, tilesetType := range tilesetTypes {
		for _, id := range ids[tilesetType] {
			visit(tilesetType, id)
		}
	}
}
