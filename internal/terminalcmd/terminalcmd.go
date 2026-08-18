package terminalcmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/srctl/dotctl/internal/promise"
)

var (
	captureOutput  bool
	capturedOutput strings.Builder
)

type Cmd struct {
	cmd  string
	args []string
	env  []string
}

func SetCaptureOutput(capture bool) {
	captureOutput = capture
}

func ResetCapturedOutput() {
	capturedOutput.Reset()
}

func CapturedOutput() string {
	return capturedOutput.String()
}

func New(cmd string, args ...string) *Cmd {
	return &Cmd{cmd: cmd, args: args}
}

func (c *Cmd) WithEnv(env ...string) *Cmd {
	c.env = env
	return c
}

func (c *Cmd) command() *exec.Cmd {
	cmd := exec.Command(c.cmd, c.args...)
	if len(c.env) > 0 {
		cmd.Env = append(os.Environ(), c.env...)
	}
	return cmd
}

func (c *Cmd) SilentlyExecute() (output string, err error) {
	out, err := c.command().CombinedOutput()
	output = string(out)
	if err != nil && strings.TrimSpace(output) != "" {
		return output, fmt.Errorf("%w: %s", err, strings.TrimSpace(output))
	}
	return output, err
}

func (c *Cmd) SilentlyExecuteAsync() (output string, err error) {
	p := promise.New(func() (string, error) { return c.SilentlyExecute() })
	output, err = p.Await()
	return output, err
}

func (c *Cmd) ExecuteToStdout() error {
	cmd := c.command()

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
	if captureOutput {
		output, err := c.SilentlyExecute()
		if output != "" {
			if capturedOutput.Len() > 0 && !strings.HasSuffix(capturedOutput.String(), "\n") {
				capturedOutput.WriteByte('\n')
			}
			capturedOutput.WriteString(output)
		}
		return err
	}

	cmd := c.command()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
