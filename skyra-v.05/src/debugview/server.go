package debugview

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"skyra-v1/src/debug"
	"skyra-v1/src/reality"
)

type NodeView struct {
	ID          string  `json:"id"`
	BeingID     string  `json:"beingId"`
	Type        string  `json:"type"`
	Content     string  `json:"content"`
	Description string  `json:"description"`
	Identity    float64 `json:"identity"`
	Familiarity float64 `json:"familiarity"`
	Recency     float64 `json:"recency"`
	Activation  float64 `json:"activation"`
	Traversals  int     `json:"traversals"`
	Children    int     `json:"children"`
}

type SliceView struct {
	BeingID      string             `json:"beingId"`
	Signal       float64            `json:"signal"`
	Load         int                `json:"load"`
	ThoughtCount int                `json:"thoughtCount"`
	Goal         string             `json:"goal"`
	Intent       string             `json:"intent"`
	Surroundings map[string]float64 `json:"surroundings"`
}

type ThreadView struct {
	ID      string             `json:"id"`
	BeingID string             `json:"beingId"`
	Members map[string]float64 `json:"members"`
}

type WorldView struct {
	Entities []NodeView   `json:"entities"`
	Slices   []SliceView  `json:"slices"`
	Threads  []ThreadView `json:"threads"`
	Origin   string       `json:"origin"`
	Impulse  string       `json:"impulse"`
	ThreadID string       `json:"threadId"`
}

func sliceViewFrom(s reality.SliceSnapshot) SliceView {
	return SliceView{
		BeingID:      s.BeingID,
		Signal:       s.Signal,
		Load:         s.Load,
		ThoughtCount: s.ThoughtCount,
		Goal:         s.Goal,
		Intent:       s.Intent,
		Surroundings: s.Surroundings,
	}
}

func Start(addr string, world *reality.Relation) {
	http.HandleFunc("/", handleUI)
	http.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		handleState(w, r, world)
	})

	debug.Log("debug view starting on", addr)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			debug.Log("debug view error:", err)
		}
	}()
}

func handleState(w http.ResponseWriter, r *http.Request, world *reality.Relation) {
	view := WorldView{
		Entities: []NodeView{},
		Slices:   []SliceView{},
		Threads:  []ThreadView{},
	}

	world.RLock()
	view.Origin = world.Origin
	view.Impulse = world.Impulse
	view.ThreadID = world.ThreadID

	for name, entity := range world.World {
		nv := nodeFromReality(name, entity, world)
		view.Entities = append(view.Entities, nv)

		if sk, ok := entity.(*reality.Skyra); ok {
			walkTopology(sk, world, &view.Entities, 0)
		}
	}

	for _, s := range world.SliceList() {
		view.Slices = append(view.Slices, sliceViewFrom(s))
	}

	for _, t := range world.ThreadList() {
		view.Threads = append(view.Threads, ThreadView{
			ID:      t.ID,
			BeingID: t.BeingID,
			Members: t.Members,
		})
	}
	world.RUnlock()

	sort.Slice(view.Entities, func(i, j int) bool {
		return view.Entities[i].Activation > view.Entities[j].Activation
	})

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(view)
}

func nodeFromReality(name string, entity reality.Reality, world *reality.Relation) NodeView {
	nv := NodeView{
		ID:         entity.ID(),
		Activation: entity.Activation(world),
	}

	if sk, ok := entity.(*reality.Skyra); ok {
		nv.BeingID = sk.BeingID
		nv.Type = sk.Type
		nv.Content = sk.Content
		nv.Description = sk.Description
		nv.Identity = sk.IdentityWeight
		nv.Familiarity = sk.FamiliarityWeight
		nv.Recency = reality.Recency(sk.TraversalCount, sk.LastTraversed)
		nv.Traversals = sk.TraversalCount
		nv.Children = len(sk.Relationships)
	} else if _, ok := entity.(*reality.User); ok {
		nv.Type = "user"
		nv.Description = name
	} else if t, ok := entity.(*reality.Thread); ok {
		nv.Type = "thread"
		nv.Description = t.ID()
	}

	return nv
}

func walkTopology(sk *reality.Skyra, world *reality.Relation, nodes *[]NodeView, depth int) {
	if depth > 5 {
		return
	}
	for name, rel := range sk.Relationships {
		if _, ok := rel.(*reality.Thread); ok {
			continue
		}
		nv := nodeFromReality(name, rel, world)
		*nodes = append(*nodes, nv)

		if child, ok := rel.(*reality.Skyra); ok && depth < 5 {
			walkTopology(child, world, nodes, depth+1)
		}
	}
}

func handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, uiHTML)
}

