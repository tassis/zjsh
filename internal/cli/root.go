package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tassis/zjsh/internal/domain"
	"github.com/tassis/zjsh/internal/platform"
	configprovider "github.com/tassis/zjsh/internal/provider/config"
	"github.com/tassis/zjsh/internal/provider/zellij"
	"github.com/tassis/zjsh/internal/provider/zoxide"
	"github.com/tassis/zjsh/internal/service"
	"github.com/tassis/zjsh/internal/version"
	"github.com/tassis/zjsh/internal/view"
)

type App struct {
	Stdout     io.Writer
	Stderr     io.Writer
	Runner     platform.Runner
	Aggregator service.Aggregator
	Launcher   service.Launcher
}

func NewApp(stdout, stderr io.Writer) App {
	runner := platform.ExecRunner{}
	return App{
		Stdout: stdout,
		Stderr: stderr,
		Runner: runner,
		Aggregator: service.Aggregator{
			Config: configprovider.Loader{},
			Zellij: zellij.Provider{Runner: runner},
			Zoxide: zoxide.Provider{Runner: runner},
		},
		Launcher: service.Launcher{Runner: runner, Env: platform.OSEnv{}},
	}
}

func Main(args []string, stdout, stderr io.Writer) int {
	return NewApp(stdout, stderr).Run(context.Background(), args)
}

func (a App) Run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		a.printUsage()
		return 1
	}
	switch args[0] {
	case "version", "--version", "-version":
		if err := a.runVersion(args[1:]); err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "error: %v\n", err)
			return 1
		}
		return 0
	case "list":
		if err := a.runList(ctx, args[1:]); err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "error: %v\n", err)
			return 1
		}
		return 0
	case "macros":
		if err := a.runMacros(ctx, args[1:]); err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "error: %v\n", err)
			return 1
		}
		return 0
	case "connect":
		if err := a.runConnect(ctx, args[1:]); err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "error: %v\n", err)
			return 1
		}
		return 0
	case "exec":
		if err := a.runExec(ctx, args[1:]); err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "error: %v\n", err)
			return 1
		}
		return 0
	case "doctor", "config":
		if args[0] == "doctor" {
			if err := a.runDoctor(ctx, args[1:]); err != nil {
				_, _ = fmt.Fprintf(a.Stderr, "error: %v\n", err)
				return 1
			}
			return 0
		}
		if err := a.runConfig(ctx, args[1:]); err != nil {
			_, _ = fmt.Fprintf(a.Stderr, "error: %v\n", err)
			return 1
		}
		return 0
	default:
		a.printUsage()
		return 1
	}
}

func (a App) runVersion(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: zjsh version")
	}
	_, err := fmt.Fprintln(a.Stdout, version.String())
	return err
}

func (a App) runList(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	jsonOutput := fs.Bool("json", false, "output entries as JSON")
	iconsOutput := fs.Bool("i", false, "show selector icons")
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := a.Aggregator.List(ctx, true)
	if err != nil {
		return err
	}
	if *jsonOutput {
		return view.WriteJSON(a.Stdout, result.Entries)
	}
	if *iconsOutput {
		return view.WriteLabels(a.Stdout, result.Entries, result.Config.Defaults.Icons)
	}
	return view.WritePlain(a.Stdout, result.Entries)
}

func (a App) runConnect(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("connect", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: zjsh connect <target>")
	}
	target := strings.TrimSpace(fs.Arg(0))
	result, err := a.Aggregator.List(ctx, true)
	if err != nil {
		return err
	}
	target = normalizeConnectTarget(target, result.Config.Defaults.Icons)
	entries := result.Entries
	entry, err := service.ResolveExact(entries, target)
	if err != nil {
		return err
	}
	entry = service.PrepareConnectEntry(entry, entries)
	return a.Launcher.Connect(ctx, entry)
}

func (a App) runMacros(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("macros", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	jsonOutput := fs.Bool("json", false, "output macros as JSON")
	iconsOutput := fs.Bool("i", false, "show selector icons")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: zjsh macros")
	}
	config, err := a.Aggregator.Config.Load(ctx)
	if err != nil {
		return err
	}
	if *jsonOutput {
		return view.WriteMacroJSON(a.Stdout, config.Macros)
	}
	if *iconsOutput {
		return view.WriteMacroLabels(a.Stdout, config.Macros, config.Defaults.Icons)
	}
	return view.WriteMacroPlain(a.Stdout, config.Macros)
}

func (a App) runExec(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("exec", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: zjsh exec <macro>")
	}
	config, err := a.Aggregator.Config.Load(ctx)
	if err != nil {
		return err
	}
	target := normalizeMacroTarget(fs.Arg(0), config.Defaults.Icons)
	macro, err := service.ResolveMacro(config.Macros, target)
	if err != nil {
		return err
	}
	return a.Launcher.ExecMacro(ctx, macro)
}

