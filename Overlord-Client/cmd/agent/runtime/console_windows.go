//go:build windows

package runtime

import (
	"context"
	"strings"

	"github.com/UserExistsError/conpty"
)

type winPty struct {
	cp     *conpty.ConPty
	waitCh chan waitResult
}

type waitResult struct {
	code int
	err  error
}

func (w *winPty) Read(p []byte) (int, error)  { return w.cp.Read(p) }
func (w *winPty) Write(p []byte) (int, error) { return w.cp.Write(p) }
func (w *winPty) Close() error                { return w.cp.Close() }

func (w *winPty) Resize(cols, rows uint16) error {
	return w.cp.Resize(int(cols), int(rows))
}

func (w *winPty) Wait() (int, error) {
	res := <-w.waitCh
	return res.code, res.err
}

func startPty(ctx context.Context, shell []string, cols, rows uint16) (ptyHandle, error) {
	cmdline := strings.Join(shell, " ")

	cp, err := conpty.Start(cmdline, conpty.ConPtyDimensions(int(cols), int(rows)))
	if err != nil {
		return nil, err
	}

	w := &winPty{cp: cp, waitCh: make(chan waitResult, 1)}
	go func() {
		code, werr := cp.Wait(context.Background())
		w.waitCh <- waitResult{code: int(code), err: werr}
	}()
	return w, nil
}
