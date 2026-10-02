//go:build !js

package platform

// localeVariables — переменные локали POSIX по старшинству
var localeVariables = []string{"LC_ALL", "LC_MESSAGES", "LANG"}

// languageFromEnv — первая значимая переменная локали ("ru_RU.UTF-8");
// C и POSIX — локаль без языка
func languageFromEnv(lookup func(string) string) string {
	for _, name := range localeVariables {
		switch value := lookup(name); value {
		case "", "C", "POSIX", "C.UTF-8":
			continue
		default:
			return value
		}
	}
	return ""
}