func (a App) printUsage() {
	_, _ = fmt.Fprintln(a.Stderr, "usage: zjsh <command>")
	_, _ = fmt.Fprintln(a.Stderr, "commands:")
	_, _ = fmt.Fprintln(a.Stderr, "  list      list aggregated entries")
	_, _ = fmt.Fprintln(a.Stderr, "  macros    list configured macros")
	_, _ = fmt.Fprintln(a.Stderr, "  connect   connect to a session or project")
	_, _ = fmt.Fprintln(a.Stderr, "  exec      execute a macro in zellij")
	_, _ = fmt.Fprintln(a.Stderr, "  version   print version")
	_, _ = fmt.Fprintln(a.Stderr, "  doctor    validate dependencies and config")
	_, _ = fmt.Fprintln(a.Stderr, "  config    manage config scaffolding")
}

func (a App) runDoctor(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: zjsh doctor")
	}
	checks := make([]doctorCheck, 0, 8)

	// zellij is the only required runtime dependency. Without it, zjsh
	// cannot list, switch, attach, create, or resurrect sessions.
	checks = append(checks, a.binaryCheck("zellij"))

	// zoxide is an optional list source. Missing zoxide should be visible
	// to the user, but it must not make doctor fail.
	checks = append(checks, a.optionalBinaryCheck("zoxide"))

	// A missing config file is valid: users can still use live sessions,
	// resurrectable sessions, zoxide paths, and the built-in `.` entry.
	configPath, err := a.Aggregator.Config.ResolvedPath()
	if err != nil {
		return err
	}
	info, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			checks = append(checks, doctorCheck{Status: doctorWarn, Label: "config", Detail: fmt.Sprintf("missing: %s", configPath)})
			return writeDoctorReport(a.Stdout, checks)
		}
		return err
	}
	if info.IsDir() {
		checks = append(checks, doctorCheck{Status: doctorFail, Label: "config", Detail: fmt.Sprintf("expected file, got directory: %s", configPath)})
		return writeDoctorReport(a.Stdout, checks)
	}
	checks = append(checks, doctorCheck{Status: doctorOK, Label: "config", Detail: configPath})

	// If the config file exists, it must be parseable and semantically valid.
	// This catches invalid KDL, invalid required values, and invalid project
	// path modes such as `path` plus `cwd true`.
	config, err := a.Aggregator.Config.Load(ctx)
	if err != nil {
		checks = append(checks, doctorCheck{Status: doctorFail, Label: "config parse", Detail: err.Error()})
		return writeDoctorReport(a.Stdout, checks)
	}
	checks = append(checks, doctorCheck{Status: doctorOK, Label: "config parse", Detail: "ok"})
	for _, project := range config.Projects {
		// `cwd true` projects use the runtime current directory, so there is no
		// static project path to check during doctor.
		if project.CWD {
			checks = append(checks, doctorCheck{Status: doctorOK, Label: "project path", Detail: fmt.Sprintf("%s: runtime cwd", project.Name)})
		} else {
			// Static project paths should expand successfully and exist on disk.
			path, expandErr := platform.ExpandHome(project.Path)
			if expandErr != nil {
				checks = append(checks, doctorCheck{Status: doctorFail, Label: "project path", Detail: fmt.Sprintf("%s: %v", project.Name, expandErr)})
				continue
			}
			if _, statErr := os.Stat(path); statErr != nil {
				checks = append(checks, doctorCheck{Status: doctorFail, Label: "project path", Detail: fmt.Sprintf("%s: %s", project.Name, path)})
			} else {
				checks = append(checks, doctorCheck{Status: doctorOK, Label: "project path", Detail: fmt.Sprintf("%s: %s", project.Name, path)})
			}
		}
		if project.LayoutFile == "" {
			continue
		}

		// layout_file is optional, but when configured it must expand and exist.
		layoutPath, expandErr := platform.ExpandHome(project.LayoutFile)
		if expandErr != nil {
			checks = append(checks, doctorCheck{Status: doctorFail, Label: "layout file", Detail: fmt.Sprintf("%s: %v", project.Name, expandErr)})
			continue
		}
		if _, statErr := os.Stat(layoutPath); statErr != nil {
			checks = append(checks, doctorCheck{Status: doctorFail, Label: "layout file", Detail: fmt.Sprintf("%s: %s", project.Name, layoutPath)})
		} else {
			checks = append(checks, doctorCheck{Status: doctorOK, Label: "layout file", Detail: fmt.Sprintf("%s: %s", project.Name, layoutPath)})
		}
	}
	return writeDoctorReport(a.Stdout, checks)
}

func (a App) binaryCheck(name string) doctorCheck {
	path, err := a.Runner.LookPath(name)
	if err != nil {
		return doctorCheck{Status: doctorFail, Label: name, Detail: err.Error()}
	}
	return doctorCheck{Status: doctorOK, Label: name, Detail: path}
}

