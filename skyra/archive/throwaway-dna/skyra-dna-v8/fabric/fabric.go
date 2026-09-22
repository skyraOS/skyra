// Package fabric is TRANSPORT — the second of the two units ("the CTU executes...
// the FABRIC transports between them", 2026-08-18/README.md). In the distributed
// model it does two jobs the single-host build did not need: it carries a signed
// provenance SEGMENT across a cut, and it FETCHES segments by content-address so the
// kernel can gather the cone.
//
// AUTHORITY: skyra/specs/2026-08-22/fabric/signed-provenance.md, on
// skyra/specs/2026-08-18/network/traffic-routing.md.
package fabric

import c "skyra-v1/src/skyra/skyra-dna-v8/causal"

// Fabric is the transport boundary the CTU submits to.
type Fabric interface {
	// Approach is the one EGRESS: submit a resolved MOVE (a cut) toward its target
	// CTU. The fabric routes by the envelope (Target, MoveID, Header) and never reads
	// the payload. traffic-routing.md.
	Approach(from c.Target, m c.Move) c.Outcome

	// Fetch pulls a signed Segment by its content-address — the per-hop operation of
	// REMEMBER's gather (kernel/remember-gather.md: "each hop that names a segment on
	// another CTU is a fetch"). Content-addressing lets the gather dedup and skip
	// anything already durable.
	Fetch(ref c.SegmentRef) (c.Segment, c.Outcome)
}

// Custody is the TRUST mechanism of the fabric boundary: it makes provenance survive
// a cut unforgeably. A CTU SIGNS the segment it executed; a downstream CTU VERIFIES
// it without trusting the sender's word. "custody transfer" is MOVE's doc phrasing;
// Custody names the sign/verify half (declared in interface-rationale.md).
// fabric/signed-provenance.md.
type Custody interface {
	// Sign fills both Signature and content-addressed Ref for a segment the local CTU executed.
	Sign(seg c.Segment) (c.Segment, c.Outcome)

	// Verify checks a segment's signature across a trust boundary. Returns Failure if
	// the signature does not hold — the point where a foreign stamp is or is not
	// trusted. Used by ctu.Recognize before a returned observation is admitted.
	Verify(seg c.Segment) c.Outcome
}

// SegmentVerifier is the public half held by transport and gather. It cannot
// sign on any CTU's behalf.
type SegmentVerifier interface {
	Verify(seg c.Segment) c.Outcome
}

// HorizonMinter is the privileged root capability held only by the kernel.
type HorizonMinter interface {
	Mint(root c.Name) (c.Horizon, c.Outcome)
}

// Authority is the bearer/public half applied to a Horizon. Attenuate uses the
// proof carried by the token; VerifyHorizon needs only the root public key.
type Authority interface {
	Attenuate(h c.Horizon, add c.Caveat) (c.Horizon, c.Outcome)
	VerifyHorizon(h c.Horizon) c.Outcome
}

// Ingress is the normalized arrival shared by local and external transports.
type Ingress interface {
	Arrive(from c.Target, m c.Move) c.Outcome
}
