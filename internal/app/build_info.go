package app

var Version = "dev"

var Debug bool

var DebugFlag = "false"

func init() {
	if DebugFlag == "true" {
		Debug = true
	}
}