const uiHTML = `<!DOCTYPE html>
<html>
<head>
<title>skyra debug</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { background: #0a0a0a; color: #ccc; font-family: monospace; font-size: 13px; padding: 16px; }
h1 { color: #5af; font-size: 16px; margin-bottom: 12px; }
h2 { color: #a8f; font-size: 14px; margin: 16px 0 8px; }
.grid { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 16px; }
.panel { background: #111; border: 1px solid #222; border-radius: 6px; padding: 12px; max-height: 80vh; overflow-y: auto; }
.node { padding: 6px 8px; margin: 2px 0; border-left: 3px solid #333; background: #151515; border-radius: 2px; }
.node:hover { background: #1a1a1a; }
.node .id { color: #5af; font-weight: bold; }
.node .type { color: #888; font-size: 11px; }
.node .content { color: #aaa; font-size: 11px; max-height: 40px; overflow: hidden; }
.node .signals { color: #666; font-size: 11px; }
.signal { display: inline-block; margin-right: 8px; }
.signal .label { color: #888; }
.signal .value { color: #fc5; }
.high { border-left-color: #5f5; }
.medium { border-left-color: #fc5; }
.low { border-left-color: #555; }
.slice { padding: 8px; margin: 4px 0; background: #151515; border-radius: 4px; }
.slice .name { color: #5af; font-weight: bold; }
.thread { padding: 8px; margin: 4px 0; background: #151515; border-radius: 4px; }
.thread .name { color: #a8f; font-weight: bold; }
.member { color: #888; font-size: 11px; padding: 2px 0; }
.meta { color: #555; margin-bottom: 12px; }
</style>
</head>
<body>
<h1>skyra debug</h1>
<div class="meta" id="meta"></div>
<div class="grid">
  <div class="panel" id="topology"><h2>Topology</h2></div>
  <div class="panel" id="slices"><h2>Slices</h2></div>
  <div class="panel" id="threads"><h2>Threads</h2></div>
</div>
<script>
function cls(activation) {
  if (activation > 0.5) return 'high';
  if (activation > 0.1) return 'medium';
  return 'low';
}

function render(data) {
  document.getElementById('meta').textContent =
    'origin: ' + data.origin + ' | impulse: ' + (data.impulse || '(none)') + ' | thread: ' + data.threadId;

  let topo = '<h2>Topology (' + data.entities.length + ' nodes)</h2>';
  for (const n of data.entities) {
    topo += '<div class="node ' + cls(n.activation) + '">' +
      '<div class="id">' + n.id + ' <span class="type">' + (n.type || '') + '</span></div>' +
      (n.content ? '<div class="content">' + n.content.substring(0, 120) + '</div>' : '') +
      (n.description ? '<div class="content">' + n.description + '</div>' : '') +
      '<div class="signals">' +
        '<span class="signal"><span class="label">act:</span> <span class="value">' + n.activation.toFixed(4) + '</span></span>' +
        '<span class="signal"><span class="label">id:</span> <span class="value">' + n.identity.toFixed(4) + '</span></span>' +
        '<span class="signal"><span class="label">fam:</span> <span class="value">' + n.familiarity.toFixed(4) + '</span></span>' +
        '<span class="signal"><span class="label">rec:</span> <span class="value">' + n.recency.toFixed(4) + '</span></span>' +
        '<span class="signal"><span class="label">trav:</span> <span class="value">' + n.traversals + '</span></span>' +
        '<span class="signal"><span class="label">children:</span> <span class="value">' + n.children + '</span></span>' +
      '</div>' +
    '</div>';
  }
  document.getElementById('topology').innerHTML = topo;

  let sl = '<h2>Slices (' + data.slices.length + ')</h2>';
  for (const s of data.slices) {
    sl += '<div class="slice">' +
      '<div class="name">' + s.beingId + '</div>' +
      '<div class="signals">' +
        '<span class="signal"><span class="label">signal:</span> <span class="value">' + s.signal.toFixed(4) + '</span></span>' +
        '<span class="signal"><span class="label">load:</span> <span class="value">' + s.load + '</span></span>' +
        '<span class="signal"><span class="label">thoughts:</span> <span class="value">' + s.thoughtCount + '</span></span>' +
      '</div>' +
      (s.goal ? '<div class="content">goal: ' + s.goal + '</div>' : '') +
      (s.intent ? '<div class="content">intent: ' + s.intent + '</div>' : '') +
    '</div>';
    if (s.surroundings && Object.keys(s.surroundings).length > 0) {
      sl += '<div class="member">surroundings:</div>';
      for (const [k,v] of Object.entries(s.surroundings)) {
        sl += '<div class="member">  ' + k + ': ' + v.toFixed(4) + '</div>';
      }
    }
  }
  document.getElementById('slices').innerHTML = sl;

  let th = '<h2>Threads (' + data.threads.length + ')</h2>';
  for (const t of data.threads) {
    th += '<div class="thread">' +
      '<div class="name">' + t.id + ' <span class="type">(' + t.beingId + ')</span></div>';
    if (t.members) {
      for (const [k,v] of Object.entries(t.members)) {
        th += '<div class="member">' + k + ': ' + v.toFixed(4) + '</div>';
      }
    }
    th += '</div>';
  }
  document.getElementById('threads').innerHTML = th;
}

function poll() {
  fetch('/api/state')
    .then(r => r.json())
    .then(render)
    .catch(e => console.error(e));
}

poll();
setInterval(poll, 1000);
</script>
</body>
</html>`
