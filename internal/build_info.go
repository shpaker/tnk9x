package internal

// Version содержит номер версии приложения и переопределяется при сборке через ldflags.
var Version = "dev"

// Debug определяет, включен ли режим отладки.
var Debug bool

// DebugFlag используется для установки режима отладки через ldflags.
var DebugFlag = "false"

func init() {
	if DebugFlag == "true" {
		Debug = true
	}
}
