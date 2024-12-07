package terminalcmd

import (
	"io"
	"os"
	"os/exec"

	"github.com/scottjr632/dotctl/internal/promise"
)

type Cmd struct {
	cmd  string
	args []string
	env  []string
}

func New(cmd string, args ...string) *Cmd {
	return &Cmd{cmd: cmd, args: args}
}

func (c *Cmd) WithEnv(env ...string) *Cmd {
	c.env = env
	return c
}

func (c *Cmd) SilentlyExecute() (output string, err error) {
	out, err := exec.Command(c.cmd, c.args...).CombinedOutput()
	return string(out), err
}

func (c *Cmd) SilentlyExecuteAsync() (output string, err error) {
	p := promise.New(func() (string, error) { return c.SilentlyExecute() })
	output, err = p.Await()
	return output, err
}

func (c *Cmd) ExecuteToStdout() error {
	cmd := exec.Command(c.cmd, c.args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err = cmd.Start(); err != nil {
		return err
	}

	if _, err = io.Copy(os.Stdout, stdout); err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (c *Cmd) ExecuteInTerminal() error {
	cmd := exec.Command(c.cmd, c.args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
