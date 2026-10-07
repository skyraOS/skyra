package observation

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// ReadLine requires a complete, bounded frame, including a terminating newline.
func ReadLine(r io.Reader, limit int) ([]byte, error) {
	b, err := bufio.NewReader(io.LimitReader(r, int64(limit+1))).ReadBytes('\n')
	if len(b) > limit {
		return nil, errors.New("response exceeds frame limit")
	}
	if err != nil {
		return nil, fmt.Errorf("incomplete frame: %w", err)
	}
	return b, nil
}

// Client keeps identity checks and the observation on the same connection.
// It exposes no general-purpose daemon method or tool dispatch operation.
type Client struct{ conn net.Conn }

func Dial(ctx context.Context, path string) (*Client, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, "unix", path)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(15 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return nil, err
	}
	return &Client{conn: conn}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

func (c *Client) exchange(method string, args map[string]any) (json.RawMessage, error) {
	request := map[string]any{"method": method}
	if args != nil {
		request["name"], request["args"] = "get_window_state", args
	}
	if err := json.NewEncoder(c.conn).Encode(request); err != nil {
		return nil, err
	}
	b, err := ReadLine(c.conn, MaxResponse)
	if err != nil {
		return nil, err
	}
	var response struct {
		OK     bool            `json:"ok"`
		Result json.RawMessage `json:"result"`
		Error  string          `json:"error"`
	}
	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, fmt.Errorf("driver: %s", response.Error)
	}
	if len(response.Result) == 0 || string(response.Result) == "null" {
		return nil, errors.New("driver omitted result")
	}
	return response.Result, nil
}

func (c *Client) Verify(s Scope) error {
	if err := s.Validate(); err != nil {
		return err
	}
	b, err := c.exchange("metadata", nil)
	if err != nil {
		return err
	}
	var metadata struct {
		DriverVersion string `json:"driver_version"`
	}
	if err := json.Unmarshal(b, &metadata); err != nil {
		return err
	}
	if metadata.DriverVersion != s.DriverVersion {
		return fmt.Errorf("driver version %q does not match %q", metadata.DriverVersion, s.DriverVersion)
	}
	b, err = c.exchange("authorization_status", nil)
	if err != nil {
		return err
	}
	var status struct {
		Mode       string `json:"permission_mode"`
		Configured bool   `json:"capability_manifest_configured"`
		Approved   bool   `json:"capability_manifest_approved_at_startup"`
		Valid      bool   `json:"capability_manifest_valid"`
		Manifest   struct {
			Hash string `json:"sha256"`
		} `json:"capability_manifest"`
	}
	if err := json.Unmarshal(b, &status); err != nil {
		return err
	}
	if status.Mode != "bounded" || !status.Configured || !status.Approved || !status.Valid || status.Manifest.Hash != s.ManifestHash {
		return errors.New("daemon does not hold the expected approved bounded manifest")
	}
	return nil
}

func (c *Client) Observe(s Scope, r Request) Result {
	out := NewResult(s)
	if err := r.Validate(); err != nil {
		return out.Failed("invalid_request", err)
	}
	if r.PID == 0 && r.WindowID == 0 {
		r.PID, r.WindowID = s.PID, s.WindowID
	}
	if r.PID != s.PID || r.WindowID != s.WindowID {
		return out.Failed("outside_scope", errors.New("request does not name the selected window"))
	}
	if err := c.Verify(s); err != nil {
		return out.Failed("scope_mismatch", err)
	}
	if r.MaxElements == 0 {
		r.MaxElements = 250
	}
	if r.MaxDepth == 0 {
		r.MaxDepth = 25
	}
	args := map[string]any{"pid": r.PID, "window_id": r.WindowID, "include_accessibility_tree": true,
		"include_screenshot": r.IncludeScreenshot == nil || *r.IncludeScreenshot, "max_dimension": MaxDimension, "max_elements": r.MaxElements, "max_depth": r.MaxDepth}
	if r.Query != "" {
		args["query"] = r.Query
	}
	b, err := c.exchange("call", args)
	if err != nil {
		return out.Failed("transport_error", err)
	}
	return Project(s, b)
}

// CallBridge is the adapter's sole operation; it can reach only its OS-granted
// socket. The bridge owns upstream discovery, scope verification and lifetime.
func CallBridge(ctx context.Context, socket string, r Request) (Result, error) {
	if err := r.Validate(); err != nil {
		return Result{}, err
	}
	c, err := Dial(ctx, socket)
	if err != nil {
		return Result{}, err
	}
	defer c.Close()
	if err := json.NewEncoder(c.conn).Encode(r); err != nil {
		return Result{}, err
	}
	b, err := ReadLine(c.conn, MaxResponse)
	if err != nil {
		return Result{}, err
	}
	var out Result
	if err := json.Unmarshal(b, &out); err != nil {
		return out, err
	}
	if out.Version != 1 || out.Operation != "desktop_observe" {
		return out, errors.New("unexpected bridge result")
	}
	if (r.PID != 0 && (out.Scope.PID != r.PID || out.Scope.WindowID != r.WindowID)) || len(out.Text) > MaxText || len(out.SnapshotID) > 512 {
		return out, errors.New("bridge result differs from the requested scope or exceeds limits")
	}
	if out.Screenshot != nil {
		if err := out.Screenshot.Validate(); err != nil {
			return out, err
		}
	}
	return out, nil
}
