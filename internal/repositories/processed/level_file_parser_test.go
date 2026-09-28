package processed

import (
	"os"
	"strings"
	"testing"

	"github.com/shpaker/tnk9x/internal/repositories/raw"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
)

func testLevelDefaults() types.LevelDefaults {
	return types.LevelDefaults{
		MaxActive:      4,
		Time3StarTicks: 180 * ticksPerSecond,
		DelayTicks:     120,
		Waves: []types.WaveSpec{
			{Tanks: []types.WaveTank{{Level: 0}}, DelayTicks: 120},
		},
	}
}

const sectionedLevel = `# comment
[level]
name = TEST LEVEL   # trailing comment
max_active = 5
time_3star = 100

[waves]
# tanks  delay start
BBf       90
PAa       60   left<=2
A              
BB        30   clear

[map]
#.@.
..~%
`

func TestParseLevelFile_Sectioned(t *testing.T) {
	file, err := parseLevelFile(sectionedLevel, testLevelDefaults())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if file.name != "TEST LEVEL" || file.maxActive != 5 {
		t.Errorf("name %q max_active %d", file.name, file.maxActive)
	}
	if file.time3StarTicks != 100*ticksPerSecond {
		t.Errorf("time_3star ticks %d", file.time3StarTicks)
	}
	if !file.explicitBonuses {
		t.Error("lowercase letters must mark explicit bonuses")
	}
	if len(file.waves) != 4 {
		t.Fatalf("waves %d, want 4", len(file.waves))
	}

	first := file.waves[0]
	if len(first.Tanks) != 3 || first.DelayTicks != 90 ||
		first.Start.Kind != types.WaveStartNow {
		t.Errorf("first wave %+v", first)
	}
	if !first.Tanks[2].HasBonus ||
		first.Tanks[2].Level != types.EnemyLevelFast {
		t.Errorf("third tank %+v, want bonus fast", first.Tanks[2])
	}

	second := file.waves[1]
	if second.Start.Kind != types.WaveStartLeft || second.Start.Left != 2 {
		t.Errorf("second wave start %+v", second.Start)
	}
	if file.waves[2].DelayTicks != 120 {
		t.Errorf("missing delay must use the default, got %d",
			file.waves[2].DelayTicks)
	}
	if file.waves[3].Start.Kind != types.WaveStartClear {
		t.Errorf("fourth wave start %+v", file.waves[3].Start)
	}

	// # в [map] — кирпич, а не комментарий
	if len(file.mapLines) != 2 || file.mapLines[0] != "#.@." {
		t.Errorf("map lines %q", file.mapLines)
	}
}

func TestParseLevelFile_Legacy(t *testing.T) {
	defaults := testLevelDefaults()
	file, err := parseLevelFile("#.\n..\n", defaults)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(file.mapLines) != 2 || file.mapLines[0] != "#." {
		t.Errorf("legacy map lines %q", file.mapLines)
	}
	if len(file.waves) != len(defaults.Waves) ||
		file.maxActive != defaults.MaxActive {
		t.Error("legacy map must use default waves and settings")
	}
}

