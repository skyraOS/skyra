package contract

// Shapes and contents (next/30). The runtime walks a decoded value once,
// bottom-up, and gives every node two hashes: its shape (leaf values left out)
// and its content (leaf values in). A node hashes its children's hashes, not
// its children (Merkle), so two values are diffed by descending only where
// content hashes differ.
//
// The encoding is pinned, so it never drifts (shape_test.go holds golden
// hashes):
//
//	node      sha256(tag ‖ body), tag one byte: s n b z o a
//	length    every variable-length part is prefixed by its length, 8 bytes big-endian
//	child     its 32 raw hash bytes
//	string    shape: "s"     content: "s" ‖ len ‖ bytes
//	number    shape: "n"     content: "n" ‖ len ‖ the json.Number text as written
//	bool      shape: "b"     content: "b" ‖ 0|1
//	null      shape: "z"     content: "z"
//	object    "o" ‖ count ‖ for each key, sorted bytewise: len(k) ‖ k ‖ child
//	array     shape: "a" ‖ count of distinct element shapes ‖ those shapes, sorted
//	          content: "a" ‖ length ‖ each element's content, in order
//	root      Shape: sha256("skyra/shape/v1" ‖ root); Content: sha256("skyra/content/v1" ‖ root)

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"slices"
	"strconv"
	"strings"
)

type sum = [32]byte

// node is a decoded value with both hashes, and its children by key or index.
type node struct {
	shape, content sum
	value          any
	keys           []string // object: sorted
	kids           []*node  // object: by sorted key; array: by index
	leaf           bool     // a string, number, bool or null
}

func hashed(v any) *node {
	switch x := v.(type) {
	case map[string]any:
		n := &node{value: v}
		for k := range x {
			n.keys = append(n.keys, k)
		}
		slices.Sort(n.keys)
		var sb, cb bytes.Buffer
		sb.WriteByte('o')
		cb.WriteByte('o')
		count(&sb, len(n.keys))
		count(&cb, len(n.keys))
		for _, k := range n.keys {
			kid := hashed(x[k])
			n.kids = append(n.kids, kid)
			part(&sb, []byte(k))
			part(&cb, []byte(k))
			sb.Write(kid.shape[:])
			cb.Write(kid.content[:])
		}
		n.shape, n.content = sha256.Sum256(sb.Bytes()), sha256.Sum256(cb.Bytes())
		return n
	case []any:
		n := &node{value: v}
		var shapes []sum
		var cb bytes.Buffer
		cb.WriteByte('a')
		count(&cb, len(x))
		for _, e := range x {
			kid := hashed(e)
			n.kids = append(n.kids, kid)
			cb.Write(kid.content[:])
			if !slices.Contains(shapes, kid.shape) {
				shapes = append(shapes, kid.shape)
			}
		}
		slices.SortFunc(shapes, func(a, b sum) int { return bytes.Compare(a[:], b[:]) })
		var sb bytes.Buffer
		sb.WriteByte('a')
		count(&sb, len(shapes))
		for _, s := range shapes {
			sb.Write(s[:])
		}
		n.shape, n.content = sha256.Sum256(sb.Bytes()), sha256.Sum256(cb.Bytes())
		return n
	}
	n := &node{value: v, leaf: true}
	var tag byte
	var body bytes.Buffer
	switch x := v.(type) {
	case string:
		tag = 's'
		part(&body, []byte(x))
	case json.Number:
		tag = 'n'
		part(&body, []byte(x))
	case bool:
		tag = 'b'
		if x {
			body.WriteByte(1)
		} else {
			body.WriteByte(0)
		}
	default: // nil: Decode produces nothing else
		tag = 'z'
	}
	n.shape = sha256.Sum256([]byte{tag})
	n.content = sha256.Sum256(append([]byte{tag}, body.Bytes()...))
	return n
}

func count(b *bytes.Buffer, n int) {
	var l [8]byte
	binary.BigEndian.PutUint64(l[:], uint64(n))
	b.Write(l[:])
}

func part(b *bytes.Buffer, p []byte) {
	count(b, len(p))
	b.Write(p)
}

func root(domain string, s sum) string {
	r := sha256.Sum256(append([]byte(domain), s[:]...))
	return hex.EncodeToString(r[:])
}

// Shape is the hex hash of a decoded value's shape: its keys and leaf kinds,
// not its leaf values, and not its arrays' lengths.
func Shape(v any) string { return root("skyra/shape/v1", hashed(v).shape) }

// Content is the hex hash of a decoded value's content: its shape and every
// leaf value, arrays in order.
func Content(v any) string { return root("skyra/content/v1", hashed(v).content) }

