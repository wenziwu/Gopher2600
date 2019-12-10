package regression

import (
	"bufio"
	"fmt"
	"gopher2600/cartridgeloader"
	"gopher2600/database"
	"gopher2600/digest"
	"gopher2600/errors"
	"gopher2600/hardware"
	"gopher2600/performance/limiter"
	"gopher2600/setup"
	"gopher2600/television"
	"io"
	"os"
	"strconv"
	"strings"
)

const digestEntryID = "digest"

type digestMode int

const (
	digestVideoOnly digestMode = iota
	digestAudioOnly
	digestVideoAndAudio
)

const (
	digestFieldMode int = iota
	digestFieldCartName
	digestFieldCartFormat
	digestFieldTVtype
	digestFieldNumFrames
	digestFieldState
	digestFieldDigest
	digestFieldNotes
	numDigestFields
)

// DigestRegression is the simplest regression type. it works by running the
// emulation for N frames and the digest recorded at that point. Regression
// passes if subsequenct runs produce the same digest value
type DigestRegression struct {
	mode      digestMode
	CartLoad  cartridgeloader.Loader
	TVtype    string
	NumFrames int
	State     bool
	stateFile string
	Notes     string
	digest    string
}

func deserialiseDigestEntry(fields database.SerialisedEntry) (database.Entry, error) {
	reg := &DigestRegression{}

	// basic sanity check
	if len(fields) > numDigestFields {
		return nil, errors.New(errors.RegressionDigestError, "too many fields")
	}
	if len(fields) < numDigestFields {
		return nil, errors.New(errors.RegressionDigestError, "too few fields")
	}

	// string fields need no conversion
	reg.CartLoad.Filename = fields[digestFieldCartName]
	reg.CartLoad.Format = fields[digestFieldCartFormat]
	reg.TVtype = fields[digestFieldTVtype]
	reg.digest = fields[digestFieldDigest]
	reg.Notes = fields[digestFieldNotes]

	var err error

	// convert mode field
	switch fields[digestFieldMode] {
	case "0":
		reg.mode = digestVideoOnly
	case "1":
		return nil, errors.New(errors.RegressionDigestError, "audio digesting not yet implemented")
	case "2":
		return nil, errors.New(errors.RegressionDigestError, "video & audio digesting not yet implemented")
	default:
		return nil, errors.New(errors.RegressionDigestError, "unrecognised mode")
	}

	// convert number of frames field
	reg.NumFrames, err = strconv.Atoi(fields[digestFieldNumFrames])
	if err != nil {
		msg := fmt.Sprintf("invalid numFrames field [%s]", fields[digestFieldNumFrames])
		return nil, errors.New(errors.RegressionDigestError, msg)
	}

	// handle state field
	if fields[digestFieldState] != "" {
		reg.State = true
		reg.stateFile = fields[digestFieldState]
	}

	return reg, nil
}

// ID implements the database.Entry interface
func (reg DigestRegression) ID() string {
	s := strings.Builder{}
	s.WriteString(digestEntryID)
	switch reg.mode {
	case digestVideoOnly:
		s.WriteString("/video")
	case digestAudioOnly:
		s.WriteString("/audio")
	case digestVideoAndAudio:
		s.WriteString("/video & audio")
	}
	return s.String()
}

// String implements the database.Entry interface
func (reg DigestRegression) String() string {
	s := strings.Builder{}
	stateFile := ""
	if reg.State {
		stateFile = "[with state]"
	}
	s.WriteString(fmt.Sprintf("[%s] %s [%s] frames=%d %s", reg.ID(), reg.CartLoad.ShortName(), reg.TVtype, reg.NumFrames, stateFile))
	if reg.Notes != "" {
		s.WriteString(fmt.Sprintf(" [%s]", reg.Notes))
	}
	return s.String()
}

// Serialise implements the database.Entry interface
func (reg *DigestRegression) Serialise() (database.SerialisedEntry, error) {
	return database.SerialisedEntry{
			strconv.Itoa(int(reg.mode)),
			reg.CartLoad.Filename,
			reg.CartLoad.Format,
			reg.TVtype,
			strconv.Itoa(reg.NumFrames),
			reg.stateFile,
			reg.digest,
			reg.Notes,
		},
		nil
}

