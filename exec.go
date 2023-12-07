package embedbin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Exec is a wrapper around embed binary file.
type Exec struct {
	filename string
	data     []byte

	path string
}

// NewExec creates a new Exec using the filename and data.
func NewExec(filename string, data []byte) (e *Exec) {
	return &Exec{filename: filename, data: data}
}

// Command creates an exec.Cmd using the args.
func (e *Exec) Command(ctx context.Context, args ...string) (*exec.Cmd, error) {
	if e.path == "" {
		path, err := createFile(e.filename, e.data, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s: %w", e.filename, err)
		}
		e.path = path
	}
	return exec.CommandContext(ctx, e.path, args...), nil
}

func tempDir() string {
	return filepath.Join(os.TempDir(), "embedbin")
}

func createFile(filename string, data []byte, mode os.FileMode) (string, error) {
	dir := tempDir()
	err := os.MkdirAll(dir, mode)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(data)
	name := filename + "-" + hex.EncodeToString(sum[:])

	defaultPath := filepath.Join(dir, name+suffix)
	err0 := saveFile(defaultPath, data, mode)
	if err0 == nil {
		return defaultPath, nil
	}

	// fallback to temp file
	tempPath, err1 := saveTempFile(dir, name, data, mode)
	if err1 == nil {
		return tempPath, nil
	}

	return "", fmt.Errorf("failed to create file %s: %w", filename, errors.Join(err0, err1))
}

func saveFile(filename string, data []byte, mode os.FileMode) error {
	stat, err := os.Stat(filename)
	if !os.IsExist(err) {
		return writeFile(filename, data, mode)
	}

	if int(stat.Size()) != len(data) {
		return writeFile(filename, data, mode)
	}

	raw, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if !bytes.Equal(raw, data) {
		return writeFile(filename, data, mode)
	}

	if stat.Mode() != mode {
		err = os.Chmod(filename, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFile(filename string, data []byte, mode os.FileMode) error {
	tmpFile := filename + ".tmp"
	err := os.WriteFile(tmpFile, data, mode)
	if err != nil {
		return err
	}

	err = os.Chmod(tmpFile, mode)
	if err != nil {
		return err
	}

	err = os.Rename(tmpFile, filename)
	if err != nil {
		return err
	}
	return nil
}

func saveTempFile(dir, name string, data []byte, mode os.FileMode) (string, error) {
	file, err := os.CreateTemp(dir, name+"-*"+suffix)
	if err != nil {
		return "", err
	}

	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return "", err
	}

	err = file.Sync()
	if err != nil {
		return "", err
	}

	err = file.Chmod(mode)
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}
