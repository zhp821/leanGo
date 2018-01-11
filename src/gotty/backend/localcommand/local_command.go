package localcommand

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/kr/pty"
	"github.com/pkg/errors"
)

const (
	DefaultCloseSignal  = syscall.SIGINT
	DefaultCloseTimeout = 10 * time.Second
)

type LocalCommand struct {
	command      string
	argv         []string
	info         string
	closeSignal  syscall.Signal
	closeTimeout time.Duration
	cmd          *exec.Cmd
	pty          *os.File
	delcmd       []byte
	ptyClosed    chan struct{}
}

func New(command string, argv []string, options ...Option) (*LocalCommand, error) {

	cmd := exec.Command(command, argv...)

	pty, err := pty.Start(cmd)
	if err != nil {
		// todo close cmd?
		return nil, errors.Wrapf(err, "failed to start command `%s`", command)
	}
	array := []byte{127, 127, 127, 127, 127, 127, 127, 127, 127, 127, 127, 127, 127, 127, 127, 127, 127, 127, 127, 127}

	ptyClosed := make(chan struct{})
	lcmd := &LocalCommand{
		command:      command,
		argv:         argv,
		info:         "",
		closeSignal:  DefaultCloseSignal,
		closeTimeout: DefaultCloseTimeout,
		cmd:          cmd,
		pty:          pty,
		delcmd:       array,
		ptyClosed:    ptyClosed,
	}

	for _, option := range options {
		option(lcmd)
	}

	// When the process is closed by the user,
	// close pty so that Read() on the pty breaks with an EOF.
	go func() {
		defer func() {
			lcmd.pty.Close()
			close(lcmd.ptyClosed)
		}()

		lcmd.cmd.Wait()
	}()

	return lcmd, nil
}

func (lcmd *LocalCommand) Read(p []byte) (n int, err error) {
	n, e := lcmd.pty.Read(p)
	return n, e
}

func (lcmd *LocalCommand) Write(p []byte) (n int, err error) {
	if p[0] == 13 {
		fmt.Printf("cmd: %s \n", lcmd.info)
		text := strings.TrimSpace(lcmd.info)
		lcmd.info = ""
		if strings.Contains(text, "\\") || strings.Contains(text, "&") || strings.Contains(text, "|") || strings.Contains(text, ";") {
			return lcmd.pty.Write(lcmd.delcmd)
		}
		if strings.HasPrefix(text, "pwd") || strings.HasPrefix(text, "exit") || strings.HasPrefix(text, "cd") || strings.HasPrefix(text, "ls") || strings.HasPrefix(text, "ps") || strings.HasPrefix(text, "less") || strings.HasPrefix(text, "tail") || strings.HasPrefix(text, "more") || strings.HasPrefix(text, "cat") {
			return lcmd.pty.Write(p)
		} else {
			return lcmd.pty.Write(lcmd.delcmd)
		}
	}
	if p[0] == 127 {
		rs := []rune(lcmd.info)
		length := len(rs)
		if length > 0 {
			lcmd.info = string(rs[0 : length-1])
			return lcmd.pty.Write(p)
		}
	}
	lcmd.info = lcmd.info + string(p[:])
	return lcmd.pty.Write(p)
}

func (lcmd *LocalCommand) Close() error {
	if lcmd.cmd != nil && lcmd.cmd.Process != nil {
		lcmd.cmd.Process.Signal(lcmd.closeSignal)
	}
	for {
		select {
		case <-lcmd.ptyClosed:
			return nil
		case <-lcmd.closeTimeoutC():
			lcmd.cmd.Process.Signal(syscall.SIGKILL)
		}
	}
}

func (lcmd *LocalCommand) WindowTitleVariables() map[string]interface{} {
	return map[string]interface{}{
		"command": lcmd.command,
		"argv":    lcmd.argv,
		"pid":     lcmd.cmd.Process.Pid,
	}
}

func (lcmd *LocalCommand) ResizeTerminal(width int, height int) error {
	window := struct {
		row uint16
		col uint16
		x   uint16
		y   uint16
	}{
		uint16(height),
		uint16(width),
		0,
		0,
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		lcmd.pty.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&window)),
	)
	if errno != 0 {
		return errno
	} else {
		return nil
	}
}

func (lcmd *LocalCommand) GetTerminalSize() (int, int, error) {
	return pty.Getsize(lcmd.pty)
}

func (lcmd *LocalCommand) closeTimeoutC() <-chan time.Time {
	if lcmd.closeTimeout >= 0 {
		return time.After(lcmd.closeTimeout)
	}

	return make(chan time.Time)
}
