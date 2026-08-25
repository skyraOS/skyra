package fabric

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"

	c "skyra-v1/src/skyra/skyra-dna-v8/causal"
)

// Security is a key factory, never injected into a machine component. It
// issues narrower capability views: root minting to the kernel, bearer
// attenuation/public verification to CTUs and transport, one CTU's signer to
// that CTU, and public segment verification to transport/gather.
type Security struct {
	mu       sync.RWMutex
	rootPub  ed25519.PublicKey
	rootPriv ed25519.PrivateKey
	ctuPub   map[string]ed25519.PublicKey
	ctuPriv  map[string]ed25519.PrivateKey
}

type horizonMinter struct{ rootPriv ed25519.PrivateKey }
type authority struct{ rootPub ed25519.PublicKey }
type custody struct {
	name c.Name
	priv ed25519.PrivateKey
	pubs map[string]ed25519.PublicKey
}
type segmentVerifier struct{ pubs map[string]ed25519.PublicKey }

func NewSecurity(ctus ...c.Name) (*Security, c.Outcome) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, c.Failure
	}
	s := &Security{
		rootPub: pub, rootPriv: priv,
		ctuPub: make(map[string]ed25519.PublicKey), ctuPriv: make(map[string]ed25519.PrivateKey),
	}
	for _, name := range ctus {
		if s.RegisterCTU(name) != c.Success {
			return nil, c.Failure
		}
	}
	return s, c.Success
}

func (s *Security) RegisterCTU(name c.Name) c.Outcome {
	if name.Value == "" {
		return c.Failure
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return c.Failure
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.ctuPub[name.Value]; exists {
		return c.Failure
	}
	s.ctuPub[name.Value], s.ctuPriv[name.Value] = pub, priv
	return c.Success
}

func clonePublicKeys(source map[string]ed25519.PublicKey) map[string]ed25519.PublicKey {
	cloned := make(map[string]ed25519.PublicKey, len(source))
	for name, key := range source {
		cloned[name] = append(ed25519.PublicKey(nil), key...)
	}
	return cloned
}

func (s *Security) Minter() HorizonMinter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &horizonMinter{rootPriv: append(ed25519.PrivateKey(nil), s.rootPriv...)}
}

func (s *Security) Authority() Authority {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &authority{rootPub: append(ed25519.PublicKey(nil), s.rootPub...)}
}

func (s *Security) Custody(name c.Name) (Custody, c.Outcome) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	priv, ok := s.ctuPriv[name.Value]
	if !ok {
		return nil, c.Failure
	}
	return &custody{
		name: name, priv: append(ed25519.PrivateKey(nil), priv...), pubs: clonePublicKeys(s.ctuPub),
	}, c.Success
}

func (s *Security) Verifier() SegmentVerifier {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &segmentVerifier{pubs: clonePublicKeys(s.ctuPub)}
}

func segmentBytes(seg c.Segment) ([]byte, bool) {
	b, err := json.Marshal(struct {
		Provenance c.Provenance
		By         c.Name
	}{seg.Provenance, seg.By})
	return b, err == nil
}

func signSegment(seg c.Segment, name c.Name, priv ed25519.PrivateKey) (c.Segment, c.Outcome) {
	if seg.By != name {
		return c.Segment{}, c.Failure
	}
	b, ok := segmentBytes(seg)
	if !ok {
		return c.Segment{}, c.Failure
	}
	digest := sha256.Sum256(b)
	seg.Ref = c.SegmentRef{Hash: hex.EncodeToString(digest[:])}
	seg.Signature = c.Signature{Bytes: ed25519.Sign(priv, digest[:])}
	return seg, c.Success
}

func verifySegment(seg c.Segment, pubs map[string]ed25519.PublicKey) c.Outcome {
	pub, ok := pubs[seg.By.Value]
	if !ok {
		return c.Failure
	}
	b, ok := segmentBytes(seg)
	if !ok {
		return c.Failure
	}
	digest := sha256.Sum256(b)
	if seg.Ref.Hash != hex.EncodeToString(digest[:]) || !ed25519.Verify(pub, digest[:], seg.Signature.Bytes) {
		return c.Failure
	}
	return c.Success
}

