// Proposal 58: shapes, as Shape.lean defines them. The shape in a base type
// is inferred by the core from what arrives, never supplied by the shell: the
// keys, recursively, with each leaf's JSON kind. Key order, array order and
// array length are not in it; key sets are strict; there is one number kind;
// a repeated key has no shape.
//
// What arrives is decided by its contract's declared output: under text it
// is a string whatever it looks like; under json it is decoded. Output that
// does not decode arrives as an error, {at, error: bytes, message}, holding
// the bytes, so every arrival has a shape. bytes is a kind JSON cannot
// express: no tool's own {"error": …} shares the error's base.
package tree

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"unicode/utf8"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/ast"
	"cel.dev/cel-go/common/operators"
)

// A ShapeKind is a leaf's JSON kind, or an array or object.
type ShapeKind uint8

const (
	ShapeNull ShapeKind = iota
	ShapeBool
	ShapeNum
	ShapeStr
	ShapeBytes // made only by the system: untrusted bytes that did not decode
	ShapeArr
	ShapeObj
)

// A Shape is a base type's structure. An array's Elems are the shapes of its
// elements, as a set: once each, sorted by encoding. An object's Fields are
// its keys, sorted bytewise, each with its value's shape.
type Shape struct {
	Kind   ShapeKind
	Elems  []Shape
	Fields []Field
}

type Field struct {
	Key   string
	Shape Shape
}

// Enc is the canonical encoding, what the base's hash is taken over.
func (s Shape) Enc() string {
	switch s.Kind {
	case ShapeNull:
		return "null"
	case ShapeBool:
		return "boolean"
	case ShapeNum:
		return "number"
	case ShapeStr:
		return "string"
	case ShapeBytes:
		return "bytes"
	case ShapeArr:
		parts := make([]string, len(s.Elems))
		for i, e := range s.Elems {
			parts[i] = e.Enc()
		}
		return "[" + strings.Join(parts, ",") + "]"
	}
	parts := make([]string, len(s.Fields))
	for i, f := range s.Fields {
		parts[i] = `"` + escape(f.Key) + `":` + f.Shape.Enc()
	}
	return "{" + strings.Join(parts, ",") + "}"
}

// escape is a key as a JSON string's contents: `"` and `\` escaped, and
// control characters as \u00xx. Nothing else is escaped.
func escape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r < 32:
			const hex = "0123456789abcdef"
			b.WriteString(`\u00`)
			b.WriteByte(hex[r/16])
			b.WriteByte(hex[r%16])
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ShapeOf is the shape of a decoded value: nil, bool, float64, string,
// []any or map[string]any, as DecodeJSON gives, or []byte, as an error
// holds.
func ShapeOf(v any) Shape {
	switch v := v.(type) {
	case nil:
		return Shape{Kind: ShapeNull}
	case bool:
		return Shape{Kind: ShapeBool}
	case float64:
		return Shape{Kind: ShapeNum}
	case string:
		return Shape{Kind: ShapeStr}
	case []byte:
		return Shape{Kind: ShapeBytes}
	case []any:
		var elems []Shape
		seen := map[string]bool{}
		for _, x := range v {
			s := ShapeOf(x)
			if e := s.Enc(); !seen[e] {
				seen[e] = true
				elems = append(elems, s)
			}
		}
		slices.SortFunc(elems, func(a, b Shape) int { return strings.Compare(a.Enc(), b.Enc()) })
		return Shape{Kind: ShapeArr, Elems: elems}
	}
	m := v.(map[string]any)
	fields := make([]Field, 0, len(m))
	for k, x := range m {
		fields = append(fields, Field{Key: k, Shape: ShapeOf(x)})
	}
	slices.SortFunc(fields, func(a, b Field) int { return strings.Compare(a.Key, b.Key) })
	return Shape{Kind: ShapeObj, Fields: fields}
}

// A Failure is why bytes did not decode.
type Failure uint8

const (
	NotUTF8   Failure = iota
	NotJSON           // not one JSON value
	Duplicate         // a key repeated in one object
)

// A DecodeError is where bytes did not decode and why. At is the offset of
// the first invalid UTF-8 byte; for a syntax error, the bytes the decoder had
// read when it found the error; for a repeated key, just past the repeat.
type DecodeError struct {
	At      int
	Failure Failure
	Key     string // the repeated key
}

