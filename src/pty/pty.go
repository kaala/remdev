package pty

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
)

// Terminal wraps a PTY-backed shell process.
type Terminal struct {
	ID    string
	cmd   *exec.Cmd
	ptmx  *os.File
	title string
	mu    sync.Mutex
	done  chan struct{}
}

// New creates a new PTY terminal with the given UUID.
func New(id string) (*Terminal, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}

	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: 80, Rows: 24})
	if err != nil {
		return nil, fmt.Errorf("start pty: %w", err)
	}

	t := &Terminal{
		ID:    id,
		cmd:   cmd,
		ptmx:  ptmx,
		title: shell,
		done:  make(chan struct{}),
	}

	return t, nil
}

func (t *Terminal) Read(p []byte) (int, error)  { return t.ptmx.Read(p) }
func (t *Terminal) Write(p []byte) (int, error) { return t.ptmx.Write(p) }

func (t *Terminal) Resize(cols, rows int) error {
	return pty.Setsize(t.ptmx, &pty.Winsize{
		Cols: uint16(cols),
		Rows: uint16(rows),
	})
}

func (t *Terminal) Title() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.title
}

func (t *Terminal) SetTitle(title string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.title = title
}

func (t *Terminal) Wait() int {
	err := t.cmd.Wait()
	t.ptmx.Close()
	close(t.done)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}

func (t *Terminal) Done() <-chan struct{} { return t.done }

func (t *Terminal) Kill() error { return t.cmd.Process.Kill() }

func (t *Terminal) Close() {
	t.ptmx.Close()
	if t.cmd.Process != nil {
		t.cmd.Process.Kill()
	}
}
