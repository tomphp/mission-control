package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/tomoram/mission-control/internal/config"
)

func runInstall(args []string) int {
	return runInstallOrUninstall(args, "install")
}

func runUninstall(args []string) int {
	return runInstallOrUninstall(args, "uninstall")
}

func runInstallOrUninstall(args []string, mode string) int {
	fs := flag.NewFlagSet(mode, flag.ContinueOnError)
	project := fs.Bool("project", false, "target ./.claude/settings.json instead of ~/.claude/settings.json")
	dryRun := fs.Bool("dry-run", false, "print what would change without writing")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	path, err := settingsPath(*project)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mission-control %s: %v\n", mode, err)
		return 1
	}

	binaryPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "mission-control %s: resolve binary path: %v\n", mode, err)
		return 1
	}

	settings, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mission-control %s: %v\n", mode, err)
		return 1
	}

	var updated config.RawSettings
	if mode == "install" {
		updated, err = config.Install(settings, binaryPath)
	} else {
		updated, err = config.Uninstall(settings, binaryPath)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "mission-control %s: %v\n", mode, err)
		return 1
	}

	if *dryRun {
		fmt.Printf("mission-control %s: dry run, would write %s\n", mode, path)
		return 0
	}

	if err := config.Save(path, updated); err != nil {
		fmt.Fprintf(os.Stderr, "mission-control %s: %v\n", mode, err)
		return 1
	}
	fmt.Printf("mission-control %s: updated %s\n", mode, path)
	return 0
}

func settingsPath(project bool) (string, error) {
	if project {
		return config.ProjectSettingsPath()
	}
	return config.GlobalSettingsPath()
}
