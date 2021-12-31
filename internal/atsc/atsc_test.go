package atsc

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	validChannelsConf = `KCTS-HD:189000000:8VSB:49:52:3
KIDS:189000000:8VSB:65:68:4
CREATE:189000000:8VSB:81:84:5
WORLD:189000000:8VSB:97:100:6`

	validChannelsConfNonstandard8VSB = "KCTS-HD:189000000:VSB_8:49:52:3"
	validChannelsConfQAM64           = "Test QAM 64:255000000:QAM_64:42:43:5"
	validChannelsConfQAM256          = "WLFI:255000000:QAM_256:66:68:4"
)

func TestParseChannelsConf(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []Channel
		wantErr bool
	}{
		{
			name:  "valid channels.conf",
			input: validChannelsConf,
			want: []Channel{
				{"KCTS-HD", 189_000_000, Modulation8VSB, 49, 52, 3},
				{"KIDS", 189_000_000, Modulation8VSB, 65, 68, 4},
				{"CREATE", 189_000_000, Modulation8VSB, 81, 84, 5},
				{"WORLD", 189_000_000, Modulation8VSB, 97, 100, 6},
			},
		},

		{
			name:  "w_scan2 nonstandard 8VSB output",
			input: validChannelsConfNonstandard8VSB,
			want: []Channel{
				{"KCTS-HD", 189_000_000, Modulation8VSB, 49, 52, 3},
			},
		},

		{
			name:  "QAM64 modulation",
			input: validChannelsConfQAM64,
			want: []Channel{
				{"Test QAM 64", 255_000_000, ModulationQAM64, 42, 43, 5},
			},
		},

		{
			name:  "QAM256 modulation",
			input: validChannelsConfQAM256,
			want: []Channel{
				{"WLFI", 255_000_000, ModulationQAM256, 66, 68, 4},
			},
		},

		{
			name:    "wrong number of fields",
			input:   "KCTS-HD:189000000:8VSB:3",
			wantErr: true,
		},

		{
			name:    "invalid frequency",
			input:   "KCTS-HD:189.0123456:8VSB:49:52:3",
			wantErr: true,
		},

		{
			name:    "invalid modulation",
			input:   "KCTS-HD:189000000:42VSB:49:52:3",
			wantErr: true,
		},

		{
			name:    "invalid PID",
			input:   "KCTS-HD:189000000:8VSB:49:52:?",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseChannelsConf(strings.NewReader(tc.input))
			if err != nil {
				if !tc.wantErr {
					t.Fatalf("unexpected error: %v", err)
				}
				t.Logf("error: %v", err)
				return
			}

			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func FuzzParseChannelsConf(f *testing.F) {
	f.Add(validChannelsConf)
	f.Add(validChannelsConfNonstandard8VSB)
	f.Add(validChannelsConfQAM64)
	f.Add(validChannelsConfQAM256)

	f.Fuzz(func(t *testing.T, inputStringConf string) {
		parsedChannels, err := ParseChannelsConf(strings.NewReader(inputStringConf))
		if err != nil {
			t.SkipNow()
		}

		encodedChannels := formatChannelsConf(t, parsedChannels)
		parsedChannels2, err := ParseChannelsConf(strings.NewReader(encodedChannels))
		if err != nil {
			t.Fatalf("error re-parsing encoded channel list: %v", err)
		}

		diff := cmp.Diff(parsedChannels, parsedChannels2)
		if diff != "" {
			t.Errorf("could not round-trip channels.conf (-want +got):\n%s", diff)
		}
	})
}

func formatChannelsConf(t *testing.T, channels []Channel) string {
	t.Helper()
	var buf bytes.Buffer
	for _, ch := range channels {
		buf.WriteString(ch.String())
		buf.WriteRune('\n')
	}
	return buf.String()
}
