// +build !cgo
// +build upgrade,ignore

package main

import (
	"fmt"
	"os/exec"
)

func buildAmalgamation(baseDir, buildFlags string) error {
	args := []string{"/f", "Makefile.msc", "sqlite3.c"}
	if buildFlags != "" {
		args = append(args, "OPTS="+buildFlags)
	}
	cmd := exec.Command("nmake", args...)
	cmd.Dir = baseDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nmake failed: %v\n\n%s", err, out)
	}
	fmt.Println("Ran nmake successfully")

	return nil
}
