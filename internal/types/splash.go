package types

// SplashViewData — сплеш-экран для отрисовки: прогресс загрузки
// ресурсов и затемнение (0 — нет, 1 — полностью чёрный)
type SplashViewData struct {
	Progress float64
	Fade     float64
}
