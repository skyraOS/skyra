package dna

func ptrObservation(value Observation) *Observation { return &value }

func parseProbe(raw string) ([]Move, error) {
	return (&RootDriver{}).Egress(raw)
}
