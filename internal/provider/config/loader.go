package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/saweima12/zjsh/internal/domain"
	"github.com/saweima12/zjsh/internal/platform"
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
		path, err := platform.ExpandHome(project.Path)
		if err != nil {
			return domain.Config{}, fmt.Errorf("project %q: path: %w", project.Name, err)
		}
		project.Path = path

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
	p := parser{tokens: tokenize(input)}
	config := defaultConfig()

	for !p.done() {
		node, err := p.nextNode()
		if err != nil {
			return domain.Config{}, err
		}
		switch node.name {
		case "defaults":
			defaults, err := decodeDefaults(node)
			if err != nil {
				return domain.Config{}, err
			}
			config.Defaults = defaultsWithFallback(config.Defaults, defaults)
		case "project":
			project, err := decodeProject(node)
			if err != nil {
				return domain.Config{}, err
			}
			if project.Path == "" {
				return domain.Config{}, fmt.Errorf("project %q: path is required", project.Name)
			}
			config.Projects = append(config.Projects, project)
		default:
			return domain.Config{}, fmt.Errorf("unsupported top-level node %q", node.name)
		}
	}
	return config, nil
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

type node struct {
	name string
	args []string
	body map[string]string
}

func decodeDefaults(n node) (domain.Defaults, error) {
	var defaults domain.Defaults
	for key, value := range n.body {
		switch key {
		case "shell":
			defaults.Shell = value
		case "restart_on_resurrection":
			defaults.RestartOnResurrection = value == "true"
		case "icon_project":
			defaults.Icons.Project = value
		case "icon_session":
			defaults.Icons.Session = value
		case "icon_resurrectable":
			defaults.Icons.Resurrectable = value
		case "icon_path":
			defaults.Icons.Path = value
		default:
			return domain.Defaults{}, fmt.Errorf("defaults: unsupported key %q", key)
		}
	}
	return defaults, nil
}

func decodeProject(n node) (domain.Project, error) {
	if len(n.args) != 1 {
		return domain.Project{}, fmt.Errorf("project node requires a single name argument")
	}
	project := domain.Project{Name: n.args[0]}
	for key, value := range n.body {
		switch key {
		case "path":
			project.Path = value
		case "session":
			project.Session = value
		case "startup":
			project.Startup = value
		case "layout":
			project.Layout = value
		case "layout_file":
			project.LayoutFile = value
		case "restart_on_resurrection":
			restart := value == "true"
			project.RestartOnResurrection = &restart
		default:
			return domain.Project{}, fmt.Errorf("project %q: unsupported key %q", project.Name, key)
		}
	}
	return project, nil
}

type parser struct {
	tokens []string
	pos    int
}

func (p *parser) done() bool {
	return p.pos >= len(p.tokens)
}

func (p *parser) nextNode() (node, error) {
	name, err := p.consumeIdent()
	if err != nil {
		return node{}, err
	}
	n := node{name: name, body: map[string]string{}}
	for !p.done() && p.peek() != "{" {
		arg, err := p.consumeValue()
		if err != nil {
			return node{}, err
		}
		n.args = append(n.args, arg)
	}
	if _, err := p.consume("{"); err != nil {
		return node{}, err
	}
	for !p.done() && p.peek() != "}" {
		key, err := p.consumeIdent()
		if err != nil {
			return node{}, err
		}
		value, err := p.consumeValue()
		if err != nil {
			return node{}, err
		}
		n.body[key] = value
	}
	if _, err := p.consume("}"); err != nil {
		return node{}, err
	}
	return n, nil
}

func (p *parser) consume(expected string) (string, error) {
	if p.done() {
		return "", fmt.Errorf("expected %q, got EOF", expected)
	}
	tok := p.tokens[p.pos]
	if tok != expected {
		return "", fmt.Errorf("expected %q, got %q", expected, tok)
	}
	p.pos++
	return tok, nil
}

func (p *parser) consumeIdent() (string, error) {
	if p.done() {
		return "", fmt.Errorf("expected identifier, got EOF")
	}
	tok := p.tokens[p.pos]
	if tok == "{" || tok == "}" {
		return "", fmt.Errorf("expected identifier, got %q", tok)
	}
	p.pos++
	return tok, nil
}

func (p *parser) consumeValue() (string, error) {
	return p.consumeIdent()
}

func (p *parser) peek() string {
	if p.done() {
		return ""
	}
	return p.tokens[p.pos]
}

func tokenize(input string) []string {
	var tokens []string
	for i := 0; i < len(input); {
		switch input[i] {
		case ' ', '\t', '\n', '\r':
			i++
		case '{', '}':
			tokens = append(tokens, input[i:i+1])
			i++
		case '"':
			value, next := scanQuoted(input, i)
			tokens = append(tokens, value)
			i = next
		case '/':
			if i+1 < len(input) && input[i+1] == '/' {
				i = skipToLineEnd(input, i+2)
				continue
			}
			fallthrough
		default:
			start := i
			for i < len(input) && !strings.ContainsRune(" \t\n\r{}", rune(input[i])) {
				if input[i] == '/' && i+1 < len(input) && input[i+1] == '/' {
					break
				}
				i++
			}
			tokens = append(tokens, input[start:i])
		}
	}
	return tokens
}

func scanQuoted(input string, start int) (string, int) {
	var b strings.Builder
	for i := start + 1; i < len(input); i++ {
		if input[i] == '\\' && i+1 < len(input) {
			b.WriteByte(input[i+1])
			i++
			continue
		}
		if input[i] == '"' {
			return b.String(), i + 1
		}
		b.WriteByte(input[i])
	}
	return b.String(), len(input)
}

func skipToLineEnd(input string, start int) int {
	for i := start; i < len(input); i++ {
		if input[i] == '\n' {
			return i + 1
		}
	}
	return len(input)
}
