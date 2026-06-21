package dna

// WorkingMemory — conversation buffer.
// Indexed by EpisodeID. Decays fast. Small capacity.
// Lives on the agent's Brain. Reference passed to Slice at boot.
// Raw recent exchange before absorption into traces.
