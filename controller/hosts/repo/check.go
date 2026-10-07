package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type checkInput struct {
	Action  string `json:"action"`
	Package string `json:"package,omitempty"`
}
type checkResult struct {
	Action     string `json:"action"`
	ExitCode   int    `json:"exit_code"`
	Output     string `json:"output,omitempty"`
	Truncated  bool   `json:"truncated,omitempty"`
	Error      string `json:"error,omitempty"`
	ReportPath string `json:"report_path,omitempty"`
	DurationMS int64  `json:"duration_ms"`
}

func checkArgs(in checkInput) (string, []string, error) {
	if in.Package != "" && in.Action != "go.test" && in.Action != "go.build" {
		return "", nil, fmt.Errorf("package is only available for Go actions")
	}
	p := in.Package
	if p == "" {
		p = "./controller/..."
	}
	if !strings.HasPrefix(p, "./controller/") || strings.Contains(strings.TrimSuffix(p, "/..."), "..") || strings.ContainsAny(p, "\\\x00 \t\n:;") {
		return "", nil, fmt.Errorf("package must be a repository-relative controller package or ./controller/...")
	}
	switch in.Action {
	case "go.test":
		return "go", []string{"test", "-count=1", "-timeout=120s", p}, nil
	case "go.build":
		return "go", []string{"build", p}, nil
	case "go.format":
		return "gofmt", []string{"-l", "controller"}, nil
	case "swift.test":
		platform := "/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer"
		return "swift", []string{"test", "--disable-sandbox", "--enable-swift-testing", "--package-path", "desktop",
			"-Xswiftc", "-F", "-Xswiftc", platform + "/Library/Frameworks",
			"-Xswiftc", "-I", "-Xswiftc", platform + "/usr/lib",
			"-Xlinker", "-L", "-Xlinker", platform + "/usr/lib",
			"-Xlinker", "-rpath", "-Xlinker", platform + "/usr/lib",
			"-Xlinker", "-rpath", "-Xlinker", platform + "/Library/Frameworks"}, nil
	case "swift.build":
		return "swift", []string{"build", "--disable-sandbox", "--package-path", "desktop", "-c", "release"}, nil
	case "python.test":
		return "python", []string{"-m", "unittest", "discover", "-s", "controller/blender", "-p", "test_*.py"}, nil
	default:
		return "", nil, fmt.Errorf("unknown action; use go.test, go.build, go.format, swift.test, swift.build or python.test")
	}
}

