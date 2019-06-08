package memory

import (
	"fmt"
	"gopher2600/errors"
)

const ejectedName = "ejected"
const ejectedHash = "nohash"
const ejectedMethod = "ejected"

// ejected implements the cartMapper interface.

type ejected struct {
	method string
}

func newEjected() *ejected {
	cart := &ejected{method: ejectedMethod}
	cart.initialise()
	return cart
}

func (cart ejected) String() string {
	return cart.method
}

func (cart *ejected) initialise() {
}

func (cart *ejected) read(addr uint16) (uint8, error) {
	return 0, errors.NewFormattedError(errors.CartridgeEjected)
}

func (cart *ejected) write(addr uint16, data uint8, isPoke bool) error {
	return errors.NewFormattedError(errors.CartridgeEjected)
}

func (cart ejected) numBanks() int {
	return 0
}

func (cart *ejected) setAddressBank(addr uint16, bank int) error {
	return errors.NewFormattedError(errors.CartridgeError, fmt.Sprintf("invalid bank (%d) for cartridge type (%s)", bank, cart.method))
}

func (cart ejected) getAddressBank(addr uint16) int {
	return 0
}

func (cart *ejected) saveBanks() interface{} {
	return nil
}

func (cart *ejected) restoreBanks(state interface{}) error {
	return nil
}