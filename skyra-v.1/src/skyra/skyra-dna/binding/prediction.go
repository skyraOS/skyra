package binding

import (
	"skyra-v1/src/reality"
)

// Prediction — the prediction engine. Sub-binding reality under Goal.
// Predicts incoming signal — what the world will do next. Sensory anticipation.
// NOT the intent graph — that's motor planning (Goal/Action, outbound).
// Prediction is inbound: "what signal will arrive?"
//
// Compares previous prediction against what actually arrived.
// The delta is the error signal.
//
// Runtime reads (from Relation):
//   Episodic — which episode am I in
//   Working Memory — what actually arrived
//   Prospective — what are the active external goals
type PredictionTrace struct {
	Predicted string
	Actual    string
	Delta     float64
}

type Prediction struct {
	id      string
	BeingID string
	Traces  map[string][]PredictionTrace // episodeID → prediction history
	Current string                       // last prediction for comparison
}

func (c *Prediction) ID() string {
	return c.id
}

func (c *Prediction) Create(r *reality.Relation) reality.Reality {
	return &Prediction{
		id:     reality.NewID(),
		Traces: make(map[string][]PredictionTrace),
	}
}

func (c *Prediction) Observe(r *reality.Relation) {
	// assemble context window for inference call:
	// 1. current episode — what's happening right now
	// 2. similar past episodes — retrieved by activation score
	// 3. previous prediction — what was predicted last time (inbound signal, not intent graph)
	//
	// inference reads all three, outputs new prediction of incoming signal
	// delta = previous prediction vs what actually arrived
}

func (c *Prediction) Express(r *reality.Relation) *reality.Relation {
	// TODO: if delta is bad enough, surface prediction error
	// TODO: if persistent, raise a goal ("need more information")
	//
	// Delta feeds three other systems:
	// 1. Episodic segmentation — spike after low deltas = new episode
	// 2. Emotion deposit trigger — high delta = high salience moment
	// 3. Confidence weighting — accurate predictions = stronger proposals at motor
	//
	// Delta patterns over time:
	//   Shrinking across episode = learning this context
	//   Growing = model is stale, something changed
	//   Consistent highs on specific signal type = blind spot
	return r
}

func (c *Prediction) Realize(r *reality.Relation) *reality.Relation {
	c.Observe(r)
	r = c.Express(r)
	return r
}

func (c *Prediction) Competence(r *reality.Relation) {
}

func (c *Prediction) Activation(r *reality.Relation) float64 {
	return 1.0
}

func (c *Prediction) Reinforce() {
}

func (c *Prediction) Decay() {
}
