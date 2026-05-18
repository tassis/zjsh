package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/saweima12/zjsh/internal/domain"
	"github.com/saweima12/zjsh/internal/platform"
	kdl "github.com/sblinch/kdl-go"
	"github.com/sblinch/kdl-go/document"
)

type Loader struct {
	Path string
}

func (l Loader) ResolvedPath() (string, error) {
	path := l.Path
	if path == "" {
		var err error
		path, err = platform.DefaultConfigPath()
		if err != nil {
			return "", err
		}
	}
	return platform.ExpandHome(path)
}

func (l Loader) Load(context.Context) (domain.Config, error) {
	path, err := l.ResolvedPath()
	if err != nil {
		return domain.Config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return domain.Config{}, err
	}
	config, err := parseConfig(string(data))
	if err != nil {
		return domain.Config{}, err
	}
	return normalizeConfigPaths(config)
}

func normalizeConfigPaths(config domain.Config) (domain.Config, error) {
	for i := range config.Projects {
		project := &config.Projects[i]
		if project.Path != "" {
			path, err := platform.ExpandHome(project.Path)
			if err != nil {
				return domain.Config{}, fmt.Errorf("project %q: path: %w", project.Name, err)
			}
			project.Path = path
		}

		if project.LayoutFile == "" {
			continue
		}
		layoutFile, err := platform.ExpandHome(project.LayoutFile)
		if err != nil {
			return domain.Config{}, fmt.Errorf("project %q: layout_file: %w", project.Name, err)
		}
		project.LayoutFile = layoutFile
	}
	return config, nil
}

func parseConfig(input string) (domain.Config, error) {
	config := defaultConfig()
	doc, err := kdl.Parse(strings.NewReader(input))
	if err != nil {
		return domain.Config{}, err
	}
	if err := validateStrictBooleanNodes(doc.Nodes); err != nil {
		return domain.Config{}, err
	}
	var raw rawConfig
	dec := kdl.NewDecoder(strings.NewReader(input))
	dec.Options.AllowUnhandledNodes = true
	if err := dec.Decode(&raw); err != nil {
		return domain.Config{}, err
	}
	config.Defaults = defaultsWithFallback(config.Defaults, raw.Defaults.toDomain())
	projects, err := raw.projects()
	if err != nil {
		return domain.Config{}, err
	}
	config.Projects = projects
	return config, nil
}

func validateStrictBooleanNodes(nodes []*document.Node) error {
	for _, node := range nodes {
		switch nodeName(node) {
		case "defaults":
			for _, child := range node.Children {
				if nodeName(child) == "restart_on_resurrection" {
					if err := validateBooleanNode(child, "defaults.restart_on_resurrection"); err != nil {
						return err
					}
				}
			}
		case "project":
			projectName := "project"
			if len(node.Arguments) > 0 {
				projectName = fmt.Sprintf("project %q", node.Arguments[0].Value)
			}
			for _, child := range node.Children {
				switch nodeName(child) {
				case "cwd", "restart_on_resurrection":
					if err := validateBooleanNode(child, projectName+"."+nodeName(child)); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func validateBooleanNode(node *document.Node, label string) error {
	if len(node.Arguments) != 1 {
		return fmt.Errorf("%s: expected a single boolean value", label)
	}
	if _, ok := node.Arguments[0].Value.(bool); !ok {
		return fmt.Errorf("%s: expected boolean value", label)
	}
	return nil
}

func nodeName(node *document.Node) string {
	if node == nil || node.Name == nil {
		return ""
	}
	name, ok := node.Name.Value.(string)
	if !ok {
		return ""
	}
	return name
}

func defaultConfig() domain.Config {
	return domain.Config{
		Defaults: domain.Defaults{
			Shell:                 "sh",
			RestartOnResurrection: false,
			Icons:                 domain.DefaultIcons(),
		},
	}
}

func defaultsWithFallback(base, parsed domain.Defaults) domain.Defaults {
	if parsed.Shell != "" {
		base.Shell = parsed.Shell
	}
	base.RestartOnResurrection = parsed.RestartOnResurrection
	base.Icons = iconsWithFallback(base.Icons, parsed.Icons)
	return base
}

func iconsWithFallback(base, parsed domain.Icons) domain.Icons {
	if parsed.Project != "" {
		base.Project = parsed.Project
	}
	if parsed.Session != "" {
		base.Session = parsed.Session
	}
	if parsed.Resurrectable != "" {
		base.Resurrectable = parsed.Resurrectable
	}
	if parsed.Path != "" {
		base.Path = parsed.Path
	}
	return base
}

type rawConfig struct {
	Defaults rawDefaults  `kdl:"defaults"`
	Projects []rawProject `kdl:"project,multiple"`
}

type rawDefaults struct {
	Shell                 string `kdl:"shell"`
	RestartOnResurrection bool   `kdl:"restart_on_resurrection"`
	IconProject           string `kdl:"icon_project"`
	IconSession           string `kdl:"icon_session"`
	IconResurrectable     string `kdl:"icon_resurrectable"`
	IconPath              string `kdl:"icon_path"`
}

func (d rawDefaults) toDomain() domain.Defaults {
	return domain.Defaults{
		Shell:                 d.Shell,
		RestartOnResurrection: d.RestartOnResurrection,
		Icons: domain.Icons{
			Project:       d.IconProject,
			Session:       d.IconSession,
			Resurrectable: d.IconResurrectable,
			Path:          d.IconPath,
		},
	}
}

type rawProject struct {
	Name                  string `kdl:",arg"`
	Path                  string `kdl:"path"`
	CWD                   bool   `kdl:"cwd"`
	Session               string `kdl:"session"`
	Startup               string `kdl:"startup"`
	Layout                string `kdl:"layout"`
	LayoutFile            string `kdl:"layout_file"`
	RestartOnResurrection *bool  `kdl:"restart_on_resurrection"`
}

func (c rawConfig) projects() ([]domain.Project, error) {
	projects := make([]domain.Project, 0, len(c.Projects))
	for _, raw := range c.Projects {
		project := domain.Project{
			Name:                  strings.TrimSpace(raw.Name),
			Path:                  raw.Path,
			CWD:                   raw.CWD,
			Session:               raw.Session,
			Startup:               raw.Startup,
			Layout:                raw.Layout,
			LayoutFile:            raw.LayoutFile,
			RestartOnResurrection: raw.RestartOnResurrection,
		}
		if project.Name == "" {
			return nil, fmt.Errorf("project node requires a single name argument")
		}
		if project.Path != "" && project.CWD {
			return nil, fmt.Errorf("project %q: path and cwd are mutually exclusive", project.Name)
		}
		if project.Path == "" && !project.CWD {
			return nil, fmt.Errorf("project %q: path or cwd true is required", project.Name)
		}
		projects = append(projects, project)
	}
	return projects, nil
}
