package render

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// RenderPNG invokes the Graphviz `dot` command line tool to render a PNG from a DOT file.
// If `dot` is not available on PATH, it returns (false, nil) without failing.
func RenderPNG(dotPath string, pngPath string) (bool, error) {
	_, err := exec.LookPath("dot")
	if err != nil {
		// Graphviz not installed on PATH - graceful skip
		return false, nil
	}

	cmd := exec.Command("dot", "-Tpng", dotPath, "-o", pngPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return false, fmt.Errorf("graphviz dot execution failed: %s", errMsg)
	}

	return true, nil
}
