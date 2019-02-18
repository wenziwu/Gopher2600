package main_test

import (
	"fmt"
	"gopher2600/hardware"
	"gopher2600/hardware/cpu/result"
	"gopher2600/television"
	"gopher2600/television/sdltv"
	"testing"
)

func BenchmarkSDLTV(b *testing.B) {
	var err error

	tv, err := sdltv.NewSDLTV("NTSC", 1.0)
	if err != nil {
		panic(fmt.Errorf("error preparing television: %s", err))
	}

	err = tv.SetFeature(television.ReqSetVisibility, true)
	if err != nil {
		panic(fmt.Errorf("error preparing television: %s", err))
	}

	vcs, err := hardware.NewVCS(tv)
	if err != nil {
		panic(fmt.Errorf("error preparing VCS: %s", err))
	}

	err = vcs.AttachCartridge("roms/ROMs/Pitfall.bin")
	if err != nil {
		panic(err)
	}

	b.ResetTimer()

	for steps := 0; steps < b.N; steps++ {
		_, _, err = vcs.Step(func(*result.Instruction) error { return nil })
		if err != nil {
			panic(err)
		}
	}
}