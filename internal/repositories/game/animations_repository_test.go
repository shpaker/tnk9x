package game

import (
	"testing"

	"github.com/shpaker/tnk25/internal/types"
	image_providers "github.com/shpaker/tnk25/internal/types/image_providers"
)

func TestNewAnimationsRepository(t *testing.T) {
	repo := NewAnimationsRepository()

	if repo == nil {
		t.Fatal("NewAnimationsRepository returned nil")
	}

	animations := repo.GetAllAnimations()
	if len(animations) != 0 {
		t.Errorf("Expected 0 animations, got %d", len(animations))
	}
}

func TestAddAnimation(t *testing.T) {
	repo := NewAnimationsRepository()

	animation := image_providers.NewAnimationProviderWithLoops(
		types.AnimationData{
			types.AnimationDataFrame{
				Image:    "frame1",
				Duration: 10,
			},
			types.AnimationDataFrame{
				Image:    "frame2",
				Duration: 10,
			},
		},
		5,
	)

	repo.AddAnimation(animation)

	animations := repo.GetAllAnimations()
	if len(animations) != 1 {
		t.Errorf("Expected 1 animation, got %d", len(animations))
	}

	if animations[0] != animation {
		t.Error("Added animation does not match")
	}
}

func TestAddMultipleAnimations(t *testing.T) {
	repo := NewAnimationsRepository()

	anim1 := image_providers.NewAnimationProviderWithLoops(
		types.AnimationData{
			types.AnimationDataFrame{
				Image:    "frame1",
				Duration: 10,
			},
		},
		3,
	)

	anim2 := image_providers.NewAnimationProvider(
		types.AnimationData{
			types.AnimationDataFrame{
				Image:    "frame1",
				Duration: 10,
			},
		},
	)

	anim3 := image_providers.NewAnimationProviderWithLoops(
		types.AnimationData{
			types.AnimationDataFrame{
				Image:    "frame1",
				Duration: 20,
			},
			types.AnimationDataFrame{
				Image:    "frame2",
				Duration: 20,
			},
		},
		10,
	)

	repo.AddAnimation(anim1)
	repo.AddAnimation(anim2)
	repo.AddAnimation(anim3)

	animations := repo.GetAllAnimations()
	if len(animations) != 3 {
		t.Errorf("Expected 3 animations, got %d", len(animations))
	}

	if animations[0] != anim1 {
		t.Error("First animation does not match")
	}

	if animations[1] != anim2 {
		t.Error("Second animation does not match")
	}

	if animations[2] != anim3 {
		t.Error("Third animation does not match")
	}
}

func TestGetAllAnimations(t *testing.T) {
	repo := NewAnimationsRepository()

	animations := repo.GetAllAnimations()
	if len(animations) != 0 {
		t.Errorf("Expected 0 animations initially, got %d", len(animations))
	}

	anim1 := image_providers.NewAnimationProvider(
		types.AnimationData{
			types.AnimationDataFrame{
				Image:    "frame1",
				Duration: 10,
			},
		},
	)

	anim2 := image_providers.NewAnimationProvider(
		types.AnimationData{
			types.AnimationDataFrame{
				Image:    "frame2",
				Duration: 10,
			},
		},
	)

	repo.AddAnimation(anim1)
	repo.AddAnimation(anim2)

	animations = repo.GetAllAnimations()
	if len(animations) != 2 {
		t.Errorf("Expected 2 animations, got %d", len(animations))
	}
}

func TestGetAllAnimationsReturnsSliceReference(t *testing.T) {
	repo := NewAnimationsRepository()

	anim := image_providers.NewAnimationProvider(
		types.AnimationData{
			types.AnimationDataFrame{
				Image:    "frame1",
				Duration: 10,
			},
		},
	)

	repo.AddAnimation(anim)

	animations := repo.GetAllAnimations()
	if len(animations) == 0 {
		t.Fatal("Expected at least one animation")
	}

	animations[0].IsAnimating = true

	animations2 := repo.GetAllAnimations()
	if !animations2[0].IsAnimating {
		t.Error("Modification of animation did not persist")
	}
}

func TestEmptyRepository(t *testing.T) {
	repo := NewAnimationsRepository()

	animations := repo.GetAllAnimations()
	if animations == nil {
		t.Fatal("GetAllAnimations should not return nil")
	}

	if len(animations) != 0 {
		t.Errorf("Expected empty slice, got %d animations", len(animations))
	}
}
