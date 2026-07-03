package world

import "math/rand"

const (
	MazeOriginX   = 478
	MazeOriginY   = 478
	MazeCellsW    = 20
	MazeCellsH    = 20
	MazeEntranceX = 479                          // first passage cell (MazeOriginX+1, MazeOriginY+1)
	MazeEntranceY = 479                          //
	MazeExitX     = MazeOriginX + 2*MazeCellsW - 1 // last passage cell column (MazeOriginX+39)
	MazeExitY     = MazeOriginY + 2*MazeCellsH     // bottom boundary wall row (MazeOriginY+40)
)

// GenerateMaze fills the maze region with walls and paths using a recursive
// backtracker. The maze occupies a (2*W+1)×(2*H+1) world-coord block anchored
// at (MazeOriginX, MazeOriginY). Cell (cx,cy) maps to world coord
// (MazeOriginX+1+2*cx, MazeOriginY+1+2*cy). The outer boundary stays walled
// except for an entrance gap carved above cell (0,0).
func (w *World) GenerateMaze(seed int64) {
	rng := rand.New(rand.NewSource(seed))

	totalW := 2*MazeCellsW + 1
	totalH := 2*MazeCellsH + 1

	// Build wall/path sets locally then commit once — avoids lock contention
	// during the recursive traversal.
	walls := make(map[coord]bool, totalW*totalH)
	paths := make(map[coord]bool, MazeCellsW*MazeCellsH*2)

	for dy := 0; dy < totalH; dy++ {
		for dx := 0; dx < totalW; dx++ {
			walls[w.wrapCoord(MazeOriginX+dx, MazeOriginY+dy)] = true
		}
	}

	visited := make([][]bool, MazeCellsH)
	for i := range visited {
		visited[i] = make([]bool, MazeCellsW)
	}

	carve := func(wx, wy int) {
		c := w.wrapCoord(wx, wy)
		delete(walls, c)
		paths[c] = true
	}

	var backtrack func(cx, cy int)
	backtrack = func(cx, cy int) {
		visited[cy][cx] = true
		carve(MazeOriginX+1+2*cx, MazeOriginY+1+2*cy)

		// Shuffle a fresh copy of directions each call so the rng drives the shape.
		dirs := [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
		rng.Shuffle(4, func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })

		for _, d := range dirs {
			ncx, ncy := cx+d[0], cy+d[1]
			if ncx < 0 || ncx >= MazeCellsW || ncy < 0 || ncy >= MazeCellsH || visited[ncy][ncx] {
				continue
			}
			// carve the wall between (cx,cy) and (ncx,ncy)
			carve(MazeOriginX+1+2*cx+d[0], MazeOriginY+1+2*cy+d[1])
			backtrack(ncx, ncy)
		}
	}

	backtrack(0, 0)

	// Entrance: open the top boundary wall above cell (0,0).
	carve(MazeEntranceX, MazeOriginY)
	// Exit: open the bottom boundary wall below the last cell (W-1, H-1).
	carve(MazeExitX, MazeExitY)

	w.mu.Lock()
	for c := range walls {
		w.walls[c] = true
	}
	for c := range paths {
		w.paths[c] = true
	}
	w.mu.Unlock()
}
