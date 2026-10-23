// Command claude-keychain is the one way a sandboxed Claude CLI reaches its
// login (proposal 47). The CLI reads and refreshes its OAuth credentials by
// running `security`. Inside the sandbox, `security` is a copy of this
// program, which forwards only those requests over a socket. The broker on the
// other end is this program too, started by the runtime outside the sandbox:
// it serves exactly one generic-password item, the service the surface pins,
// and nothing else in the user's keychain. The sandbox is granted neither
// /usr/bin/security nor a read of ~/Library/Keychains.
//
//	claude-keychain serve -socket S -service SVC -watch-fd 3   (the broker)
//	security find-generic-password -a A -w -s SVC               (the shim)
//
// The broker exits when the watch descriptor reaches EOF, which is when the
// resident holding its other end exits. It never logs a secret.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
)

// security is the real tool the broker runs; a variable so tests can stand
// in for it.
var security = "/usr/bin/security"

// Exit codes the CLI distinguishes: an item that is not there, and a keychain
// that is locked.
const (
	notFound = 44
	refused  = 1
)

// stdinLimit is the longest command `security -i` reads from stdin; the CLI
// itself passes longer payloads as arguments instead.
const stdinLimit = 4032

type request struct {
	Op      string `json:"op"` // "find" or "add"
	Service string `json:"service"`
	Hex     string `json:"hex,omitempty"`
}

type response struct {
	Code int    `json:"code"`
	Out  string `json:"out,omitempty"`
}

func main() {
	if filepath.Base(os.Args[0]) == "security" {
		os.Exit(shim(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, dial(os.Getenv("SKYRA_KEYCHAIN_SOCKET"))))
	}
	if len(os.Args) < 2 || os.Args[1] != "serve" {
		fmt.Fprintln(os.Stderr, "usage: claude-keychain serve -socket S -service SVC -watch-fd N")
		os.Exit(2)
	}
	flags := flag.NewFlagSet("serve", flag.ExitOnError)
	socket := flags.String("socket", "", "unix socket to serve")
	service := flags.String("service", "", "the one generic-password service served")
	watch := flags.Int("watch-fd", -1, "descriptor whose EOF ends the broker")
	flags.Parse(os.Args[2:])
	if err := serve(*socket, *service, *watch); err != nil {
		fmt.Fprintln(os.Stderr, "claude-keychain:", err)
		os.Exit(1)
	}
}

// broker serves one item of the login keychain for one account.
type broker struct {
	service, account string
	mu               sync.Mutex // one security call at a time
}

func newBroker(service string) (*broker, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	if !quotable(service) || !quotable(u.Username) {
		return nil, errors.New("service and account must be plain names")
	}
	return &broker{service: service, account: u.Username}, nil
}

func serve(socket, service string, watch int) error {
	if socket == "" || watch < 0 {
		return errors.New("a socket and a watch descriptor are required")
	}
	b, err := newBroker(service)
	if err != nil {
		return err
	}
	os.Remove(socket)
	l, err := net.Listen("unix", socket)
	if err != nil {
		return err
	}
	defer os.Remove(socket)
	if err = os.Chmod(socket, 0o600); err != nil {
		return err
	}
	go func() {
		io.Copy(io.Discard, os.NewFile(uintptr(watch), "watch"))
		l.Close()
	}()
	for {
		c, err := l.Accept()
		if err != nil {
			return nil // the watch closed the listener
		}
		go func() {
			defer c.Close()
			var req request
			if json.NewDecoder(io.LimitReader(c, 64<<10)).Decode(&req) != nil {
				return
			}
			json.NewEncoder(c).Encode(b.answer(req))
		}()
	}
}

// answer runs one request against the real keychain. A request for any other
// service is answered as if that item did not exist.
func (b *broker) answer(req request) response {
	if req.Service != b.service {
		return response{Code: notFound}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	switch req.Op {
	case "find":
		out, code := run(exec.Command(security, "find-generic-password", "-a", b.account, "-w", "-s", b.service))
		return response{Code: code, Out: out}
	case "add":
		if !isHex(req.Hex) {
			return response{Code: refused}
		}
		line := `add-generic-password -U -a "` + b.account + `" -s "` + b.service + `" -X "` + req.Hex + "\"\n"
		if len(line) <= stdinLimit {
			cmd := exec.Command(security, "-i")
			cmd.Stdin = strings.NewReader(line)
			_, code := run(cmd)
			return response{Code: code}
		}
		_, code := run(exec.Command(security, "add-generic-password", "-U", "-a", b.account, "-s", b.service, "-X", req.Hex))
		return response{Code: code}
	}
	return response{Code: refused}
}

// run returns a command's stdout and exit code; its stderr is dropped, since
// it can echo what it was given.
func run(cmd *exec.Cmd) (string, int) {
	out, err := cmd.Output()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return "", exit.ExitCode()
	}
	if err != nil {
		return "", refused
	}
	return string(out), 0
}