func (v *custody) Sign(seg c.Segment) (c.Segment, c.Outcome) {
	return signSegment(seg, v.name, v.priv)
}
func (v *custody) Verify(seg c.Segment) c.Outcome { return verifySegment(seg, v.pubs) }
func (v *segmentVerifier) Verify(seg c.Segment) c.Outcome {
	return verifySegment(seg, v.pubs)
}

func rootMessage(root c.Name, next []byte) []byte {
	b, _ := json.Marshal(struct {
		Domain string
		Root   c.Name
		Next   []byte
	}{"skyra-horizon-root-v1", root, next})
	return b
}

func caveatMessage(previous []byte, caveat c.Caveat, next []byte) []byte {
	b, _ := json.Marshal(struct {
		Domain   string
		Previous []byte
		Caveat   c.Caveat
		Next     []byte
	}{"skyra-horizon-caveat-v1", previous, caveat, next})
	return b
}

func (m *horizonMinter) Mint(root c.Name) (c.Horizon, c.Outcome) {
	if root.Value == "" {
		return c.Horizon{}, c.Failure
	}
	proofPub, proofPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return c.Horizon{}, c.Failure
	}
	sig := ed25519.Sign(m.rootPriv, rootMessage(root, proofPub))
	return c.Horizon{
		Root: root, Keys: [][]byte{append([]byte(nil), proofPub...)},
		Chain: []c.Signature{{Bytes: sig}}, Proof: append([]byte(nil), proofPriv...),
	}, c.Success
}

func (a *authority) Attenuate(h c.Horizon, add c.Caveat) (c.Horizon, c.Outcome) {
	if a.VerifyHorizon(h) != c.Success || add.Constraint == "" || len(h.Proof) != ed25519.PrivateKeySize {
		return c.Horizon{}, c.Failure
	}
	current := ed25519.PrivateKey(h.Proof)
	nextPub, nextPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return c.Horizon{}, c.Failure
	}
	previous := h.Chain[len(h.Chain)-1].Bytes
	sig := ed25519.Sign(current, caveatMessage(previous, add, nextPub))
	narrowed := c.Horizon{Root: h.Root}
	narrowed.Caveats = append(append([]c.Caveat(nil), h.Caveats...), add)
	for _, key := range h.Keys {
		narrowed.Keys = append(narrowed.Keys, append([]byte(nil), key...))
	}
	narrowed.Keys = append(narrowed.Keys, append([]byte(nil), nextPub...))
	for _, link := range h.Chain {
		narrowed.Chain = append(narrowed.Chain, c.Signature{Bytes: append([]byte(nil), link.Bytes...)})
	}
	narrowed.Chain = append(narrowed.Chain, c.Signature{Bytes: sig})
	narrowed.Proof = append([]byte(nil), nextPriv...)
	return narrowed, c.Success
}

func (a *authority) VerifyHorizon(h c.Horizon) c.Outcome {
	if h.Root.Value == "" || len(h.Keys) != len(h.Caveats)+1 || len(h.Chain) != len(h.Keys) {
		return c.Failure
	}
	if len(h.Keys[0]) != ed25519.PublicKeySize ||
		!ed25519.Verify(a.rootPub, rootMessage(h.Root, h.Keys[0]), h.Chain[0].Bytes) {
		return c.Failure
	}
	for i, caveat := range h.Caveats {
		if len(h.Keys[i]) != ed25519.PublicKeySize || len(h.Keys[i+1]) != ed25519.PublicKeySize ||
			!ed25519.Verify(ed25519.PublicKey(h.Keys[i]), caveatMessage(h.Chain[i].Bytes, caveat, h.Keys[i+1]), h.Chain[i+1].Bytes) {
			return c.Failure
		}
	}
	if len(h.Proof) != ed25519.PrivateKeySize ||
		!ed25519.PublicKey(h.Keys[len(h.Keys)-1]).Equal(ed25519.PrivateKey(h.Proof).Public()) {
		return c.Failure
	}
	return c.Success
}

var (
	_ HorizonMinter   = (*horizonMinter)(nil)
	_ Authority       = (*authority)(nil)
	_ Custody         = (*custody)(nil)
	_ SegmentVerifier = (*segmentVerifier)(nil)
)
