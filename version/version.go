package version

import "fmt"

var (
	Version   string
	GitCommit string

	HumanVersion = fmt.Sprintf("v%s (%s)", Version, GitCommit)
)
