package cli

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"

	"github.com/go-git/go-git/v5/config"

	"github.com/prskr/git-age/core/ports"
)

type InstallCliHandler struct{}

func (h *InstallCliHandler) Run() (err error) {
	cfgPath, err := ports.GlobalGitConfigPath()
	if err != nil {
		return fmt.Errorf("failed to locate global git config: %w", err)
	}

	cfgFile, err := os.OpenFile(cfgPath, os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open global git config: %w", err)
	}

	defer func() {
		err = errors.Join(err, cfgFile.Close())
	}()

	cfg, err := config.ReadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to read global git config: %w", err)
	}

	if _, err := cfgFile.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to prepare config file for writing: %w", err)
	}

	filterSection := cfg.Raw.Section("filter")
	if filterSection.HasSubsection("age") {
		slog.Info("git-age already installed")
		return nil
	}

	ageSection := filterSection.Subsection("age")
	ageSection.SetOption("clean", "git-age clean -- %f")
	ageSection.SetOption("smudge", "git-age smudge -- %f")
	ageSection.SetOption("required", strconv.FormatBool(true))

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	cfgBytes, err := cfg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if _, err := cfgFile.Write(cfgBytes); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
