package modules

import (
	"fmt"
	"github.com/davidscholberg/go-i3barjson"
	"os/exec"
	"strings"
)

// Volume represents the configuration for the volume display block.
type Volume struct {
	BlockConfigBase `yaml:",inline"`
}

// UpdateBlock updates the volume display block.
// Currently, only the ALSA master channel volume is supported.
func (c Volume) UpdateBlock(b *i3barjson.Block) {
	b.Color = c.Color
	fullTextFmt := fmt.Sprintf("%s%%s", c.Label)
	amixerCmd := "amixer"
	amixerArgs := []string{"-D", "default", "get", "Master"}
	out, err := exec.Command(amixerCmd, amixerArgs...).Output()
	if err != nil {
		b.FullText = fmt.Sprintf(fullTextFmt, err.Error())
		return
	}
	outStr := string(out)
	iBegin := strings.Index(outStr, "[")
	if iBegin == -1 {
		b.FullText = fmt.Sprintf(fullTextFmt, "cannot parse amixer output")
		return
	}
	iEnd := strings.Index(outStr, "]")
	if iEnd == -1 {
		b.FullText = fmt.Sprintf(fullTextFmt, "cannot parse amixer output")
		return
	}
	b.FullText = fmt.Sprintf(fullTextFmt, outStr[iBegin+1:iEnd])
}