// CleanUp implements the database.Entry interface
func (reg DigestRegression) CleanUp() error {
	err := os.Remove(reg.stateFile)
	if _, ok := err.(*os.PathError); ok {
		return nil
	}
	return err
}

// regress implements the regression.Regressor interface
func (reg *DigestRegression) regress(newRegression bool, output io.Writer, msg string) (bool, string, error) {
	output.Write([]byte(msg))

	tv, err := television.NewTelevision(reg.TVtype)
	if err != nil {
		return false, "", errors.New(errors.RegressionDigestError, err)
	}
	defer tv.End()

	dig, err := digest.NewVideo(tv)
	if err != nil {
		return false, "", errors.New(errors.RegressionDigestError, err)
	}

	vcs, err := hardware.NewVCS(dig)
	if err != nil {
		return false, "", errors.New(errors.RegressionDigestError, err)
	}

	err = setup.AttachCartridge(vcs, reg.CartLoad)
	if err != nil {
		return false, "", errors.New(errors.RegressionDigestError, err)
	}

	// list of state information. we'll either save this in the event of
	// newRegression being true; or we'll use it to compare to the entries in
	// the specified state file
	state := make([]string, 0, 1024)

	// display progress meter every 1 second
	limiter, err := limiter.NewFPSLimiter(1)
	if err != nil {
		return false, "", errors.New(errors.RegressionDigestError, err)
	}

	// add the starting state of the tv
	if reg.State {
		state = append(state, tv.String())
	}

	// run emulation
	err = vcs.RunForFrameCount(reg.NumFrames, func(frame int) (bool, error) {
		if limiter.HasWaited() {
			output.Write([]byte(fmt.Sprintf("\r%s[%d/%d (%.1f%%)]", msg, frame, reg.NumFrames, 100*(float64(frame)/float64(reg.NumFrames)))))
		}

		// store tv state at every step
		if reg.State {
			state = append(state, tv.String())
		}

		return true, nil
	})

	if err != nil {
		return false, "", errors.New(errors.RegressionDigestError, err)
	}

	if newRegression {
		reg.digest = dig.Hash()

		if reg.State {
			// create a unique filename
			reg.stateFile = uniqueFilename("state", reg.CartLoad)

			// check that the filename is unique
			nf, _ := os.Open(reg.stateFile)

			// no need to bother with returned error. nf tells us everything we
			// need
			if nf != nil {
				msg := fmt.Sprintf("state recording file already exists (%s)", reg.stateFile)
				return false, "", errors.New(errors.RegressionDigestError, msg)
			}
			nf.Close()

			// create new file
			nf, err = os.Create(reg.stateFile)
			if err != nil {
				msg := fmt.Sprintf("error creating state recording file: %s", err)
				return false, "", errors.New(errors.RegressionDigestError, msg)
			}
			defer nf.Close()

			for i := range state {
				s := fmt.Sprintf("%s\n", state[i])
				if n, err := nf.WriteString(s); err != nil || len(s) != n {
					msg := fmt.Sprintf("error writing state recording file: %s", err)
					return false, "", errors.New(errors.RegressionDigestError, msg)
				}
			}
		}

		return true, "", nil
	}

	// if we reach this point then this is a regression test (not adding a new
	// test)

	// compare new state tracking with recorded state tracking
	if reg.State {
		nf, err := os.Open(reg.stateFile)
		if err != nil {
			msg := fmt.Sprintf("old state recording file not present (%s)", reg.stateFile)
			return false, "", errors.New(errors.RegressionDigestError, msg)
		}
		defer nf.Close()

		reader := bufio.NewReader(nf)

		for i := range state {
			s, _ := reader.ReadString('\n')
			s = strings.TrimRight(s, "\n")

			// ignore blank lines
			if s == "" {
				continue
			}

			if s != state[i] {
				failm := fmt.Sprintf("state mismatch line %d: expected %s (%s)", i, s, state[i])
				return false, failm, nil
			}
		}

		// check that we've consumed all the lines in the recorded state file
		_, err = reader.ReadString('\n')
		if err == nil || err != io.EOF {
			failm := "unexpected end of state. entries remaining in recorded state file"
			return false, failm, nil
		}

	}

	if dig.Hash() != reg.digest {
		return false, "digest mismatch", nil
	}

	return true, "", nil
}
