package memorymap

import (
	"fmt"
	"strings"
)

// Summary builds a string detailing all the areas in memory
func Summary() string {
	var area, current Area
	var a, sa uint16

	s := strings.Builder{}

	// look up area of first address in memory
	_, current = MapAddress(uint16(0), true)

	// for every address in the range 0 to MemtopCart...
	for a = uint16(1); a <= MemtopCart; a++ {
		// ...get the area name of that address.
		_, area = MapAddress(a, true)

		// if the area has changed print out the summary line...
		if area != current {
			s.WriteString(fmt.Sprintf("%04x -> %04x\t%s\n", sa, a-uint16(1), current.String()))

			// ...update current area and start address of the area
			current = area
			sa = a
		}
	}

	// write last line of summary
	s.WriteString(fmt.Sprintf("%04x -> %04x\t%s\n", sa, a-uint16(1), area.String()))

	return s.String()
}