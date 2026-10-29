package main

import "testing"

func TestWorkToolVisibilityDependsOnlyOnScope(t *testing.T) {
	for _, fresh := range []bool{false, true} {
		for _, header := range []string{"", "delegate", "design:", "blender:"} {
			c := continuation{Reach: reachOf(map[string]*schema{"design.tools": {Type: "object"}, "blender.tools": {Type: "object"}}, nil)}
			c.Input.Header, c.Input.Fresh = header, fresh
			all := c.Reach
			for _, selected := range [][]string{nil, {}, {"design.tools"}} {
				c.Reach = holding(all, selected)
				got := map[string]bool{}
				for _, tool := range tools(c) {
					got[tool.name] = true
				}
				if got["design_tools"] != (selected == nil || len(selected) == 1) || got["blender_tools"] != (selected == nil) {
					t.Fatalf("fresh=%v header=%q selection=%v: %v", fresh, header, selected, got)
				}
			}
			c.Reach = all
			delete(c.Reach, "blender.tools")
			for _, tool := range tools(c) {
				if tool.name == "blender_tools" {
					t.Fatal("offered tool outside inherited reach")
				}
			}
		}
	}
}
