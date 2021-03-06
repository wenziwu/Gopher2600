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

package regression

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"time"

	"github.com/jetsetilly/gopher2600/curated"
	"github.com/jetsetilly/gopher2600/database"
	"github.com/jetsetilly/gopher2600/paths"
)

// ansi code for clear line.
const ansiClearLine = "\033[2K"

// the location of the regressionDB file and the location of any regression
// scripts. these should be wrapped by paths.ResourcePath().
const regressionDBFile = "regressionDB"
const regressionScripts = "regressionScripts"

// Sentinal errors to indicate skip and quite events during the RegressRun() function.
const (
	regressionSkipped   = "regression skipped"
	regressionQuitEarly = "regression quit early"
)

// Regressor is the generic entry type in the regressionDB.
type Regressor interface {
	database.Entry

	// perform the regression test for the regression type. the newRegression
	// flag is for convenience really (or "logical binding", as the structured
	// programmers would have it)
	//
	// message is the string that is to be printed during the regression
	//
	// returns: success boolean; any failure message (not always appropriate;
	// and error state
	regress(newRegression bool, output io.Writer, message string, continueCheck func() bool) (bool, string, error)
}

// when starting a database session we need to register what entries we will
// find in the database.
func initDBSession(db *database.Session) error {
	if err := db.RegisterEntryType(videoEntryID, deserialiseVideoEntry); err != nil {
		return err
	}

	if err := db.RegisterEntryType(playbackEntryID, deserialisePlaybackEntry); err != nil {
		return err
	}

	if err := db.RegisterEntryType(logEntryID, deserialiseLogEntry); err != nil {
		return err
	}

	return nil
}

// RegressList displays all entries in the database.
func RegressList(output io.Writer) error {
	if output == nil {
		return curated.Errorf("regression: list: io.Writer should not be nil (use a nopWriter)")
	}

	dbPth, err := paths.ResourcePath("", regressionDBFile)
	if err != nil {
		return curated.Errorf("regression: list: %v", err)
	}

	db, err := database.StartSession(dbPth, database.ActivityReading, initDBSession)
	if err != nil {
		return err
	}
	defer db.EndSession(false)

	return db.ForEach(func(key int, e database.Entry) error {
		output.Write([]byte(fmt.Sprintf("%03d %s\n", key, e.String())))
		return nil
	})
}

// RegressAdd adds a new regression handler to the database.
func RegressAdd(output io.Writer, reg Regressor) error {
	if output == nil {
		return curated.Errorf("regression: add: io.Writer should not be nil (use a nopWriter)")
	}

	dbPth, err := paths.ResourcePath("", regressionDBFile)
	if err != nil {
		return curated.Errorf("regression: add: %v", err)
	}

	db, err := database.StartSession(dbPth, database.ActivityCreating, initDBSession)
	if err != nil {
		return err
	}
	defer db.EndSession(true)

	msg := fmt.Sprintf("adding: %s", reg)
	_, _, err = reg.regress(true, output, msg, func() bool { return false })
	if err != nil {
		return err
	}

	output.Write([]byte(ansiClearLine))
	output.Write([]byte(fmt.Sprintf("\radded: %s\n", reg)))

	return db.Add(reg)
}

// RegressRedux removes and adds an entry using the same parameters.
func RegressRedux(output io.Writer, confirmation io.Reader) error {
	if output == nil {
		return curated.Errorf("regression: redux: io.Writer should not be nil (use a nopWriter)")
	}

	if confirmation == nil {
		return curated.Errorf("regression: redux: io.Reader should not be nil")
	}

	output.Write([]byte("redux is a dangerous operation. it will rerun all compatible regression entries.\n"))
	output.Write([]byte("redux? (y/n): "))
	if !confirm(confirmation) {
		return nil
	}

	output.Write([]byte("sure? (y/n): "))
	if !confirm(confirmation) {
		return nil
	}

	dbPth, err := paths.ResourcePath("", regressionDBFile)
	if err != nil {
		return curated.Errorf("regression: redux: %v", err)
	}

	db, err := database.StartSession(dbPth, database.ActivityCreating, initDBSession)
	if err != nil {
		return err
	}
	defer db.EndSession(true)

	return db.ForEach(func(key int, e database.Entry) error {
		switch reg := e.(type) {
		case *VideoRegression:
			err = redux(db, output, key, reg)
			if err != nil {
				return curated.Errorf("regression: redux: %v", err)
			}

		case *LogRegression:
			err = redux(db, output, key, reg)
			if err != nil {
				return curated.Errorf("regression: redux: %v", err)
			}

		default:
			output.Write([]byte(fmt.Sprintf("skipped: %s\n", reg)))
		}

		return nil
	})
}

func redux(db *database.Session, output io.Writer, key int, reg Regressor) error {
	err := db.Delete(key)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("reduxing: %s", reg)

	_, _, err = reg.regress(true, output, msg, func() bool { return false })
	if err != nil {
		return err
	}

	output.Write([]byte(ansiClearLine))
	output.Write([]byte(fmt.Sprintf("\rreduxed: %s\n", reg)))

	err = db.Add(reg)
	if err != nil {
		return err
	}
	return nil
}

