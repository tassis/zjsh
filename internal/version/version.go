package version

const AppName = "zjsh"

var Version = "dev"

func String() string {
	return AppName + " " + Version
}
