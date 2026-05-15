package domain

type EntryType string

const (
	EntrySession EntryType = "session"
	EntryPath    EntryType = "path"
	EntryProject EntryType = "project"
)

type SessionState string

const (
	SessionStateNone          SessionState = ""
	SessionStateLive          SessionState = "live"
	SessionStateResurrectable SessionState = "resurrectable"
)

type Entry struct {
	Name                  string       `json:"name"`
	Type                  EntryType    `json:"type"`
	Sources               []string     `json:"sources"`
	Path                  string       `json:"path,omitempty"`
	SessionName           string       `json:"session_name,omitempty"`
	SessionState          SessionState `json:"session_state,omitempty"`
	Shell                 string       `json:"shell,omitempty"`
	Startup               string       `json:"startup,omitempty"`
	Layout                string       `json:"layout,omitempty"`
	LayoutFile            string       `json:"layout_file,omitempty"`
	RestartOnResurrection bool         `json:"restart_on_resurrection,omitempty"`
	Score                 int          `json:"score"`
	Order                 int          `json:"-"`
}
