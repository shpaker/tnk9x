package stage

import (
	"fmt"
	"os"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

// Интеграционный smoke-тест: каждый реальный ogg-ассет декодируется
// в непустой PCM, пригодный для NewInfiniteLoop (длина кратна 4 —
// 16-битный стерео семпл). Полный адаптер не тестируем: audio.Context
// создаётся один раз на процесс и ненадёжен в headless-окружении
func TestDecodePCM_AllSoundAssets(t *testing.T) {
	for _, soundID := range types.AllSoundIDs() {
		path := fmt.Sprintf("../../../assets/sounds/%s.ogg", soundID)

		oggData, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			t.Skipf("Пропуск интеграционного теста: %s не найден", path)
		}
		if err != nil {
			t.Fatalf("чтение %s: %v", path, err)
		}

		pcm, err := decodePCM(44100, oggData)
		if err != nil {
			t.Fatalf("decodePCM(%s): %v", soundID, err)
		}
		if len(pcm) == 0 {
			t.Errorf("%s: ожидался непустой PCM", soundID)
		}
		if len(pcm)%4 != 0 {
			t.Errorf("%s: длина PCM %d не кратна 4", soundID, len(pcm))
		}
	}
}
