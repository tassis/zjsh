package domain

type Config struct {
	Defaults Defaults
	Projects []Project
}

type Defaults struct {
	Shell                 string
	RestartOnResurrection bool
	Icons                 Icons
}

type Icons struct {
	Project       string
	Session       string
	Resurrectable string
	Path          string
}

func DefaultIcons() Icons {
	return Icons{
		Project:       "◆",
		Session:       "●",
		Resurrectable: "↺",
		Path:          "→",
	}
}

type Project struct {
	Name                  string
	Path                  string
	CWD                   bool
	Session               string
	Startup               string
	Layout                string
	LayoutFile            string
	RestartOnResurrection *bool
}
