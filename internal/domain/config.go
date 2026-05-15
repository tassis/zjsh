package domain

type Config struct {
	Defaults Defaults
	Projects []Project
}

type Defaults struct {
	Shell                 string
	RestartOnResurrection bool
}

type Project struct {
	Name                  string
	Path                  string
	Session               string
	Startup               string
	Layout                string
	LayoutFile            string
	RestartOnResurrection *bool
}
