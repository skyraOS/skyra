package codex

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"skyra-v1/src/skyra/skyra-dna-16/host"
	"strings"
)

// CodexLaunchPolicy is trusted kernel input for one possibility-scoped host.
// CodexHome contains only the authentication material and empty configuration
// selected for that possibility.
type CodexLaunchPolicy struct {
	Binary               string
	ExpectedVersion      string
	CodexHome            string
	WorkingDir           string
	ExecutableSearchPath []string
}

var codexContainmentOverrides = []string{
	`web_search="disabled"`,
	`features.shell_tool=false`,
	`features.unified_exec=false`,
	`features.view_image=false`,
	`features.image_generation=false`,
	`features.apps=false`,
	`apps._default.enabled=false`,
	`features.plugins=false`,
	`features.remote_plugin=false`,
	`features.browser_use=false`,
	`features.browser_use_external=false`,
	`features.browser_use_full_cdp_access=false`,
	`features.computer_use=false`,
	`features.multi_agent=false`,
	`agents.enabled=false`,
	`features.skill_mcp_dependency_install=false`,
	`history.persistence="none"`,
	`analytics.enabled=false`,
}

func (policy CodexLaunchPolicy) ProcessSpec(ctx context.Context) (host.ProcessSpec, error) {
	if ctx == nil || policy.Binary == "" || policy.ExpectedVersion == "" || policy.CodexHome == "" || policy.WorkingDir == "" || len(policy.ExecutableSearchPath) == 0 {
		return host.ProcessSpec{}, errors.New("complete Codex launch policy required")
	}
	for _, directory := range policy.ExecutableSearchPath {
		if !filepath.IsAbs(directory) || filepath.Clean(directory) != directory {
			return host.ProcessSpec{}, errors.New("Codex executable search path entries must be absolute canonical paths")
		}
	}
	for name, path := range map[string]string{"binary": policy.Binary, "Codex home": policy.CodexHome, "working directory": policy.WorkingDir} {
		if !filepath.IsAbs(path) || filepath.Clean(path) != path {
			return host.ProcessSpec{}, fmt.Errorf("%s must be an absolute canonical path", name)
		}
	}
	version, err := codexBinaryVersion(ctx, policy.Binary)
	if err != nil {
		return host.ProcessSpec{}, err
	}
	if version != policy.ExpectedVersion {
		return host.ProcessSpec{}, fmt.Errorf("Codex version %q rejected; expected %q", version, policy.ExpectedVersion)
	}
	args := []string{"--strict-config", "--ask-for-approval", "never", "--sandbox", "read-only"}
	for _, override := range codexContainmentOverrides {
		args = append(args, "-c", override)
	}
	args = append(args, "app-server", "--stdio")
	return host.ProcessSpec{
		Path: policy.Binary,
		Args: args,
		Dir:  policy.WorkingDir,
		Env: []string{
			"CODEX_HOME=" + policy.CodexHome,
			"PATH=" + strings.Join(policy.ExecutableSearchPath, string(filepath.ListSeparator)),
		},
	}, nil
}

func codexBinaryVersion(ctx context.Context, binary string) (string, error) {
	output, err := exec.CommandContext(ctx, binary, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("read Codex version: %w", err)
	}
	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", errors.New("Codex returned an empty version")
	}
	return version, nil
}

// StartCodexHost admits the pinned binary before creating live machinery.
func StartCodexHost(ctx context.Context, policy CodexLaunchPolicy) (*host.ProcessHost, error) {
	spec, err := policy.ProcessSpec(ctx)
	if err != nil {
		return nil, err
	}
	return host.StartProcessHost(ctx, spec)
}
