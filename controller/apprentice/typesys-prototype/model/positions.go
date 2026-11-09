package model

// Key positions, as Model.lean defines them. Memory stores a template by
// what it reads in the key, never by $n: an elimination or an introduction
// takes a $n of its own but no place in the key, which holds only what was
// seen, so an arrival after one is numbered later while its place in the
// key is not. Store inlines a template and names each reference by its key
// position; Bind names it by $n again in the frame it runs in.

// SeenAt is the $n a frame has seen, a tool or a value, in order: key
// position s is the s-th, from 1.
func (f Frame) SeenAt() []int {
	var ns []int
	for i, it := range f.Items() {
		if it.Tag == ItemTool || it.Tag == ItemSeen {
			ns = append(ns, i+1)
		}
	}
	return ns
}

// ToKey is $n's key position, if it was seen.
func (f Frame) ToKey(n int) (int, bool) {
	for i, m := range f.SeenAt() {
		if m == n {
			return i + 1, true
		}
	}
	return 0, false
}

// ToItem is key position s's $n.
func (f Frame) ToItem(s int) (int, bool) {
	ns := f.SeenAt()
	if s < 1 || s > len(ns) {
		return 0, false
	}
	return ns[s-1], true
}

// Store is a template as memory stores it: each argument inlined, so it
// reads only what was seen, and each reference, a join's parts among them,
// a key position. It fails on an argument that reads nothing seen.
func (f Frame) Store(t Template) (Template, bool) {
	return t.mapRefs(func(n int) (int, bool) { return f.ToKey(n) }, f.Inline)
}

// Bind is a stored template in a frame: each key position its $n there.
func (f Frame) Bind(t Template) (Template, bool) {
	return t.mapRefs(func(p int) (int, bool) { return f.ToItem(p) }, nil)
}

// mapRefs is a template with each reference renamed by to, after inline
// when inline is given; false when to has no name for one.
func (t Template) mapRefs(to func(int) (int, bool), inline func(int, []Sel) Arg) (Template, bool) {
	out := Template{Op: t.Op}
	for _, a := range t.Args {
		if a.Pieces != nil {
			return Template{}, false
		}
		if inline != nil && a.Lit == nil && a.Name == "" && a.Join == nil {
			a = inline(a.Var, a.Sels)
		}
		switch {
		case a.Join != nil:
			ps := make([]Part, len(a.Join))
			for i, p := range a.Join {
				if p.Str == nil {
					n, ok := to(p.Var)
					if !ok {
						return Template{}, false
					}
					p.Var = n
				}
				ps[i] = p
			}
			a.Join = ps
		case a.Lit == nil && a.Name == "":
			n, ok := to(a.Var)
			if !ok {
				return Template{}, false
			}
			a.Var = n
		}
		out.Args = append(out.Args, a)
	}
	return out, true
}