func (a App) optionalBinaryCheck(name string) doctorCheck {
	path, err := a.Runner.LookPath(name)
	if err != nil {
		return doctorCheck{Status: doctorWarn, Label: name, Detail: err.Error() + " (optional)"}
	}
	return doctorCheck{Status: doctorOK, Label: name, Detail: path + " (optional)"}
}

func (a App) runConfig(_ context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: zjsh config init")
	}
	switch args[0] {
	case "init":
		return a.runConfigInit(args[1:])
	default:
		return fmt.Errorf("unknown config command %q", args[0])
	}
}

func (a App) runConfigInit(args []string) error {
	fs := flag.NewFlagSet("config init", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	pathFlag := fs.String("path", "", "write config to this path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: zjsh config init [--path <file>]")
	}
	loader := a.Aggregator.Config
	if *pathFlag != "" {
		loader.Path = *pathFlag
	}
	path, err := loader.ResolvedPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config already exists: %s", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(sampleConfigForGOOS(runtime.GOOS)), 0o644); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(a.Stdout, "wrote %s\n", path)
	return nil
}

type doctorStatus string

const (
	doctorOK   doctorStatus = "OK"
	doctorWarn doctorStatus = "WARN"
	doctorFail doctorStatus = "FAIL"
)

type doctorCheck struct {
	Status doctorStatus
	Label  string
	Detail string
}

func writeDoctorReport(w io.Writer, checks []doctorCheck) error {
	failures := 0
	warnings := 0
	for _, check := range checks {
		if _, err := fmt.Fprintf(w, "[%s] %s: %s\n", check.Status, check.Label, check.Detail); err != nil {
			return err
		}
		switch check.Status {
		case doctorFail:
			failures++
		case doctorWarn:
			warnings++
		}
	}
	_, err := fmt.Fprintf(w, "summary: %d fail, %d warn\n", failures, warnings)
	if err != nil {
		return err
	}
	if failures > 0 {
		return fmt.Errorf("doctor found %d failing checks", failures)
	}
	return nil
}

func sampleConfigForGOOS(goos string) string {
	if goos == "windows" {
		return strings.TrimSpace(`defaults {
  shell "sh"
  restart_on_resurrection false
  icon_project "◆"
  icon_session "●"
  icon_resurrectable "↺"
  icon_path "→"
}

project "api" {
  path "C:\\Users\\example\\work\\api"
  session "api"
}

project "infra" {
  path "C:\\Users\\example\\work\\infra"
  layout "compact"
}

project "ops" {
  path "C:\\Users\\example\\work\\ops"
  layout_file "C:\\Users\\example\\AppData\\Roaming\\zellij\\layouts\\ops.kdl"
}
`) + "\n"
	}
	return strings.TrimSpace(`defaults {
  shell "sh"
  restart_on_resurrection false
  icon_project "◆"
  icon_session "●"
  icon_resurrectable "↺"
  icon_path "→"
}

project "api" {
  path "/Users/example/work/api"
  session "api"
  restart_on_resurrection true
}

project "infra" {
  path "/Users/example/work/infra"
  layout "compact"
}

project "ops" {
  path "/Users/example/work/ops"
  layout_file "/Users/example/.config/zellij/layouts/ops.kdl"
}
`) + "\n"
}

func normalizeConnectTarget(value string, icons domain.Icons) string {
	value = strings.TrimSpace(value)
	for _, prefix := range connectLabelPrefixes(icons) {
		if strings.HasPrefix(value, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(value, prefix))
		}
	}
	return value
}

func normalizeMacroTarget(value string, icons domain.Icons) string {
	value = strings.TrimSpace(value)
	for _, prefix := range macroLabelPrefixes(icons) {
		if strings.HasPrefix(value, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(value, prefix))
		}
	}
	return value
}

func connectLabelPrefixes(icons domain.Icons) []string {
	defaultIcons := domain.DefaultIcons()
	values := []string{
		defaultIcons.Project,
		defaultIcons.Session,
		defaultIcons.Resurrectable,
		defaultIcons.Path,
		icons.Project,
		icons.Session,
		icons.Resurrectable,
		icons.Path,
		"•",
	}
	prefixes := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		if value == "" {
			continue
		}
		prefix := value + " "
		if _, ok := seen[prefix]; ok {
			continue
		}
		seen[prefix] = struct{}{}
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}

func macroLabelPrefixes(icons domain.Icons) []string {
	defaultIcons := domain.DefaultIcons()
	values := []string{defaultIcons.Macro, icons.Macro}
	prefixes := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		if value == "" {
			continue
		}
		prefix := value + " "
		if _, ok := seen[prefix]; ok {
			continue
		}
		seen[prefix] = struct{}{}
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}