// RegressDelete removes a cartridge from the regression db.
func RegressDelete(output io.Writer, confirmation io.Reader, key string) error {
	if output == nil {
		return curated.Errorf("regression: delete: io.Writer should not be nil (use a nopWriter)")
	}

	if confirmation == nil {
		return curated.Errorf("regression: delete: io.Reader should not be nil")
	}

	v, err := strconv.Atoi(key)
	if err != nil {
		return curated.Errorf("regression: delete: invalid key [%s]", key)
	}

	dbPth, err := paths.ResourcePath("", regressionDBFile)
	if err != nil {
		return curated.Errorf("regression: delete: %v", err)
	}

	db, err := database.StartSession(dbPth, database.ActivityModifying, initDBSession)
	if err != nil {
		return err
	}
	defer db.EndSession(true)

	ent, err := db.SelectKeys(nil, v)
	if err != nil {
		return curated.Errorf("regression: delete: %v", err)
	}

	output.Write([]byte(fmt.Sprintf("%s\ndelete? (y/n): ", ent)))
	if confirm(confirmation) {
		err = db.Delete(v)
		if err != nil {
			return err
		}
		output.Write([]byte(fmt.Sprintf("deleted test #%s from regression database\n", key)))
	}

	return nil
}

// RegressRun runs all the tests in the regression database. filterKeys
// list specified which entries to test. an empty keys list means that every
// entry should be tested.
func RegressRun(output io.Writer, verbose bool, filterKeys []string) error {
	if output == nil {
		return curated.Errorf("regression: run: io.Writer should not be nil (use a nopWriter)")
	}

	dbPth, err := paths.ResourcePath("", regressionDBFile)
	if err != nil {
		return curated.Errorf("regression: run: %v", err)
	}

	db, err := database.StartSession(dbPth, database.ActivityReading, initDBSession)
	if err != nil {
		return curated.Errorf("regression: run: %v", err)
	}
	defer db.EndSession(false)

	// make sure any supplied keys list is in order
	keysV := make([]int, 0, len(filterKeys))
	for k := range filterKeys {
		v, err := strconv.Atoi(filterKeys[k])
		if err != nil {
			return curated.Errorf("regression: run: invalid key [%s]", filterKeys[k])
		}
		keysV = append(keysV, v)
	}
	sort.Ints(keysV)

	numSucceed := 0
	numFail := 0
	numError := 0
	numSkipped := 0

	defer func() {
		output.Write([]byte(fmt.Sprintf("regression tests: %d succeed, %d fail, %d skipped", numSucceed, numFail, numSkipped)))

		if numError > 0 {
			output.Write([]byte(fmt.Sprintf(" [with %d errors]", numError)))
		}
		output.Write([]byte("\n"))
	}()

	// quitEarly will be set if an interrupt signal is received twice within a
	// quarter of a second. this will cause the onSelect function to return the
	// regressionQuitEarly error
	quitEarly := false

	// check for interrupt signal. a single interrupt signal skips the current
	// regression entry. a second interrupt signal within a quarter of a second
	// quits the entire regression test
	intChan := make(chan os.Signal, 1)
	signal.Notify(intChan, os.Interrupt)
	skipCheck := func() bool {
		select {
		case <-intChan:
			select {
			case <-intChan:
				quitEarly = true
			case <-time.After(250 * time.Millisecond):
				return true
			}
		default:
		}
		return false
	}

	// selectKeys() calls this onSelect function for every key entry
	onSelect := func(ent database.Entry) error {
		// if the quitEarly flag has been set then return the
		// regressionQuitEarly error
		if quitEarly {
			return curated.Errorf(regressionQuitEarly)
		}

		// database entry should also satisfy Regressor interface
		reg, ok := ent.(Regressor)
		if !ok {
			return curated.Errorf("regression: run: database entry does not satisfy Regressor interface")
		}

		// run regress() function with message. message does not have a
		// trailing newline
		msg := fmt.Sprintf("running: %s", reg)
		ok, failm, err := reg.regress(false, output, msg, skipCheck)

		// once regress() has completed we clear the line ready for the
		// completion message
		output.Write([]byte(ansiClearLine))

		// print completion message depending on result of regress()
		if err != nil {
			if curated.Has(err, regressionSkipped) {
				numSkipped++
				output.Write([]byte(fmt.Sprintf("\rskipped: %s\n", reg)))
			} else {
				numError++
				output.Write([]byte(fmt.Sprintf("\rerror: %s\n", reg)))

				// output any error message on following line
				if verbose {
					output.Write([]byte(fmt.Sprintf("  ^^ %s\n", err)))
				}
			}
		} else if !ok {
			numFail++
			output.Write([]byte(fmt.Sprintf("\rfailure: %s\n", reg)))
			if verbose && failm != "" {
				output.Write([]byte(fmt.Sprintf("  ^^ %s\n", failm)))
			}
		} else {
			numSucceed++
			output.Write([]byte(fmt.Sprintf("\rsucceed: %s\n", reg)))
		}

		return nil
	}

	_, err = db.SelectKeys(onSelect, keysV...)

	// filter out regressionQuitEarly errors
	if err != nil && !curated.Is(err, regressionQuitEarly) {
		return curated.Errorf("regression: run: %v", err)
	}

	return nil
}

// returns true if response from user begins with 'y' or 'Y'.
func confirm(confirmation io.Reader) bool {
	confirm := make([]byte, 32)
	_, err := confirmation.Read(confirm)
	if err != nil {
		return false
	}

	if confirm[0] == 'y' || confirm[0] == 'Y' {
		return true
	}
	return false
}
