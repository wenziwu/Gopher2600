package memory

import (
	"fmt"
	"gopher2600/errors"
	"io"
	"strings"
)

// from bankswitch_sizes.txt:
//
// -E7: Only M-Network used this scheme.  This has to be the most complex
// method used in any cart! :-)  It allows for the capability of 2K of RAM;
// although it doesn't have to be used (in fact, only one cart used it-
// Burgertime).  This is similar to the 3F type with a few changes.  There are
// now 8 2K banks, instead of 4.  The last 2K in the cart always points to the
// last 2K of the ROM image, while the first 2K is selectable.  You access 1FE0
// to 1FE6 to select which 2K bank. Note that you cannot select the last 2K of
// the ROM image into the lower 2K of the cart!  Accessing 1FE7 selects 1K of
// RAM at 1000-17FF instead of ROM!  The 2K of RAM is broken up into two 1K
// sections.  One 1K section is mapped in at 1000-17FF if 1FE7 has been
// accessed.  1000-13FF is the write port, while 1400-17FF is the read port.
// The second 1K of RAM appears at 1800-19FF.  1800-18FF is the write port
// while 1900-19FF is the read port.  You select which 256 byte block appears
// here by accessing 1FF8 to 1FFB.

func fingerprintMnetwork(b []byte) bool {
	threshold := 2
	for i := 0; i < len(b)-3; i++ {
		if b[i] == 0x7e && b[i+1] == 0x66 && b[i+2] == 0x66 && b[i+3] == 0x66 {
			threshold--
		}
		if threshold == 0 {
			return true
		}
	}

	return false
}

type mnetwork struct {
	method string
	banks  [][]uint8

	// m-network cartridges divide memory into two 2k segments. the upper
	// segment always points to the the last bank so we only need to keep track
	// of which bank the lower segment is pointing to
	//
	// a lowerSegment of 7 means that tha last bank has been paged into the first
	// lowerSegment *BUT* that the lower 1K points to ramLower
	lowerSegment int
	upperSegment int

	//  o ramLower is read through addresses 0x1000 to 0x13ff and written
	//  through addresses 0x1400 to 0x17ff **WHEN LOWER SEGMENT POINTS TO BANK 7**
	//
	//  o ramUpper is read through addresses 0x1900 to 0x19dd and written
	//  through address 0x1800 to 0x18ff **IN ALL CASES**
	//
	//  o the ramUpperSegment which is read however can be changed
	//
	// (addresses quoted above are of course masked so that they fall into the
	// allocation range)
	//
	// switching of both segment pointers is performed by the function
	// bankSwitchOnAccess()
	ramLower [1024]uint8
	ramUpper [4][256]uint8

	// (not all m-network cartridges have any RAM but we'll allocate it for all
	// instances)
}

func newMnetwork(cf io.ReadSeeker) (cartMapper, error) {
	cart := &mnetwork{method: "m-network (E7)"}

	cart.banks = make([][]uint8, cart.numBanks())

	cf.Seek(0, io.SeekStart)

	for k := 0; k < cart.numBanks(); k++ {
		const bankSize = 2048
		cart.banks[k] = make([]uint8, bankSize)

		// read cartridge
		n, err := cf.Read(cart.banks[k])
		if err != nil {
			return nil, err
		}
		if n != bankSize {
			return nil, errors.New(errors.CartridgeFileError, "not enough bytes in the cartridge file")
		}
	}

	cart.initialise()

	return cart, nil
}

func (cart mnetwork) String() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("%s Banks: %d", cart.method, cart.lowerSegment))
	if cart.lowerSegment == 7 {
		s.WriteString("+lower RAM")
	}
	s.WriteString(fmt.Sprintf(", upper RAM bank %d", cart.upperSegment))
	return s.String()
}

func (cart *mnetwork) initialise() {
	cart.lowerSegment = cart.numBanks() - 2
	cart.upperSegment = 0
}

