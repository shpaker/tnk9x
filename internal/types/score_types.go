package types

// ScorePlayerViewData — итоги этапа одного игрока: общий счёт забега
// и число убитых врагов по уровням 0-3
type ScorePlayerViewData struct {
	Score uint
	Kills [4]uint
}

// ScoreViewData — данные экрана подсчёта очков после этапа
type ScoreViewData struct {
	StageNumber uint
	PlayerCount uint
	HiScore     uint
	Player1     ScorePlayerViewData
	Player2     ScorePlayerViewData
}