// Error is what failed, where, and an example of what was expected:
// Shape.lean's Failure.message.
func (e *DecodeError) Error() string {
	at := itoa(e.At)
	switch e.Failure {
	case NotUTF8:
		return "not UTF-8 at byte " + at + ": text is UTF-8, e.g. é is the two bytes c3 a9"
	case Duplicate:
		return "not JSON at byte " + at + `: the key "` + escape(e.Key) +
			`" appears twice; each key appears once, e.g. {"a": 1, "b": 2}`
	}
	return "not JSON at byte " + at + `: the output is one JSON value, e.g. {"name": "a", "size": 3}`
}

// firstInvalid is the offset of the first byte that does not begin a UTF-8
// sequence, or -1.
func firstInvalid(b []byte) int {
	for i := 0; i < len(b); {
		r, n := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && n <= 1 {
			return i
		}
		i += n
	}
	return -1
}

// DecodeJSON decodes one JSON value, refusing a repeated key at any depth.
// Numbers are one kind, float64. Its error is a *DecodeError.
func DecodeJSON(b []byte) (any, error) {
	if i := firstInvalid(b); i >= 0 {
		return nil, &DecodeError{At: i, Failure: NotUTF8}
	}
	if !json.Valid(b) {
		var v any
		var se *json.SyntaxError
		if err := json.Unmarshal(b, &v); errors.As(err, &se) {
			return nil, &DecodeError{At: int(se.Offset), Failure: NotJSON}
		}
		return nil, &DecodeError{Failure: NotJSON}
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	return decodeValue(dec)
}

func decodeValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, &DecodeError{At: int(dec.InputOffset()), Failure: NotJSON}
	}
	switch t := tok.(type) {
	case json.Delim:
		if t == '[' {
			xs := []any{}
			for dec.More() {
				x, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				xs = append(xs, x)
			}
			dec.Token()
			return xs, nil
		}
		m := map[string]any{}
		for dec.More() {
			kt, _ := dec.Token()
			k := kt.(string)
			if _, dup := m[k]; dup {
				return nil, &DecodeError{At: int(dec.InputOffset()), Failure: Duplicate, Key: k}
			}
			x, err := decodeValue(dec)
			if err != nil {
				return nil, err
			}
			m[k] = x
		}
		dec.Token()
		return m, nil
	case json.Number:
		f, err := t.Float64()
		if err != nil { // out of float64's range
			return nil, &DecodeError{At: int(dec.InputOffset()), Failure: NotJSON}
		}
		return f, nil
	}
	return tok, nil
}

// An Output is a contract's declared output. The zero value is text: a
// contract without a declaration is text.
type Output uint8

const (
	OutputText Output = iota
	OutputJSON
)

// ErrorValue is an error observation, {at: number, error: bytes, message:
// string}: the bytes that did not decode, untouched; where the first error
// is; and what failed, with an example of what was expected.
func ErrorValue(b []byte, e *DecodeError) map[string]any {
	return map[string]any{"error": b, "at": float64(e.At), "message": e.Error()}
}

// Arrive is what arrives from a contract's output bytes, and its shape.
// Bytes that are not UTF-8 are the error under either declaration. Under
// text it is the string, even when it reads as JSON; under json, the decoded
// value, or the error saying where and why it did not decode.
func Arrive(out Output, b []byte) (any, Shape) {
	var v any
	if i := firstInvalid(b); i >= 0 {
		v = ErrorValue(slices.Clone(b), &DecodeError{At: i, Failure: NotUTF8})
	} else if out == OutputText {
		v = string(b)
	} else if d, err := DecodeJSON(b); err != nil {
		v = ErrorValue(slices.Clone(b), err.(*DecodeError))
	} else {
		v = d
	}
	return v, ShapeOf(v)
}

// FailedValue is what a process that exited non-zero gives: its status,
// and what it wrote to stdout and stderr, as bytes, untouched; a failing
// process's output need not follow its contract. bytes is a kind JSON
// cannot express, so no tool's own output shares a failure's base.
func FailedValue(exit int, stdout, stderr []byte) map[string]any {
	return map[string]any{"exit": float64(exit), "stdout": stdout, "stderr": stderr}
}

// ArriveExit is what arrives from a process: exiting 0, its stdout read as
// its contract declares (Arrive); otherwise the failure.
func ArriveExit(out Output, stdout []byte, exit int, stderr []byte) (any, Shape) {
	if exit == 0 {
		return Arrive(out, stdout)
	}
	v := FailedValue(exit, slices.Clone(stdout), slices.Clone(stderr))
	return v, ShapeOf(v)
}

// --- guards typed by shape ---

