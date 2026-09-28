package processed

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

const testCampaign = `# comment
[campaign]
name = MAIN

[pack]
name = ONE
levels = 1-3

[pack]
name = TWO
levels = 4,5, 7-8
order = any
unlock_stars = 5
unlock_after = 3
`

func TestGetCampaign(t *testing.T) {
	fileRepository := NewMockFileRepository()
	fileRepository.AddFile("levels/main.bccamp", []byte(testCampaign))

	campaign, err := NewCampaignRepository(fileRepository).GetCampaign("main")
	if err != nil {
		t.Fatalf("GetCampaign: %v", err)
	}

	if campaign.GetName() != "MAIN" || len(campaign.GetPacks()) != 2 {
		t.Fatalf("campaign %q with %d packs",
			campaign.GetName(), len(campaign.GetPacks()))
	}
	first, second := campaign.GetPacks()[0], campaign.GetPacks()[1]
	if first.Order != types.PackOrderSequential || len(first.Levels) != 3 {
		t.Errorf("first pack %+v", first)
	}
	want := []int{4, 5, 7, 8}
	for i, level := range want {
		if second.Levels[i] != level {
			t.Errorf("second pack levels %v, want %v", second.Levels, want)
			break
		}
	}
	if second.Order != types.PackOrderAny ||
		second.UnlockStars != 5 || second.UnlockAfter != 3 {
		t.Errorf("second pack %+v", second)
	}

	if next, ok := campaign.NextLevel(3); !ok || next != 4 {
		t.Errorf("next after 3: %d %v", next, ok)
	}
	if _, ok := campaign.NextLevel(8); ok {
		t.Error("the last level has no next level")
	}
	if campaign.MaxStars() != 21 {
		t.Errorf("max stars %d, want 21", campaign.MaxStars())
	}
}

func TestGetCampaign_Errors(t *testing.T) {
	tests := map[string]string{
		"no packs":         "[campaign]\nname = X\n",
		"empty pack":       "[pack]\nname = A\n",
		"duplicate level":  "[pack]\nlevels = 1-2\n[pack]\nlevels = 2\n",
		"unlock ahead":     "[pack]\nlevels = 1\nunlock_after = 2\n[pack]\nlevels = 2\n",
		"bad range":        "[pack]\nlevels = 3-1\n",
		"bad order":        "[pack]\nlevels = 1\norder = random\n",
		"unknown key":      "[pack]\nlevels = 1\ncolor = red\n",
		"unknown section":  "[bonus]\n",
		"campaign not top": "[pack]\nlevels = 1\n[campaign]\n",
		"no section":       "levels = 1\n",
	}
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			fileRepository := NewMockFileRepository()
			fileRepository.AddFile("levels/main.bccamp", []byte(data))
			_, err := NewCampaignRepository(fileRepository).GetCampaign("main")
			if err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestGetCampaign_MissingFile(t *testing.T) {
	_, err := NewCampaignRepository(NewMockFileRepository()).GetCampaign("main")
	if err == nil {
		t.Error("expected an error for a missing campaign")
	}
}
