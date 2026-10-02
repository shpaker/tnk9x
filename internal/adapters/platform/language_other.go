//go:build !js && !darwin && !windows

package platform

// osLanguage — кроме переменных локали язык ОС узнать неоткуда
func osLanguage() string {
	return ""
}
