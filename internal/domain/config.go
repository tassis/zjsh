package domain

type Config struct {
	Defaults Defaults
	Projects []Project
	Macros   []Macro
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
	Macro         string
}

func DefaultIcons() Icons {
	return Icons{
		Project:       "◆",
		Session:       "●",
		Resurrectable: "↺",
		Path:          "→",
		Macro:         "▶",
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

type Macro struct {
	Name string   `json:"name"`
	CWD  string   `json:"cwd,omitempty"`
	Exec []string `json:"exec"`
}
