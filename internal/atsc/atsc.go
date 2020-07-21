// Package atsc provides representations of ATSC television channel information
// for use by a TV tuner card.
//
// For the purposes of this package, channels.conf files are those using the
// azap-compatible format described at
// https://www.mythtv.org/wiki/Adding_Digital_Cable_Channels_(For_ATSC/QAM_Tuner_Cards_--_USA/Canada)#channels.conf_Format.
//
// The https://github.com/stefantalpalaru/w_scan2 utility is useful for
// generating a compatible file:
//
//   ./w_scan2 -f a -c us -X >> channels.conf
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
	// Modulation8VSB may optionally be parsed as "VSB_8" in a channels.conf file,
	// and will be normalized to this value.
	Modulation8VSB   Modulation = "8VSB"
	ModulationQAM64  Modulation = "QAM_64"
	ModulationQAM256 Modulation = "QAM_256"
)

// numChannelFields is used to ensure that each channels.conf line is valid. It
// must match the number of fields in Channel.
const numChannelFields = 6

// Channel represents the definition of an ATSC television channel.
type Channel struct {
	Name       string
	Frequency  uint
	Modulation Modulation
	VideoPID   uint
	AudioPID   uint
	ProgramID  uint
}

func (c Channel) String() string {
	return fmt.Sprintf(
		"%s:%d:%s:%d:%d:%d",
		c.Name, c.Frequency, c.Modulation, c.VideoPID, c.AudioPID, c.ProgramID,
	)
}

// ParseChannelsConf parses Channels from a channels.conf file.
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
		if len(fields) != numChannelFields {
			return nil, fmt.Errorf(
				"channels.conf line %d has %d fields, expected %d",
				line, len(fields), numChannelFields,
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
			return Modulation("") // Invalid; should never leave this function
		}

		channels = append(channels, Channel{
			Name:       fields[0],
			Frequency:  parseUint(fields[1]),
			Modulation: parseModulation(fields[2]),
			VideoPID:   parseUint(fields[3]),
			AudioPID:   parseUint(fields[4]),
			ProgramID:  parseUint(fields[5]),
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
