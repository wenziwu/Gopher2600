// This file is part of Gopher2600.
//
// Gopher2600 is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gopher2600 is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gopher2600.  If not, see <https://www.gnu.org/licenses/>.

package debugger

// support functions for the rewind package that require more knowledge of
// the debugger than would otherwise be available.

import (
	"github.com/jetsetilly/gopher2600/disassembly"
	"github.com/jetsetilly/gopher2600/gui"
	"github.com/jetsetilly/gopher2600/logger"
)

// CatchupLoop is an implementation of the rewind.Runner interface.
//
// Runs the emulation from it's current state until the supplied continueCheck
// callback function returns false.
func (dbg *Debugger) CatchUpLoop(continueCheck func() bool) error {
	var err error

	dbg.lastBank = dbg.VCS.Mem.Cart.GetBank(dbg.VCS.CPU.PC.Address())
	dbg.lastResult, err = dbg.Disasm.FormatResult(dbg.lastBank, dbg.VCS.CPU.LastResult, disassembly.EntryLevelExecuted)
	if err != nil {
		return nil
	}

	for continueCheck() {
		err = dbg.VCS.Step(func() error {
			return dbg.reflect.Check(dbg.lastBank)
		})
		if err != nil {
			return err
		}

		dbg.lastBank = dbg.VCS.Mem.Cart.GetBank(dbg.VCS.CPU.PC.Address())
		dbg.lastResult, err = dbg.Disasm.FormatResult(dbg.lastBank, dbg.VCS.CPU.LastResult, disassembly.EntryLevelExecuted)
		if err != nil {
			return err
		}
	}

	return nil
}

// PushRewind is a special case of PushRawEvent(). It prevents too many pushed
// Rewind.Goto*() function calls. To be used from the GUI thread.
func (dbg *Debugger) PushRewind(fn int, last bool) bool {
	select {
	case dbg.rewinding <- true:
	default:
		return true
	}

	dbg.PushRawEvent(func() {
		defer func() {
			<-dbg.rewinding
		}()

		dbg.scr.ReqFeature(gui.ReqState, gui.StateRewinding)
		if last {
			err := dbg.Rewind.GotoLast()
			if err != nil {
				logger.Log("debugger", err.Error())
			}
		} else {
			err := dbg.Rewind.GotoFrame(fn)
			if err != nil {
				logger.Log("debugger", err.Error())
			}
		}
		dbg.scr.ReqFeature(gui.ReqState, gui.StatePaused)
	})

	return false
}

// PushGotoCoords is a special case of PushRawEvent(). It wraps a pushed call
// to rewind.GotoFrameCoords() in gui.ReqRewinding true/false.
func (dbg *Debugger) PushGotoCoords(scanline int, horizpos int) {
	dbg.PushRawEvent(func() {
		dbg.scr.ReqFeature(gui.ReqState, gui.StateGotoCoords)
		err := dbg.Rewind.GotoFrameCoords(scanline, horizpos)
		if err != nil {
			logger.Log("debugger", err.Error())
		}
		dbg.scr.ReqFeature(gui.ReqState, gui.StatePaused)
	})
}
