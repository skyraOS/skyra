package placegateway

import "skyra-v1/src/skyra"

// PlaceRef is the smallest thing the gateway needs to know about a target
// place. Keep it small until the place model settles.
type PlaceRef struct {
	ID   string
	Kind string
}

// Handler is the first-pass gateway contract. The runtime addresses a place
// and sends an impulse; the gateway routes that combination to the correct
// backing adapter or helper.
type Handler interface {
	Handle(place PlaceRef, impulse skyra.Impulse) (any, error)
}

// PrimitiveHandler is implemented by adapter helpers that accept one primitive
// for one family of places.
type PrimitiveHandler interface {
	HandlePrimitive(place PlaceRef, impulse skyra.Impulse) (any, error)
}

