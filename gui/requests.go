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

package gui

import "github.com/jetsetilly/gopher2600/hardware/memory/cartridge/plusrom"

// FeatureReq is used to request the setting of a gui attribute
// eg. toggling the overlay.
type FeatureReq string

// FeatureReqData represents the information associated with a FeatureReq. See
// commentary for the defined FeatureReq values for the underlying type.
type FeatureReqData interface{}

// EmulationState indicates to the GUI that the debugger is in a particular
// state.
type EmulationState int

// List of valid emulation states.
const (
	StatePaused EmulationState = iota
	StateRunning
	StateRewinding
	StateGotoCoords
)

// List of valid feature requests. argument must be of the type specified or
// else the interface{} type conversion will fail and the application will
// probably crash.
//
// Note that, like the name suggests, these are requests, they may or may not
// be satisfied depending other conditions in the GUI.
const (
	// visibility can be interpreted by the gui implementation in different
	// ways. at it's simplest it should set the visibility of the TV screen.
	ReqSetVisibility    FeatureReq = "ReqSetVisibility"    // bool
	ReqToggleVisibility FeatureReq = "ReqToggleVisibility" // none

	// notify GUI of emulation state. the GUI should use this to alter how
	// infomration, particularly the display of the PixelRenderer.
	ReqState FeatureReq = "ReqState" // EmulationState

	// whether gui should use vsync or not. not all gui modes have to obey this
	// but for presentation/play modes it's a good idea to have it set.
	ReqVSync FeatureReq = "ReqVSync" // bool

	// the following requests should set or toggle visual elements of the debugger.
	ReqSetDbgColors    FeatureReq = "ReqSetDbgColors"    // bool
	ReqToggleDbgColors FeatureReq = "ReqToggleDbgColors" // none
	ReqSetCropping     FeatureReq = "ReqSetCropping"     // bool
	ReqToggleCropping  FeatureReq = "ReqToggleCropping"  // none
	ReqSetOverlay      FeatureReq = "ReqSetOverlay"      // bool
	ReqToggleOverlay   FeatureReq = "ReqToggleOverlay"   // none
	ReqCRTeffects      FeatureReq = "ReqCRTeffects"      // bool

	// the add VCS request is used to associate the gui with an emulated VCS.
	// a debugger does not need to send this request if it already sends a
	// ReqAddDebugger request (which it should).
	ReqAddVCS FeatureReq = "ReqAddVCS" // *hardware.VCS

	// the add debugger request must be made by the debugger if debug access to
	// the the machine is required by the GUI.
	ReqAddDebugger FeatureReq = "ReqAddDebugger" // *debugger.Debugger

	// the event channel is used to by the GUI implementation to send
	// information back to the main program. the GUI may or may not be in its
	// own go routine but regardless, the event channel is used for this
	// purpose.
	ReqSetEventChan FeatureReq = "ReqSetEventChan" // chan gui.Event()

	// playmode is called whenever the play/debugger looper is changed. like
	// all other requests this may not do anything, depending on the GUI
	// specifics.
	ReqSetPlaymode FeatureReq = "ReqSetPlaymode" // bool

	// trigger a save preferences event. usually performed before gui is
	// destroyed or before some other destructive action.
	ReqSavePrefs FeatureReq = "ReqSavePrefs" // none

	// triggered when cartridge is being change.
	ReqChangingCartridge FeatureReq = "ReqChangingCartridge" // bool

	// special request for PlusROM cartridges.
	ReqPlusROMFirstInstallation FeatureReq = "ReqPlusROMFirstInstallation" // PlusROMFirstInstallation
)

// PlusROMFirstInstallation is used to pass information to the GUI as part of
// the request.
type PlusROMFirstInstallation struct {
	Finish chan error
	Cart   *plusrom.PlusROM
}
