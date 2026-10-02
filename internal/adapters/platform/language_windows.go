//go:build windows

package platform

import "golang.org/x/sys/windows"

// osLanguage — первый язык интерфейса Windows ("ru-RU")
func osLanguage() string {
	languages, err := windows.GetUserPreferredUILanguages(
		windows.MUI_LANGUAGE_NAME,
	)
	if err != nil || len(languages) == 0 {
		return ""
	}
	return languages[0]
}