func TestParseLevelFile_Errors(t *testing.T) {
	tests := map[string]string{
		"unknown section": "[level]\n[boss]\n[map]\n..",
		"unknown key":     "[level]\nspeed = 3\n[map]\n..",
		"bad max_active":  "[level]\nmax_active = 11\n[map]\n..",
		"bad letter":      "[waves]\nBXB\n[map]\n..",
		"bad delay":       "[waves]\nBB fast\n[map]\n..",
		"bad start":       "[waves]\nBB 10 later\n[map]\n..",
		"bad left":        "[waves]\nBB 10 left<=x\n[map]\n..",
		"empty waves":     "[waves]\n[map]\n..",
		"missing map":     "[level]\nname = X\n",
		"duplicate":       "[map]\n..\n[map]\n..",
		"too many fields": "[waves]\nBB 10 now extra\n[map]\n..",
	}
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := parseLevelFile(data, testLevelDefaults()); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestParseLevelFile_ErrorHasLineNumber(t *testing.T) {
	_, err := parseLevelFile("[waves]\nBB\nBQ\n[map]\n..", testLevelDefaults())
	if err == nil || !strings.Contains(err.Error(), "line 3") {
		t.Errorf("error %v, want a line 3 reference", err)
	}
}

func TestParseLevelDefaults(t *testing.T) {
	defaults, err := ParseLevelDefaults(4, 60, 100, []string{"BB", "FA 50"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if defaults.Time3StarTicks != 60*ticksPerSecond ||
		len(defaults.Waves) != 2 ||
		defaults.Waves[0].DelayTicks != 100 ||
		defaults.Waves[1].DelayTicks != 50 {
		t.Errorf("defaults %+v", defaults)
	}

	if _, err := ParseLevelDefaults(0, 60, 100, []string{"BB"}); err == nil {
		t.Error("max_active 0 must be rejected")
	}
	if _, err := ParseLevelDefaults(4, 60, 100, nil); err == nil {
		t.Error("empty default waves must be rejected")
	}
}

func TestGetLevel_SectionedEntity(t *testing.T) {
	fileRepository := NewMockFileRepository()
	fileRepository.AddFile("levels/3.bcmap", []byte(sectionedLevel))
	repository := NewMapsDataRepository(
		fileRepository,
		&testutil.FakeTilesetRegistry{},
		testLevelDefaults(),
	)

	level, err := repository.GetLevel(3, 8)
	if err != nil {
		t.Fatalf("GetLevel: %v", err)
	}
	if level.GetNumber() != 3 || level.GetName() != "TEST LEVEL" {
		t.Errorf("level %d %q", level.GetNumber(), level.GetName())
	}
	if level.GetTotalEnemies() != 9 {
		t.Errorf("total enemies %d, want 9", level.GetTotalEnemies())
	}
	if counts := level.GetEnemyCounts(); counts != [4]uint{4, 1, 1, 3} {
		t.Errorf("enemy counts %v", counts)
	}
	if !repository.HasLevel(3) || repository.HasLevel(4) {
		t.Error("HasLevel must reflect existing files")
	}
}

// Без имени в файле уровень называется по номеру
func TestGetLevel_DefaultName(t *testing.T) {
	fileRepository := NewMockFileRepository()
	fileRepository.AddFile("levels/7.bcmap", []byte("..\n.."))
	repository := NewMapsDataRepository(
		fileRepository,
		&testutil.FakeTilesetRegistry{},
		testLevelDefaults(),
	)

	level, err := repository.GetLevel(7, 8)
	if err != nil {
		t.Fatalf("GetLevel: %v", err)
	}
	if level.GetName() != "STAGE 07" {
		t.Errorf("name %q, want STAGE 07", level.GetName())
	}
}

// Все карты и кампания из assets разбираются без ошибок
func TestAssets_CampaignLevelsParse(t *testing.T) {
	fileRepository := raw.NewFileRepository(os.DirFS("../../../assets"))

	campaign, err := NewCampaignRepository(fileRepository).GetCampaign("main")
	if err != nil {
		t.Fatalf("campaign: %v", err)
	}

	levels := NewMapsDataRepository(
		fileRepository,
		&testutil.FakeTilesetRegistry{},
		testLevelDefaults(),
	)
	count, err := levels.GetLevelsCount()
	if err != nil {
		t.Fatalf("levels count: %v", err)
	}

	listed := 0
	for _, pack := range campaign.GetPacks() {
		for _, number := range pack.Levels {
			listed++
			level, err := levels.GetLevel(number, 8)
			if err != nil {
				t.Errorf("level %d: %v", number, err)
				continue
			}
			size := level.GetMap().GetSizePx()
			if size.Width != 26*8 || size.Height != 26*8 {
				t.Errorf("level %d: map size %v, want 26x26", number, size)
			}
			if level.GetTotalEnemies() == 0 {
				t.Errorf("level %d has no enemies", number)
			}
		}
	}
	if listed != count {
		t.Errorf("campaign lists %d levels, folder has %d", listed, count)
	}
}