// Present is a value's shape as text a model reads, e.g.
// {media: [{url: string}], text: string}. Keys are sorted; an array of
// several element shapes is [T | U].
func Present(v any) string {
	var b strings.Builder
	present(&b, hashed(v))
	return b.String()
}

func present(b *strings.Builder, n *node) {
	switch n.value.(type) {
	case map[string]any:
		b.WriteByte('{')
		for i, k := range n.keys {
			if i > 0 {
				b.WriteString(", ")
			}
			if ident(k) {
				b.WriteString(k)
			} else {
				b.WriteString(strconv.Quote(k))
			}
			b.WriteString(": ")
			present(b, n.kids[i])
		}
		b.WriteByte('}')
	case []any:
		var seen []sum
		var alts []string
		for _, kid := range n.kids {
			if !slices.Contains(seen, kid.shape) {
				seen = append(seen, kid.shape)
				var e strings.Builder
				present(&e, kid)
				alts = append(alts, e.String())
			}
		}
		slices.Sort(alts)
		b.WriteString("[" + strings.Join(alts, " | ") + "]")
	case string:
		b.WriteString("string")
	case json.Number:
		b.WriteString("number")
	case bool:
		b.WriteString("bool")
	default:
		b.WriteString("null")
	}
}

func ident(k string) bool {
	if k == "" {
		return false
	}
	for i, c := range k {
		if c != '_' && !('a' <= c && c <= 'z') && !('A' <= c && c <= 'Z') && (i == 0 || !('0' <= c && c <= '9')) {
			return false
		}
	}
	return true
}

// Delta is what changed from one value to another, by path ("$", ".key",
// "[i]", and ["key"] for a key that is not an identifier).
type Delta struct {
	Changed []string `json:"changed"` // paths in both whose content differs: leaves, or where the kind changed
	Added   []string `json:"added"`   // paths in b and not in a
	Removed []string `json:"removed"` // paths in a and not in b
	Leaves  int      `json:"leaves"`  // scalar leaves in b
	Ratio   float64  `json:"ratio"`   // len(Changed) / Leaves; 0 when Leaves is 0
}

// Diff walks two values together, skipping every subtree whose content hashes
// are equal. Array elements are compared by index.
func Diff(a, b any) Delta {
	ha, hb := hashed(a), hashed(b)
	d := Delta{Changed: []string{}, Added: []string{}, Removed: []string{}, Leaves: leaves(hb)}
	diff(&d, "$", ha, hb)
	slices.Sort(d.Changed)
	slices.Sort(d.Added)
	slices.Sort(d.Removed)
	if d.Leaves > 0 {
		d.Ratio = float64(len(d.Changed)) / float64(d.Leaves)
	}
	return d
}

func diff(d *Delta, path string, a, b *node) {
	if a.content == b.content {
		return
	}
	switch av := a.value.(type) {
	case map[string]any:
		if _, ok := b.value.(map[string]any); !ok {
			break
		}
		for i, k := range a.keys {
			if j, ok := slices.BinarySearch(b.keys, k); ok {
				diff(d, Key(path, k), a.kids[i], b.kids[j])
			} else {
				d.Removed = append(d.Removed, Key(path, k))
			}
		}
		for _, k := range b.keys {
			if _, ok := av[k]; !ok {
				d.Added = append(d.Added, Key(path, k))
			}
		}
		return
	case []any:
		if _, ok := b.value.([]any); !ok {
			break
		}
		for i := 0; i < max(len(a.kids), len(b.kids)); i++ {
			at := path + "[" + strconv.Itoa(i) + "]"
			switch {
			case i >= len(b.kids):
				d.Removed = append(d.Removed, at)
			case i >= len(a.kids):
				d.Added = append(d.Added, at)
			default:
				diff(d, at, a.kids[i], b.kids[i])
			}
		}
		return
	}
	d.Changed = append(d.Changed, path)
}

func leaves(n *node) int {
	if n.leaf {
		return 1
	}
	total := 0
	for _, kid := range n.kids {
		total += leaves(kid)
	}
	return total
}

// Key extends a path by one object key: .key, or ["key"] for a key that is
// not an identifier.
func Key(path, key string) string {
	if ident(key) {
		return path + "." + key
	}
	return path + "[" + strconv.Quote(key) + "]"
}

// Type is a return's type (next/30): the actor that returned it, the contract
// it was invoked through, and the shape of what it returned. The arguments
// are hex hashes; the actor is its identity, never its name.
func Type(actor, contract, shape string) string {
	var b bytes.Buffer
	b.WriteString("skyra/type/v1")
	part(&b, []byte(actor))
	part(&b, []byte(contract))
	part(&b, []byte(shape))
	h := sha256.Sum256(b.Bytes())
	return hex.EncodeToString(h[:])
}