func (cart *mnetwork) read(addr uint16) (uint8, error) {
	var data uint8

	if addr >= 0x0000 && addr <= 0x07ff {
		if cart.lowerSegment == 7 && addr >= 0x0400 {
			data = cart.ramLower[addr&0x03ff]
		} else {
			data = cart.banks[cart.lowerSegment][addr&0x07ff]
		}
	} else if addr >= 0x0800 && addr <= 0x0fff {
		if addr >= 0x0900 && addr <= 0x09ff {
			// access upper 1k of ram if cart.segment is pointing to ram and
			// the address is in the write range
			data = cart.ramUpper[cart.upperSegment][addr&0x00ff]
		} else {
			// if address is not in ram space then read from the last rom bank
			data = cart.banks[cart.numBanks()-1][addr&0x07ff]
			cart.bankSwitchOnAccess(addr)
		}
	} else {
		return 0, errors.New(errors.UnreadableAddress, addr)
	}

	return data, nil
}

func (cart *mnetwork) write(addr uint16, data uint8) error {
	if addr >= 0x0000 && addr <= 0x07ff {
		if addr <= 0x03ff && cart.lowerSegment == 7 {
			cart.ramLower[addr&0x03ff] = data
			return nil
		}
	} else if addr >= 0x0800 && addr <= 0x08ff {
		cart.ramUpper[cart.upperSegment][addr&0x00ff] = data
		return nil
	} else if cart.bankSwitchOnAccess(addr) {
		return nil
	}

	return errors.New(errors.UnwritableAddress, addr)
}

func (cart *mnetwork) bankSwitchOnAccess(addr uint16) bool {
	switch addr {
	case 0x0fe0:
		cart.lowerSegment = 0
	case 0x0fe1:
		cart.lowerSegment = 1
	case 0x0fe2:
		cart.lowerSegment = 2
	case 0x0fe3:
		cart.lowerSegment = 3
	case 0x0fe4:
		cart.lowerSegment = 4
	case 0x0fe5:
		cart.lowerSegment = 5
	case 0x0fe6:
		cart.lowerSegment = 6

		// from bankswitch_sizes.txt: "Note that you cannot select the last 2K
		// of the ROM image into the lower 2K of the cart!  Accessing 1FE7
		// selects 1K of RAM at 1000-17FF instead of ROM!"
		//
		// we're using bank number -1 to indicate the use of RAM
	case 0x0fe7:
		cart.lowerSegment = 7

		// from bankswitch_size.txt: "You select which 256 byte block appears
		// here by accessing 1FF8 to 1FFB."
		//
		// "here" refers to the read range 0x0900 to 0x09ff and the write range
		// 0x0800 to 0x08ff
	case 0x0ff8:
		cart.upperSegment = 0
	case 0x0ff9:
		cart.upperSegment = 1
	case 0x0ffa:
		cart.upperSegment = 2
	case 0x0ffb:
		cart.upperSegment = 3

	default:
		return false
	}

	return true
}

func (cart *mnetwork) numBanks() int {
	return 8 // eight banks of 2k
}

func (cart *mnetwork) getBank(addr uint16) (bank int) {
	if addr >= 0x0000 && addr <= 0x07ff {
		return cart.lowerSegment
	}
	return cart.numBanks() - 1
}

func (cart *mnetwork) setBank(addr uint16, bank int) error {
	if bank < 0 || bank > cart.numBanks() {
		return errors.New(errors.CartridgeError, fmt.Sprintf("invalid bank (%d) for cartridge type (%s)", bank, cart.method))
	}

	if addr >= 0x0000 && addr <= 0x07ff {
		cart.lowerSegment = bank
	} else if addr >= 0x0800 && addr <= 0x0fff {
		// last segment always points to the last bank
	} else {
		return errors.New(errors.CartridgeError, fmt.Sprintf("invalid address (%d) for cartridge type (%s)", bank, cart.method))
	}

	return nil
}

func (cart *mnetwork) saveState() interface{} {
	// !!TODO: ram
	return cart.lowerSegment
}

func (cart *mnetwork) restoreState(state interface{}) error {
	// !!TODO: ram
	cart.lowerSegment = state.(int)
	return nil
}

func (cart *mnetwork) ram() []uint8 {
	return []uint8{}
}

func (cart *mnetwork) listen(addr uint16, data uint8) error {
	return errors.New(errors.CartridgeListen, addr)
}