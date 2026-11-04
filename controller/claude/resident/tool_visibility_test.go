package main

import (
	"slices"
	"testing"
)

// A thread is offered what it was granted and nothing else: reach alone
// offers nothing, and a grant of what is not in reach offers nothing.
func TestToolsAreOnlyWhatTheThreadIsGranted(t *testing.T) {
	for _, tc := range []struct {
		grant []string
		want  []string
	}{
		{nil, nil},
		{[]string{}, nil},
		{[]string{"design.tools"}, []string{"design_tools"}},
		{[]string{"blender_tools"}, []string{"blender_tools"}},
		{[]string{"design.tools", "blender.tools"}, []string{"blender_tools", "design_tools"}},
		{[]string{"repo.write"}, nil},
	} {
		c := continuation{Reach: reachOf(map[string]*schema{"design.tools": {Type: "object"}, "blender.tools": {Type: "object"}}, nil)}
		c.Input.Tools = tc.grant
		var got []string
		for _, tool := range tools(c) {
			got = append(got, tool.Name)
		}
		slices.Sort(got)
		if !slices.Equal(got, tc.want) {
			t.Errorf("granted %v: offered %v, want %v", tc.grant, got, tc.want)
		}
	}
}