// dial returns how the shim reaches the broker.
func dial(socket string) func(request) (response, error) {
	return func(req request) (response, error) {
		if socket == "" {
			return response{}, errors.New("no keychain broker")
		}
		c, err := net.Dial("unix", socket)
		if err != nil {
			return response{}, err
		}
		defer c.Close()
		if err = json.NewEncoder(c).Encode(req); err != nil {
			return response{}, err
		}
		var resp response
		err = json.NewDecoder(c).Decode(&resp)
		return resp, err
	}
}

// shim answers the `security` commands the Claude CLI runs, and refuses every
// other one: it can read and update its own login, not delete it, and not
// reach any other item.
func shim(args []string, stdin io.Reader, stdout, stderr io.Writer, call func(request) (response, error)) int {
	if len(args) == 0 {
		return refuse(stderr, "")
	}
	switch args[0] {
	case "show-keychain-info":
		return 0 // unlocked: the broker answers for the real keychain
	case "find-generic-password":
		f, ok := parse(args[1:], map[string]bool{"-a": true, "-s": true}, map[string]bool{"-w": true})
		if !ok || f["-s"] == "" {
			return refuse(stderr, args[0])
		}
		resp, err := call(request{Op: "find", Service: f["-s"]})
		if err != nil {
			fmt.Fprintln(stderr, "security: keychain broker unavailable")
			return refused
		}
		if resp.Code == 0 && f["-w"] != "" {
			io.WriteString(stdout, resp.Out)
		}
		return resp.Code
	case "add-generic-password":
		return add(args[1:], stderr, call)
	case "-i":
		b, err := io.ReadAll(io.LimitReader(stdin, 1<<20))
		if err != nil {
			return refused
		}
		code := 0
		for _, line := range strings.Split(string(b), "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			words, ok := split(line)
			if !ok || len(words) == 0 || words[0] != "add-generic-password" {
				return refuse(stderr, "interactive command")
			}
			if c := add(words[1:], stderr, call); c != 0 {
				code = c
			}
		}
		return code
	}
	return refuse(stderr, args[0])
}

func add(args []string, stderr io.Writer, call func(request) (response, error)) int {
	f, ok := parse(args, map[string]bool{"-a": true, "-s": true, "-X": true}, map[string]bool{"-U": true})
	if !ok || f["-s"] == "" || f["-U"] == "" || !isHex(f["-X"]) {
		return refuse(stderr, "add-generic-password")
	}
	resp, err := call(request{Op: "add", Service: f["-s"], Hex: f["-X"]})
	if err != nil {
		fmt.Fprintln(stderr, "security: keychain broker unavailable")
		return refused
	}
	return resp.Code
}

func refuse(stderr io.Writer, what string) int {
	fmt.Fprintf(stderr, "security: %s is not available in this sandbox\n", strings.TrimSpace(what+" command"))
	return refused
}

// parse reads flags that take a value and flags that do not; anything else is
// refused. A boolean flag that is present maps to "1".
func parse(args []string, valued, boolean map[string]bool) (map[string]string, bool) {
	f := map[string]string{}
	for i := 0; i < len(args); i++ {
		switch {
		case boolean[args[i]]:
			f[args[i]] = "1"
		case valued[args[i]] && i+1 < len(args):
			f[args[i]] = args[i+1]
			i++
		default:
			return nil, false
		}
	}
	return f, true
}

// split tokenizes one `security -i` line: words separated by spaces, a
// double-quoted word taken whole.
func split(line string) ([]string, bool) {
	var words []string
	r := bufio.NewReader(strings.NewReader(strings.TrimSpace(line)))
	var word strings.Builder
	in, quoted := false, false
	for {
		c, _, err := r.ReadRune()
		if err != nil {
			break
		}
		switch {
		case c == '"':
			quoted = !quoted
			in = true
		case c == ' ' && !quoted:
			if in {
				words = append(words, word.String())
				word.Reset()
				in = false
			}
		default:
			word.WriteRune(c)
			in = true
		}
	}
	if quoted {
		return nil, false
	}
	if in {
		words = append(words, word.String())
	}
	return words, true
}

func isHex(s string) bool {
	if s == "" || len(s)%2 != 0 {
		return false
	}
	return strings.Trim(s, "0123456789abcdefABCDEF") == ""
}

// quotable is a name that can be written inside a `security -i` quoted word.
func quotable(s string) bool {
	return s != "" && !strings.ContainsAny(s, "\"\\\n\r\x00")
}