// celType is a position's CEL type from its shape. An object is a map, so a
// field's kind is checked where the guard selects it (CheckGuard), not by
// CEL.
func celType(s Shape) *cel.Type {
	switch s.Kind {
	case ShapeNull:
		return cel.NullType
	case ShapeBool:
		return cel.BoolType
	case ShapeNum:
		return cel.DoubleType
	case ShapeStr:
		return cel.StringType
	case ShapeBytes:
		return cel.BytesType
	case ShapeArr:
		if len(s.Elems) == 1 {
			return cel.ListType(celType(s.Elems[0]))
		}
		return cel.ListType(cel.DynType)
	}
	return cel.MapType(cel.StringType, cel.DynType)
}

// CheckGuard checks a stored chain against the shapes of its base's
// positions, at commit: every key a condition selects is in the shape, and
// every condition compiles over the positions' types. A guard reading a key
// its shape lacks, or a position with no shape, is refused.
func CheckGuard(chain []Link, shapes map[Position]Shape) error {
	byName := map[string]Shape{}
	for p, s := range shapes {
		byName[p.cel()] = s
	}
	for _, l := range chain {
		parsed, iss := parseEnv.Parse(l.Cond)
		if iss != nil && iss.Err() != nil {
			return iss.Err()
		}
		if err := checkKeys(parsed.NativeRep().Expr(), byName); err != nil {
			return err
		}
		opts := []cel.EnvOption{cel.CrossTypeNumericComparisons(true)}
		for _, name := range idents(parsed.NativeRep().Expr()) {
			if s, ok := byName[name]; ok {
				opts = append(opts, cel.Variable(name, celType(s)))
			} else if isPosition(name) {
				return errors.New(name + " is not a position of this base")
			}
		}
		env, err := cel.NewEnv(opts...)
		if err != nil {
			return err
		}
		if _, iss := env.Compile(l.Cond); iss != nil && iss.Err() != nil {
			return errors.New("the guard does not check: " + iss.Err().Error())
		}
	}
	return nil
}

// checkKeys walks every selection rooted at a position and follows it down
// the shape: a key the shape lacks, or a key selected from what is not an
// object, is refused. A test-only selection (has) asks, so it is not
// refused.
func checkKeys(e ast.Expr, shapes map[string]Shape) error {
	if _, err := follow(e, shapes); err != nil {
		return err
	}
	switch e.Kind() {
	case ast.CallKind:
		c := e.AsCall()
		if t := c.Target(); t != nil && t.Kind() != ast.UnspecifiedExprKind {
			if err := checkKeys(t, shapes); err != nil {
				return err
			}
		}
		for _, a := range c.Args() {
			if err := checkKeys(a, shapes); err != nil {
				return err
			}
		}
	case ast.SelectKind:
		return checkKeys(e.AsSelect().Operand(), shapes)
	case ast.ComprehensionKind:
		c := e.AsComprehension()
		for _, x := range []ast.Expr{c.IterRange(), c.AccuInit(), c.LoopCondition(), c.LoopStep(), c.Result()} {
			if err := checkKeys(x, shapes); err != nil {
				return err
			}
		}
	}
	return nil
}

// follow is the shape a selection chain reaches, if it is rooted at a
// position and every step is known; a missing key is an error.
func follow(e ast.Expr, shapes map[string]Shape) (*Shape, error) {
	switch e.Kind() {
	case ast.IdentKind:
		if s, ok := shapes[e.AsIdent()]; ok {
			return &s, nil
		}
	case ast.SelectKind:
		sel := e.AsSelect()
		if sel.IsTestOnly() {
			return nil, nil
		}
		return field(sel.Operand(), sel.FieldName(), shapes)
	case ast.CallKind:
		c := e.AsCall()
		if c.FunctionName() != operators.Index {
			return nil, nil
		}
		if key, ok := literalString(c.Args()[1]); ok {
			return field(c.Args()[0], key, shapes)
		}
		s, err := follow(c.Args()[0], shapes)
		if err != nil || s == nil || s.Kind != ShapeArr || len(s.Elems) != 1 {
			return nil, err
		}
		return &s.Elems[0], nil
	}
	return nil, nil
}

func field(operand ast.Expr, key string, shapes map[string]Shape) (*Shape, error) {
	s, err := follow(operand, shapes)
	if err != nil || s == nil {
		return nil, err
	}
	if s.Kind != ShapeObj {
		return nil, errors.New("selects " + key + " from " + s.Enc() + ", which has no keys")
	}
	for i := range s.Fields {
		if s.Fields[i].Key == key {
			return &s.Fields[i].Shape, nil
		}
	}
	return nil, errors.New("selects " + key + ", not a key of " + s.Enc())
}
