package embedbin

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestHello(t *testing.T) {
	var code = "./testdata/helloworld.c"
	var bin = "./testdata/helloworld" + suffix
	err := exec.Command("gcc", code, "-o", bin).Run()
	if err != nil {
		t.Fatal(err)
	}

	binData, err := os.ReadFile(bin)
	if err != nil {
		t.Fatal(err)
	}

	e := NewExec("hello", binData)
	cmd, err := e.Command(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	out := bytes.NewBuffer(nil)
	cmd.Stdout = out
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(out.String()) != "Hello World!" {
		t.Fatalf("out: %q", out.String())
	}
}
