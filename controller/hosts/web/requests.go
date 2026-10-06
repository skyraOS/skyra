package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type requestRecord struct {
	RequestID    string `json:"requestId"`
	Status       string `json:"status"`
	TS           int64  `json:"ts"`
	Code         string `json:"code,omitempty"`
	Message      string `json:"message,omitempty"`
	Fingerprint  string `json:"fingerprint"`
	Conversation string `json:"conversation"`
}

// Caller holds mu. Admission is durable before work starts. Unknown outcomes
// from a previous host are retained and never automatically executed again.
func (h *webhost) admit(m inbound, conversation string) error {
	if h.requests == nil {
		h.requests = map[string]*requestRecord{}
	}
	h.requests[m.RequestID] = &requestRecord{RequestID: m.RequestID, Status: "accepted", TS: time.Now().UnixMilli(), Fingerprint: fingerprint(m), Conversation: conversation}
	if err := h.persistRequests(); err != nil {
		delete(h.requests, m.RequestID)
		return err
	}
	return nil
}

func (h *webhost) finishRequest(id, status, code, message string) {
	if r := h.requests[id]; r != nil {
		r.Status = status
		r.Code = code
		r.Message = message
		r.TS = time.Now().UnixMilli()
	}
	if err := h.persistRequests(); err != nil {
		fmt.Fprintf(h.log, "webhost: request ledger: %v\n", err)
	}
}

func (h *webhost) persistRequests() error {
	if h.dir == "" {
		return nil
	} // in-memory protocol unit tests
	if err := os.MkdirAll(h.dir, 0700); err != nil {
		return err
	}
	b, err := json.Marshal(h.requests)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(h.dir, ".requests-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), filepath.Join(h.dir, ".requests.json")); err != nil {
		return err
	}
	d, err := os.Open(h.dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

func (h *webhost) loadRequests() error {
	h.requests = map[string]*requestRecord{}
	b, err := os.ReadFile(filepath.Join(h.dir, ".requests.json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, &h.requests); err != nil {
		return err
	}
	if h.requests == nil {
		h.requests = map[string]*requestRecord{}
	}
	for id, r := range h.requests {
		if r == nil || r.RequestID != id {
			return fmt.Errorf("invalid request ledger")
		}
		if r.Status == "accepted" {
			r.Status = "failed"
			r.Code = "delivery-unknown"
			r.Message = "the host restarted; the outcome is unknown, inspect history before retrying"
		}
	}
	return h.persistRequests()
}

func (h *webhost) requestSnapshot() []map[string]any {
	ids := make([]string, 0, len(h.requests))
	for id := range h.requests {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		r := h.requests[id]
		out = append(out, map[string]any{"requestId": id, "status": r.Status, "ts": r.TS, "code": r.Code, "message": r.Message})
	}
	return out
}

// Preserve request provenance when the user actor sends a refreshed transcript.
func (h *webhost) acceptTranscript(c *conversation, transcript []entry) {
	for i := range transcript {
		if i < len(c.entries) && transcript[i].Role == c.entries[i].Role && transcript[i].Text == c.entries[i].Text {
			transcript[i].RequestID = c.entries[i].RequestID
		}
	}
	if c.pending != "" {
		id := c.pending
		for i := len(transcript) - 1; i >= 0; i-- {
			transcript[i].RequestID = &id
			if transcript[i].Role == "user" {
				break
			}
		}
		status := "returned"
		if len(transcript) > 0 && len(transcript[len(transcript)-1].Text) >= 3 && transcript[len(transcript)-1].Text[:3] == "⚠" {
			status = "failed"
		}
		h.finishRequest(id, status, "", "")
		c.pending = ""
	}
	c.entries = stamp(transcript)
}

func fingerprint(m inbound) string {
	b, _ := json.Marshal(m)
	digest := sha256.Sum256(b)
	return hex.EncodeToString(digest[:])
}
