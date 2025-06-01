package output

import (
	"github.com/ffuf/ffuf/v2/pkg/api/output"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

func NewOutputProviderByName(name string, conf *ffuf.Config) ffuf.OutputProvider {
	switch name {
	case "api":
		return output.NewAPIOutput(conf)
	default:
		return NewStdoutput(conf)
	}
}
