package internal

// Version содержит номер версии приложения и переопределяется при сборке через ldflags.
var Version = "dev"

// Debug определяет, включен ли режим отладки.
var Debug bool = false
