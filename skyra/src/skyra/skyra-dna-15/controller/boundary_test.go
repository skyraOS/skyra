package controller

import (
	"testing"

	"skyra-v1/src/skyra/skyra-dna-15/driver"
)

// TestControllerHoldsOnlyReadDoor is the in-package white-box counterpart to the
// black-box capability-separation checks in ctu/boot_test.go. The concrete
// *DriverRegistry implements both Resolver and Registrar, but the controller is only
// ever handed the narrowed resolver adapter. Verify from inside the package — where the
// private field is visible — that the controller's internal registry exposes no write
// door: it cannot be asserted to a driver.Registrar, nor is it the concrete registry.
func TestControllerHoldsOnlyReadDoor(t *testing.T) {
	r := driver.NewDriverRegistry()
	cc, err := NewContinuationController(r.Resolver())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cc.registry.(driver.Registrar); ok {
		t.Fatal("controller's internal resolver exposes registration")
	}
	if _, ok := cc.registry.(*driver.DriverRegistry); ok {
		t.Fatal("controller holds the concrete registry, not a narrowed resolver")
	}
}
