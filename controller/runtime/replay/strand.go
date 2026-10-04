package replay

import (
	"errors"
	"fmt"

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
)

// Strand is what the strand syscall hands its requester (next/32, H4): the
// stamps from an observation up to the invoke that caused it, made through the
// contract the requester named, and every record of the work beneath that
// invoke. The requester is a worker, so it may hold hashes.
//
// A strand is seen one Horizon at a time. A call into a narrowed child appears
// as one call and its return: Opaque maps that return to the call, and nothing
// in between is included.
type Strand struct {
	Chain   []string                 `json:"chain"`
	Records map[string]record.Record `json:"records"`
	Opaque  map[string]string        `json:"opaque,omitempty"`
}

// refused is a walk that cannot hand the requester a strand: its text is the
// refusal, a horizon fault of the syscall.
type refused string

func (r refused) Error() string { return string(r) }

// Refused reports whether a ReadStrand error is the walk's refusal, rather
// than the tree failing to read.
func Refused(err error) bool {
	var r refused
	return errors.As(err, &r)
}

func refusal(format string, args ...any) error { return refused(fmt.Sprintf(format, args...)) }

// ReadStrand walks the tree from observation up, through each record's
// predecessor, to the invoke of target that fits schema, emitted in the
// requester's Horizon (origin). It refuses when no such invoke is reached, and
// when the walk would leave the requester's Horizon. Then it gathers every
// record beneath that invoke: each resolve's discharges and the chains that
// led to them, stopping at the calls into narrowed children. Every Horizon on
// the way is re-derived from genesis, as replay does; records are immutable,
// so the same walk gives the same strand.
func ReadStrand(root, observation, target string, schema *contract.Schema, origin string) (Strand, error) {
	w := walker{root: root, seen: map[string]horizon.Horizon{}}
	st := Strand{Records: map[string]record.Record{}, Opaque: map[string]string{}}

	// frame is the Horizon of the frame a record belongs to: the one an
	// invoke opens or a resolve resumes, else the one it was emitted under.
	frame := func(stamp string, r record.Record) (horizon.Horizon, error) {
		if r.State == record.Resolve || r.State == record.Invoke || r.Predecessor == "" {
			return w.at(stamp)
		}
		return w.at(r.Predecessor)
	}
	// step reads one record of the walk. In another Horizon, it may only be
	// the discharge of a call into a narrowed child made from the requester's
	// Horizon: then the walk jumps to that call, and next is the call's
	// predecessor.
	step := func(stamp string) (r record.Record, next string, err error) {
		if r, err = tree.Load(root, stamp); err != nil {
			return r, "", err
		}
		f, err := frame(stamp, r)
		if err != nil {
			return r, "", err
		}
		if f.Origin == origin {
			st.Records[stamp] = r
			return r, r.Predecessor, nil
		}
		call := f.Frame
		c, err := tree.Load(root, call)
		if err != nil {
			return r, "", err
		}
		caller, err := w.at(c.Predecessor)
		if err != nil {
			return r, "", err
		}
		if (r.State != record.Return && r.State != record.Fault) || f.Origin != call || caller.Origin != origin {
			return r, "", refusal("the walk would leave the requester's Horizon at %s", record.Short(stamp))
		}
		st.Records[stamp], st.Records[call] = r, c
		st.Opaque[stamp] = call
		return r, c.Predecessor, nil
	}

	// The cause: up from the frame the observation belongs to, through each
	// waiting caller, to the invoke made through the contract. A frame in
	// another Horizon is passed only if it is a narrowed child called from
	// the requester's Horizon; above the requester's Horizon, the walk stops.
	o, err := tree.Load(root, observation)
	if err != nil {
		return Strand{}, refusal("no record %s", record.Short(observation))
	}
	f, err := frame(observation, o)
	if err != nil {
		return Strand{}, err
	}
	cause := ""
	for cause == "" {
		inv, err := tree.Load(root, f.Frame)
		if err != nil {
			return Strand{}, err
		}
		if inv.Predecessor == "" || inv.State != record.Invoke {
			return Strand{}, refusal("no invoke of %s through that contract caused this observation", target)
		}
		if f.Origin != origin {
			caller, err := w.at(inv.Predecessor)
			if err != nil {
				return Strand{}, err
			}
			if f.Origin != f.Frame || caller.Origin != origin {
				return Strand{}, refusal("the observation is outside the requester's Horizon")
			}
		} else if inv.To == target && fits(schema, inv.Payload) {
			cause = f.Frame
			break
		}
		if f.Frame == origin || f.Resume == nil {
			return Strand{}, refusal("no invoke of %s through that contract caused this observation, in the requester's Horizon", target)
		}
		f = *f.Resume
	}

	// The chain: the observation's predecessors, up to the cause.
	for s := observation; ; {
		_, next, err := step(s)
		if err != nil {
			return Strand{}, err
		}
		st.Chain = append(st.Chain, s)
		if call, jumped := st.Opaque[s]; jumped {
			st.Chain = append(st.Chain, call)
			if call == cause {
				break
			}
		}
		if s == cause {
			break
		}
		if next == "" {
			return Strand{}, errors.New("the observation's chain does not reach its cause")
		}
		s = next
	}

	// Beneath it: every discharge each resolve names, and the chain that led
	// to it, up to a record already gathered.
	var beneath func(stamp string) error
	beneath = func(stamp string) error {
		for s := stamp; s != ""; {
			if _, done := st.Records[s]; done {
				return nil
			}
			r, next, err := step(s)
			if err != nil {
				return err
			}
			if err := discharges(r, beneath); err != nil {
				return err
			}
			s = next
		}
		return nil
	}
	for _, s := range st.Chain {
		if err := discharges(st.Records[s], beneath); err != nil {
			return Strand{}, err
		}
	}
	if len(st.Opaque) == 0 {
		st.Opaque = nil
	}
	return st, nil
}

// discharges runs f on every discharge a resolve names but its predecessor.
func discharges(r record.Record, f func(string) error) error {
	if r.State != record.Resolve {
		return nil
	}
	for _, d := range r.Deps {
		if d != r.Predecessor {
			if err := f(d); err != nil {
				return err
			}
		}
	}
	return nil
}

func fits(s *contract.Schema, payload []byte) bool {
	v, err := contract.Decode(payload)
	return err == nil && contract.Fits(s, v) == nil
}
