// +build !cgo
// +build upgrade,ignore
// +build !windows

package main

import (
	"fmt"
	"os/exec"
)

func buildAmalgamation(baseDir, buildFlags string) error {
	args := []string{"configure"}
	if buildFlags != "" {
		args = append(args, "CFLAGS="+buildFlags)
	}
	cmd := exec.Command("sh", args...)
	cmd.Dir = baseDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("configure failed: %v\n\n%s", err, out)
	}
	fmt.Println("Ran configure successfully")

	cmd = exec.Command("make", "sqlite3.c")
	cmd.Dir = baseDir
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("make failed: %v\n\n%s", err, out)
	}
	fmt.Println("Ran make successfully")

	return nil
}
