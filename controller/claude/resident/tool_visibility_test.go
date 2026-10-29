package main

import (
	"strings"
	"testing"
)

func TestWorkToolVisibilityDependsOnlyOnScope(t *testing.T) {
	for _, fresh := range []bool{false, true} {
		for _, header := range []string{"", "delegate", "design:", "blender:"} {
			c := continuation{Tools: map[string]*schema{"design.tools": {Type: "object"}, "blender.tools": {Type: "object"}}}
			c.Input.Header, c.Input.Fresh = header, fresh
			for _, selected := range [][]string{nil, {}, {"design_tools"}} {
				c.Held = nil
				if selected != nil {
					c.Held = map[string][]*schema{}
					for _, name := range selected {
						target := strings.ReplaceAll(name, "_", ".")
						c.Held[target] = []*schema{c.Tools[target]}
					}
				}
				got := map[string]bool{}
				for _, tool := range tools(c) {
					got[tool.Name] = true
				}
				if got["design_tools"] != (selected == nil || len(selected) == 1) || got["blender_tools"] != (selected == nil) {
					t.Fatalf("fresh=%v header=%q selection=%v: %v", fresh, header, selected, got)
				}
			}
			delete(c.Tools, "blender.tools")
			c.Held = nil
			for _, tool := range tools(c) {
				if tool.Name == "blender_tools" {
					t.Fatal("offered tool outside inherited reach")
				}
			}
		}
	}
}
