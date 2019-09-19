package playmode

import (
	"fmt"
	"gopher2600/errors"
	"gopher2600/gui"
	"gopher2600/gui/sdl"
	"gopher2600/hardware"
	"gopher2600/recorder"
	"gopher2600/setup"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"
)

// Play sets the emulation running - without any debugging features
func Play(cartridgeFile, tvType string, scaling float32, stable bool, recording string, newRecording bool) error {
	if recorder.IsPlaybackFile(cartridgeFile) {
		return errors.New(errors.PlayError, "specified cartridge is a playback file. use -recording flag")
	}

	playtv, err := sdl.NewGUI(tvType, scaling, nil)
	if err != nil {
		return errors.New(errors.PlayError, err)
	}

	vcs, err := hardware.NewVCS(playtv)
	if err != nil {
		return errors.New(errors.PlayError, err)
	}

	// stk, err := sticks.NewSplaceStick()
	// if err != nil {
	// 	return errors.NewFormattedError(errors.PlayError, err)
	// }
	// vcs.Ports.Player0.Attach(stk)

	// create default recording file name if no name has been supplied
	if newRecording && recording == "" {
		shortCartName := path.Base(cartridgeFile)
		shortCartName = strings.TrimSuffix(shortCartName, path.Ext(cartridgeFile))
		n := time.Now()
		timestamp := fmt.Sprintf("%04d%02d%02d_%02d%02d%02d", n.Year(), n.Month(), n.Day(), n.Hour(), n.Minute(), n.Second())
		recording = fmt.Sprintf("recording_%s_%s", shortCartName, timestamp)
	}

	// note that we attach the cartridge in three different branches below - we
	// need to do this at different times depending on whether a new recording
	// or playback is taking place; or if it's just a regular playback

	if recording != "" {
		if newRecording {
			rec, err := recorder.NewRecorder(recording, vcs)
			if err != nil {
				return errors.New(errors.PlayError, err)
			}

			defer func() {
				rec.End()
			}()

			vcs.Ports.Player0.AttachTranscriber(rec)
			vcs.Ports.Player1.AttachTranscriber(rec)
			vcs.Panel.AttachTranscriber(rec)

			// attaching cartridge after recorder and transcribers have been
			// setup because we want to catch any setup events in the recording

			err = setup.AttachCartridge(vcs, cartridgeFile)
			if err != nil {
				return errors.New(errors.PlayError, err)
			}

		} else {
			plb, err := recorder.NewPlayback(recording)
			if err != nil {
				return err
			}

			if cartridgeFile != "" && cartridgeFile != plb.CartFile {
				return errors.New(errors.PlayError, "cartridge doesn't match name in the playback recording")
			}

			// not using setup.AttachCartridge. if the playback was recorded with setup
			// changes the events will have been copied into the playback script and
			// will be applied that way
			err = vcs.AttachCartridge(plb.CartFile)
			if err != nil {
				return errors.New(errors.PlayError, err)
			}

			// now that we've attached the cartridge check the hash against the
			// playback has (if it exists)
			if plb.CartHash != vcs.Mem.Cart.Hash {
				return errors.New(errors.PlayError, "cartridge hash doesn't match hash in playback recording")
			}

			err = plb.AttachToVCS(vcs)
			if err != nil {
				return errors.New(errors.PlayError, err)
			}
		}
	} else {
		err = setup.AttachCartridge(vcs, cartridgeFile)
		if err != nil {
			return errors.New(errors.PlayError, err)
		}
	}

	// connect gui
	guiChannel := make(chan gui.Event, 2)
	playtv.SetEventChannel(guiChannel)

	// request television visibility
	request := gui.ReqSetVisibilityStable
	if !stable {
		request = gui.ReqSetVisibility
	}
	err = playtv.SetFeature(request, true)
	if err != nil {
		return errors.New(errors.PlayError, err)
	}

	// we need to make sure we call the deferred function rec.End() even when
	// ctrl-c is pressed. redirect interrupt signal to an os.Signal channel
	intChan := make(chan os.Signal, 1)
	signal.Notify(intChan, os.Interrupt)

	// run and handle gui events
	err = vcs.Run(func() (bool, error) {
		select {
		case <-intChan:
			return false, nil
		case ev := <-guiChannel:
			switch ev.ID {
			case gui.EventWindowClose:
				return false, nil
			case gui.EventKeyboard:
				err = KeyboardEventHandler(ev.Data.(gui.EventDataKeyboard), playtv, vcs)
				return err == nil, err
			}
		default:
		}
		return true, nil
	})

	if err != nil {
		if errors.Is(err, errors.PowerOff) {
			return nil
		}
		return errors.New(errors.PlayError, err)
	}

	return nil
}
