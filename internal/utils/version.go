package utils

import "fmt"

var (
	// Version версия приложения
	Version = "1.0.0"
	// BuildTime время сборки
	BuildTime = "unknown"
	// GitCommit хэш коммита
	GitCommit = "unknown"
)

// Info возвращает информацию о версии
func Info() string {
	return fmt.Sprintf("Version: %s, BuildTime: %s, GitCommit: %s", Version, BuildTime, GitCommit)
}

// ShortInfo возвращает краткую информацию о версии
func ShortInfo() string {
	return fmt.Sprintf("v%s", Version)
}
