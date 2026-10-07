package instance

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
)

func prepareClaude(x machine.StartResident, root string) (*preparedResident, error) {
	s := x.Run.Surface
	if s.ResidentProfile != "darwin-claude-v1" {
		return nil, errors.New("unknown resident profile")
	}
	claudeLink, err := exec.LookPath("claude")
	if err != nil {
		return nil, errors.New("claude CLI not found in PATH")
	}
	claudePath, err := filepath.EvalSymlinks(claudeLink)
	if err != nil {
		return nil, err
	}
	claudeLinkDir, err := filepath.Abs(filepath.Dir(claudeLink))
	if err != nil {
		return nil, err
	}
	base := filepath.Join(filepath.Dir(root), "providers")
	if err = os.MkdirAll(base, 0700); err != nil {
		return nil, err
	}
	run, err := os.MkdirTemp(base, "claude-")
	if err != nil {
		return nil, err
	}
	run, err = filepath.EvalSymlinks(run)
	if err != nil {
		return nil, err
	}
	home, tmp, work := filepath.Join(run, "home"), filepath.Join(run, "tmp"), filepath.Join(run, "work")
	for _, dir := range []string{home, tmp, work} {
		if err = os.Mkdir(dir, 0700); err != nil {
			return nil, err
		}
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	auth, err := filepath.EvalSymlinks(filepath.Join(userHome, ".claude"))
	if err != nil {
		return nil, err
	}
	cacheOut, err := exec.Command("/usr/bin/getconf", "DARWIN_USER_CACHE_DIR").Output()
	if err != nil {
		return nil, err
	}
	cache, err := filepath.EvalSymlinks(strings.TrimSpace(string(cacheOut)))
	if err != nil {
		return nil, err
	}
	profile, err := horizon.RenderClaudeResident(x.Horizon, x.Floor, horizon.ClaudePaths{
		Home:          home,
		Scratch:       tmp,
		ClaudeBin:     claudePath,
		ClaudeLinkDir: claudeLinkDir,
		Auth:          auth,
		Cache:         cache,
		UserHome:      userHome,
	})
	if err != nil {
		return nil, err
	}
	if err = os.WriteFile(filepath.Join(run, "actor.sb"), []byte(profile), 0600); err != nil {
		return nil, err
	}
	user := os.Getenv("USER")
	if user == "" {
		user = filepath.Base(userHome)
	}
	p := &preparedResident{
		profile: profile,
		work:    work,
		env:     []string{"PATH=/usr/bin:/bin:/usr/local/bin:" + claudeLinkDir, "HOME=" + userHome, "TMPDIR=" + tmp, "USER=" + user},
	}
	return p, nil
}
