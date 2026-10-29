package main

import "testing"

func TestWorkToolVisibilityDependsOnlyOnScope(t *testing.T) {
	for _, fresh := range []bool{false, true} {
		for _, header := range []string{"", "delegate", "design:", "blender:"} {
			c := continuation{Tools: map[string]*schema{"design.tools": {Type: "object"}, "blender.tools": {Type: "object"}}}
			c.Input.Header, c.Input.Fresh = header, fresh
			for _, selected := range [][]string{nil, {}, {"design_tools"}} {
				c.WorkTools = nil
				if selected != nil {
					c.WorkTools = &selected
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
			c.WorkTools = nil
			for _, tool := range tools(c) {
				if tool.Name == "blender_tools" {
					t.Fatal("offered tool outside inherited reach")
				}
			}
		}
	}
}
