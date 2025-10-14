package utils

import "fmt"

// FormatGreeting форматирует приветственное сообщение
func FormatGreeting(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// IsEmpty проверяет, является ли строка пустой
func IsEmpty(s string) bool {
	return s == ""
}

// TruncateString обрезает строку до указанной длины
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
