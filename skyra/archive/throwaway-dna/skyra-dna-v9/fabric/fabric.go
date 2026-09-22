// Package fabric is TRANSPORT — the second of the two units ("the CTU executes... the
// FABRIC transports between them", 2026-08-18/README.md). It carries a cut's TRANSFER
// legs across the wire, FETCHes signed segments by content-address for REMEMBER's
// gather, and holds the SIGN/VERIFY trust half. Unchanged in shape from v8 except that
// what it carries is a Transfer (the dissolved-Slice leg), not v8's transport Move.
//
// AUTHORITY: skyra/specs/2026-08-22/fabric/signed-provenance.md, on
// skyra/specs/2026-08-18/network/traffic-routing.md; wire form per
// skyra/specs/2026-08-23/possibility-graph.md §8.
package fabric

import c "skyra-v1/src/skyra/skyra-dna-v9/causal"

// Fabric is the transport boundary the CTU submits to.
type Fabric interface {
	// Approach is the one EGRESS: submit a resolved Transfer (a cut leg) toward its target
	// CTU. The fabric routes by the envelope (Target, MoveID, Header) and never reads the
	// child's payload. from carries the return endpoint so the response leg can be formed
	// without a home identity in the child. traffic-routing.md.
	Approach(from c.Target, t c.Transfer) c.Outcome

	// Fetch pulls a signed Segment by its content-address — the per-hop operation of
	// REMEMBER's gather. Content-addressing lets the gather dedup and skip anything already
	// durable. kernel/remember-gather.md.
	Fetch(ref c.SegmentRef) (c.Segment, c.Outcome)
}

// Custody is the TRUST mechanism of the fabric boundary: it makes provenance survive a
// cut unforgeably. A CTU SIGNS the segment its move produced; a downstream CTU VERIFIES
// it without trusting the sender's word. fabric/signed-provenance.md.
type Custody interface {
	// Sign fills both Signature and content-addressed Ref for a segment the local CTU produced.
	Sign(seg c.Segment) (c.Segment, c.Outcome)

	// Verify checks a segment's signature across a trust boundary — the point where a foreign
	// stamp is or is not trusted. Used by ctu.Recognize before a returned observation is admitted.
	Verify(seg c.Segment) c.Outcome
}

// SegmentVerifier is the public half held by transport and gather. It cannot sign on
// any CTU's behalf.
type SegmentVerifier interface {
	Verify(seg c.Segment) c.Outcome
}

// HorizonMinter is the privileged root capability held ONLY by the kernel — the reason
// origination (MINT) is the kernel's by construction. horizon-and-root.md.
type HorizonMinter interface {
	Mint(root c.Name) (c.Horizon, c.Outcome)
}

// Authority is the bearer/public half applied to a Horizon. Attenuate uses the proof
// carried by the token (BOUND); VerifyHorizon needs only the root public key.
type Authority interface {
	Attenuate(h c.Horizon, add c.Caveat) (c.Horizon, c.Outcome)
	VerifyHorizon(h c.Horizon) c.Outcome
}

// Ingress is the normalized arrival shared by local and external transports — a Transfer
// leg reaching GREET (request) or RECOGNIZE (response). traffic-routing.md.
type Ingress interface {
	Arrive(from c.Target, t c.Transfer) c.Outcome
}
