package main_test

import (
	"flag"
	"log"
	"os/exec"
	"testing"
)

var expected = flag.String("expect", "0", "Expected exit code for compiled executable")

func TestCompile(t *testing.T) {

	defer ExecuteProgram(t)
	CompileProgram(t)
}

func CompileProgram(t *testing.T) {
	runCompile := exec.Command("go", "run", ".", "../../test/testfile.pn")
	if err := runCompile.Run(); err != nil {
		t.Errorf("Compilation did not complete, error: %v", err)
	}
}

func ExecuteProgram(t *testing.T) {
	runExec := exec.Command("./testfile")
	runExec.Dir = "../../test"
	out, err := runExec.Output()
	if err != nil {
		t.Errorf("Program exited with %v, expected %v", err, *expected)
	} else {
		log.Print("PROGRAM OUTPUT: " + "\n\n" + string(out) + "\n\n")
	}

}
