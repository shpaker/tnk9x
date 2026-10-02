//go:build darwin

package platform

import (
	"os/exec"
	"strings"
)

// osLanguage — первый язык интерфейса macOS (AppleLanguages):
// запуск из Finder не задаёт переменных локали
func osLanguage() string {
	output, err := exec.Command("defaults", "read", "-g", "AppleLanguages").
		Output()
	if err != nil {
		return ""
	}
	// Вывод — список plist: ( "ru-RU", "en-US" )
	_, rest, found := strings.Cut(string(output), `"`)
	if !found {
		return ""
	}
	language, _, _ := strings.Cut(rest, `"`)
	return language
}