// Copy only build inputs, without following links or importing scratch code,
// repository metadata, caches or existing executable output.
func snapshot(source, dest string) error {
	root, err := os.OpenRoot(source)
	if err != nil {
		return err
	}
	defer root.Close()
	var total int64
	for _, name := range []string{"go.mod", "go.sum", "controller", "desktop", "docs", "skyra"} {
		start := filepath.Join(source, name)
		if _, err := os.Lstat(start); os.IsNotExist(err) {
			continue
		}
		err := filepath.WalkDir(start, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && (d.Name() == ".git" || d.Name() == ".skyra" || d.Name() == ".build" || d.Name() == ".swiftpm" || d.Name() == "__pycache__" || d.Name() == "node_modules" || d.Name() == ".venv" || d.Name() == "bin" || d.Name() == "build") {
				return filepath.SkipDir
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink build input refused: %s", path)
			}
			rel, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}
			out := filepath.Join(dest, rel)
			if d.IsDir() {
				return os.MkdirAll(out, 0755)
			}
			st, err := d.Info()
			if err != nil {
				return err
			}
			if !st.Mode().IsRegular() {
				return fmt.Errorf("nonregular build input")
			}
			total += st.Size()
			if total > 128<<20 {
				return fmt.Errorf("build inputs exceed 128 MiB")
			}
			// O_NOFOLLOW also rejects a file swapped to a link after WalkDir.
			input, err := root.OpenFile(rel, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
			if err != nil {
				return err
			}
			defer input.Close()
			st, err = input.Stat()
			if err != nil || !st.Mode().IsRegular() {
				return fmt.Errorf("build input changed")
			}
			output, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, st.Mode().Perm()&0755)
			if err != nil {
				return err
			}
			defer output.Close()
			_, err = io.Copy(output, io.LimitReader(input, 128<<20))
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func checkProfile(job string, reads []string, swiftReplacement bool) string {
	quote := func(s string) string { b, _ := json.Marshal(s); return string(b) }
	rules := []string{"(version 1)", "(deny default)", "(allow process-fork)", "(allow process-exec)", "(allow signal (target same-sandbox))", "(allow sysctl-read)", "(allow file-read-metadata)", "(allow file-map-executable)", "(allow mach-lookup)", "(allow iokit-open)", "(allow ipc-posix-shm)", "(allow ipc-posix-sem)"}
	for _, p := range append([]string{job, "/System", "/usr", "/bin", "/sbin", "/Library/Developer", "/Library/Frameworks", "/Library/Apple", "/private/etc", "/private/var/db", "/private/var/run/com.apple.security.cryptexd"}, reads...) {
		if p == "" {
			continue
		}
		rules = append(rules, "(allow file-read* (subpath "+quote(p)+"))")
	}
	rules = append(rules, "(allow file-read* (literal \"/\") (literal \"/dev/null\") (literal \"/dev/random\") (literal \"/dev/urandom\"))", "(allow file-write* (subpath "+quote(job)+") (literal \"/dev/null\"))",
		// Tests may start local fixtures; internet and external Unix sockets are denied.
		"(allow network-bind (local ip \"localhost:*\") (local unix-socket (subpath "+quote(job)+")))", "(allow network-inbound (local ip \"localhost:*\"))", "(allow network-outbound (remote ip \"localhost:*\") (remote unix-socket (subpath "+quote(job)+")))")
	rules = append(rules, "(allow file-read* (literal \"/Library/Preferences/com.apple.dt.Xcode.plist\"))")
	// Foundation's atomic writes use the OS replacement directory rather than
	// TMPDIR. Grant only Swift build/test replacement folders, not the pool.
	if swiftReplacement {
		pool, _ := filepath.EvalSymlinks(os.TempDir())
		pattern := "^" + regexp.QuoteMeta(filepath.Join(pool, "TemporaryItems")) + "/NSIRD_swift-(build|test|driver)_[A-Za-z0-9]+(/.*)?$"
		rules = append(rules, "(allow file-read* file-write* (regex "+quote(pattern)+"))")
		rules = append(rules, "(allow file-read* (literal "+quote(filepath.Join(pool, "xcrun_db"))+") )")
		cachePattern := "^" + regexp.QuoteMeta(pool) + "/xcrun_db-[A-Za-z0-9]+$"
		rules = append(rules, "(allow file-read* file-write* (regex "+quote(cachePattern)+"))")
		old, _ := filepath.Glob(filepath.Join(pool, "TemporaryItems", "NSIRD_swift-*"))
		caches, _ := filepath.Glob(filepath.Join(pool, "xcrun_db-*"))
		old = append(old, caches...)
		for _, path := range old {
			rules = append(rules, "(deny file-read* file-write* (subpath "+quote(path)+"))")
		}
	}
	return strings.Join(rules, "\n")
}

func (g repository) check(in checkInput) checkResult {
	r := checkResult{Action: in.Action, ExitCode: -1}
	start := time.Now()
	tool, args, err := checkArgs(in)
	if err != nil {
		r.Error = "invalid_request: " + err.Error()
		return r
	}
	tools := map[string]string{"go": filepath.Join(g.goRoot, "bin/go"), "gofmt": filepath.Join(g.goRoot, "bin/gofmt"), "swift": "/Applications/Xcode.app/Contents/Developer/Toolchains/XcodeDefault.xctoolchain/usr/bin/swift", "python": g.python}
	binary := tools[tool]
	if binary == "" {
		r.Error = "tool_unavailable"
		return r
	}
	if _, err = os.Stat(binary); err != nil {
		r.Error = "tool_unavailable: " + tool
		return r
	}
	job, err := os.MkdirTemp("", "skyra-check-")
	if err != nil {
		r.Error = err.Error()
		return r
	}
	job, err = filepath.EvalSymlinks(job)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	defer os.RemoveAll(job)
	workspace := filepath.Join(job, "workspace")
	os.MkdirAll(workspace, 0700)
	if err = snapshot(g.path, workspace); err != nil {
		r.Error = "snapshot_failed: " + err.Error()
		return r
	}
	for _, d := range []string{"home", "tmp", "cache", "modcache"} {
		os.MkdirAll(filepath.Join(job, d), 0700)
	}
	profile := filepath.Join(job, "check.sb")
	reads := []string{g.goRoot, g.moduleCache, "/Applications/Xcode.app/Contents"}
	if g.python != "" {
		reads = append(reads, filepath.Dir(filepath.Dir(g.python)))
	}
	if err = os.WriteFile(profile, []byte(checkProfile(job, reads, tool == "swift")), 0600); err != nil {
		r.Error = err.Error()
		return r
	}
	if tool == "swift" {
		args = append(args, "--cache-path", filepath.Join(job, "cache"), "--config-path", filepath.Join(job, "home/config"), "--security-path", filepath.Join(job, "home/security"), "--disable-keychain", "--disable-netrc", "--manifest-cache", "none")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", append([]string{"-f", profile, binary}, args...)...)
	cmd.Dir = workspace
	cmd.Env = []string{"PATH=" + filepath.Join(g.goRoot, "bin") + ":/usr/bin:/bin:/usr/sbin:/sbin", "HOME=" + filepath.Join(job, "home"), "TMPDIR=" + filepath.Join(job, "tmp"), "GOCACHE=" + filepath.Join(job, "cache"), "GOMODCACHE=" + g.moduleCache, "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "GOTELEMETRY=off", "CGO_ENABLED=0", "PYTHONDONTWRITEBYTECODE=1", "DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer"}
	cmd.Env = append(cmd.Env, "CLANG_MODULE_CACHE_PATH="+filepath.Join(job, "cache"), "SWIFTPM_MODULECACHE_OVERRIDE="+filepath.Join(job, "cache"), "XDG_CACHE_HOME="+filepath.Join(job, "cache"))
	cmd.Env = append(cmd.Env, "SDKROOT=/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 2 * time.Second
	var output boundedOutput
	cmd.Stdout = &output
	cmd.Stderr = &output
	err = cmd.Run()
	r.Output = string(output.data)
	r.Truncated = output.truncated
	r.DurationMS = time.Since(start).Milliseconds()
	if err == nil {
		r.ExitCode = 0
		if in.Action == "go.format" && strings.TrimSpace(r.Output) != "" {
			r.ExitCode = 1
			r.Error = "formatting_required"
		}
	} else {
		r.Error = "check_failed: " + err.Error()
		if exit, ok := err.(*exec.ExitError); ok {
			r.ExitCode = exit.ExitCode()
		}
		if ctx.Err() != nil {
			r.Error = "check_timeout"
		}
	}
	// Reports alone are published; generated executables stay in the disposable job.
	name := filepath.Base(job) + ".json"
	rel := filepath.Join("ideas", "checks", name)
	root, e := os.OpenRoot(g.path)
	if e != nil {
		r.Error += "; report_failed: " + e.Error()
		return r
	}
	defer root.Close()
	if e = root.MkdirAll("ideas/checks", 0755); e == nil {
		var f *os.File
		f, e = root.OpenFile(rel, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0644)
		if e == nil {
			r.ReportPath = rel
			data, _ := json.MarshalIndent(r, "", "  ")
			_, e = f.Write(append(data, '\n'))
			f.Close()
		}
	}
	if e != nil {
		r.Error += "; report_failed: " + e.Error()
	}
	return r
}
