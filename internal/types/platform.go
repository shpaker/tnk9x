package types

// RewardStatus — исход запроса рекламы за вознаграждение
type RewardStatus int

const (
	// RewardStatusNone — запроса нет или его исход уже получен
	RewardStatusNone RewardStatus = iota
	// RewardStatusPending — реклама ещё показывается
	RewardStatusPending
	// RewardStatusGranted — реклама досмотрена, награда положена
	RewardStatusGranted
	// RewardStatusDenied — реклама закрыта раньше, недоступна
	// или завершилась ошибкой: награды нет
	RewardStatusDenied
)
