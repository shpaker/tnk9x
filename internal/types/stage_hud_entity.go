package types

// StageHUDData — данные сессии уровня для боковой панели HUD:
// плоский DTO, чтобы рендер не зависел от сущностей сессии
type StageHUDData struct {
	EnemiesForSpawn uint
	PlayerCount     uint
	Player1Lives    uint
	Player2Lives    uint
}
