package exec

import (
	"os/exec"
	"bytes"
	"fmt"
)

func Wait(format string, a ...interface{}) (string, string, error) {
    var stdout, stderr bytes.Buffer
	var cmd *exec.Cmd
	cmd = exec.Command("/bin/bash", "-c", fmt.Sprintf(format, a...))
	cmd.Stdout = &stdout
    cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
func NoWait(format string, a ...interface{}) error {
	var stdout, stderr bytes.Buffer
	var cmd *exec.Cmd
	cmd = exec.Command("/bin/bash", "-c", fmt.Sprintf(format, a...))
	cmd.Stdout = &stdout
    cmd.Stderr = &stderr
	return cmd.Start()
}
