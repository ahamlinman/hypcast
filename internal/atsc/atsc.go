// Package atsc provides representations of ATSC television channel information.
package atsc

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Modulation represents the modulation of an ATSC television channel.
type Modulation string

// The following are the normalized Modulation values for a Channel.
const (
	// Modulation8VSB may also be parsed as "VSB_8" in a channels.conf file.
	Modulation8VSB   Modulation = "8VSB"
	ModulationQAM64  Modulation = "QAM_64"
	ModulationQAM256 Modulation = "QAM_256"
)

// Channel represents the definition of an ATSC television channel.
type Channel struct {
	Name        string
	FrequencyHz uint
	Modulation  Modulation
	VideoPID    uint
	AudioPID    uint
	ProgramID   uint
}

// String returns the representation of c in the azap-compatible format
// described by ParseChannelsConf.
func (c Channel) String() string {
	return fmt.Sprintf(
		"%s:%d:%s:%d:%d:%d",
		c.Name, c.FrequencyHz, c.Modulation, c.VideoPID, c.AudioPID, c.ProgramID,
	)
}

// ParseChannelsConf parses Channels from an azap-compatible channels.conf file
// read from r.
//
// Each line of the file defines a single channel, and is formatted as 6
// colon-separated fields corresponding to the fields of Channel as follows:
//
//   Name:FrequencyHz:Modulation:VideoPID:AudioPID:ProgramID
//
// FrequencyHz, VideoPID, AudioPID, and ProgramID are all represented in decimal
// form.
//
// The https://github.com/stefantalpalaru/w_scan2 utility is useful for
// generating a compatible file. For example, to scan for terrestrial broadcast
// channels in the United States of America:
//
//   w_scan2 -f a -c us -X > channels.conf
func ParseChannelsConf(r io.Reader) ([]Channel, error) {
	var (
		channels []Channel
		err      error
		line     = 0
		scanner  = bufio.NewScanner(r)
	)

	for scanner.Scan() {
		line++
		fields := strings.Split(scanner.Text(), ":")

		const expectedFields = 6
		if len(fields) != expectedFields {
			return nil, fmt.Errorf(
				"channels.conf line %d has %d fields, expected %d",
				line, len(fields), expectedFields,
			)
		}

		parseUint := func(s string) uint {
			i, ierr := strconv.ParseUint(s, 10, 0)
			if ierr != nil && err == nil {
				err = fmt.Errorf("channels.conf line %d has invalid field value %q", line, s)
			}
			return uint(i)
		}

		parseModulation := func(s string) Modulation {
			switch s {
			case "8VSB", "VSB_8":
				return Modulation8VSB

			case "QAM_64":
				return ModulationQAM64

			case "QAM_256":
				return ModulationQAM256
			}

			if err == nil {
				err = fmt.Errorf("channels.conf line %d has unknown modulation %q", line, s)
			}
			// This value has no semantic meaning in the error case, we just have to
			// return *something* of the right type.
			return Modulation8VSB
		}

		channels = append(channels, Channel{
			Name:        fields[0],
			FrequencyHz: parseUint(fields[1]),
			Modulation:  parseModulation(fields[2]),
			VideoPID:    parseUint(fields[3]),
			AudioPID:    parseUint(fields[4]),
			ProgramID:   parseUint(fields[5]),
		})
		if err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("unable to read channels.conf: %w", err)
	}

	return channels, nil
}
