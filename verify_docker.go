package main
import (
	"fmt"
	"sort"
"strings"
)

func verifyRunningDockerContainers(output string, expectedContainers []string) []string {
	var errors []string

	lines := strings.Split(output, "\n")

	// remove first char of container name from docker inspect output which is '/'
	for i, _ := range lines {
		if len(lines[i]) > 0 {
			lines[i] = lines[i][1:]
		}
	}

	sort.Strings(lines)

	for _, name := range expectedContainers {
		i := sort.Search(len(lines),
			func(i int) bool { return lines[i] >= name })
		if (i >= len(lines) || (i < len(lines) && lines[i] != name)) {
			errors = append(errors, fmt.Sprintf("Docker container '%s' is not running\n", name))
		}
	}

	return errors
}
