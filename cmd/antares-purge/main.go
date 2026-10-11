// Command antares-purge provisions a stock Antares install into a red-team
// configuration and can undo it. It is the antares analogue of dsh-purge:
// identity, standing orders, a skill pack, a gate plugin, roles and config,
// with install / uninstall / reapply / status / doctor lifecycle and backups.
package main

import (
	"fmt"
	"os"

	"github.com/enowdev/antares-purge/internal/deep"
	"github.com/enowdev/antares-purge/internal/install"
	"github.com/enowdev/antares-purge/internal/paths"
)

const usage = `antares-purge — provision Antares as a red-team install

usage:
  antares-purge install [--force] [--dry-run]   write the preset (backs up first)
  antares-purge uninstall [--dry-run]           restore the pre-install state
  antares-purge reapply                         re-write preset files, keep identity
  antares-purge status                          show what is installed
  antares-purge doctor                          check the install and show drift
  antares-purge deep [--tag <ref>] [--dry-run]  patch Antares's prompt and rebuild
  antares-purge help

The preset is written into the Antares home (%s).
`

func main() {
	args := os.Args[1:]
	cmd := "help"
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}
	var err error
	switch cmd {
	case "install":
		err = cmdInstall(args)
	case "uninstall":
		err = cmdUninstall(args)
	case "reapply":
		err = cmdReapply(args)
	case "status":
		err = cmdStatus(args)
	case "doctor":
		err = cmdDoctor(args)
	case "deep":
		err = cmdDeep(args)
	case "help", "-h", "--help":
		fmt.Printf(usage, paths.Home())
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		fmt.Printf(usage, paths.Home())
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func has(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func value(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func cmdInstall(args []string) error {
	opts := install.Options{DryRun: has(args, "--dry-run"), Force: has(args, "--force")}
	res, err := install.Run(opts)
	if err != nil {
		return err
	}
	verb := "wrote"
	if opts.DryRun {
		verb = "would write"
	}
	for _, w := range res.Written {
		fmt.Printf("  %s %s\n", verb, w)
	}
	for _, s := range res.Skipped {
		fmt.Printf("  kept  %s\n", s)
	}
	if opts.DryRun {
		fmt.Println("dry run — nothing written")
		return nil
	}
	fmt.Printf("installed. backup at %s\n", res.RunDir)
	fmt.Println("next: set your authorized scope in", paths.Path("purge-scope.txt"))
	fmt.Println("      then arm the gate: PURGE_MODE=enforce in the plugin env")
	return nil
}

func cmdUninstall(args []string) error {
	dry := has(args, "--dry-run")
	actions, err := install.Uninstall(dry)
	if err != nil {
		return err
	}
	for _, a := range actions {
		fmt.Println(" ", a)
	}
	if dry {
		fmt.Println("dry run — nothing changed")
	} else {
		fmt.Println("uninstalled.")
	}
	return nil
}

func cmdReapply(args []string) error {
	// Reapply refreshes the preset's own files (skills, roles, plugin, config)
	// but preserves the user's SOUL.md and AGENTS.md.
	opts := install.Options{DryRun: has(args, "--dry-run")}
	res, err := install.Run(opts)
	if err != nil {
		return err
	}
	for _, w := range res.Written {
		fmt.Println("  refreshed", w)
	}
	for _, s := range res.Skipped {
		fmt.Println("  kept     ", s)
	}
	fmt.Println("reapplied.")
	return nil
}

func cmdStatus(args []string) error {
	m, err := install.Status()
	if err != nil {
		return err
	}
	fmt.Printf("antares home: %s\n", paths.Home())
	if m.InstalledAt.IsZero() {
		fmt.Println("state:        not installed")
		return nil
	}
	fmt.Printf("state:        installed %s (preset v%s, mode %s)\n",
		m.InstalledAt.Format("2006-01-02 15:04:05Z"), m.Version, m.Mode)
	fmt.Printf("files:        %d tracked\n", len(m.Entries))
	return nil
}

func cmdDoctor(args []string) error {
	drift, err := install.Doctor()
	if err != nil {
		return err
	}
	if len(drift) == 0 {
		fmt.Println("ok — install matches the manifest, no drift")
		return nil
	}
	fmt.Println("drift detected:")
	for _, d := range drift {
		fmt.Println("  -", d)
	}
	fmt.Println("\nrun `antares-purge reapply` to refresh preset files, or")
	fmt.Println("`antares-purge uninstall` to restore the pre-install state.")
	return nil
}

func cmdDeep(args []string) error {
	cfg := deep.DefaultConfig()
	if tag := value(args, "--tag"); tag != "" {
		cfg.Tag = tag
	}
	cfg.DryRun = has(args, "--dry-run")
	log, err := deep.Apply(cfg)
	if err != nil {
		return err
	}
	for _, l := range log {
		fmt.Println(" ", l)
	}
	if cfg.DryRun {
		fmt.Println("dry run — nothing built")
	}
	return nil
}
