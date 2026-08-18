package cmds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/srctl/dotctl/internal/config"
	"github.com/srctl/dotctl/internal/terminalcmd"
)

type response struct {
	OK    bool           `json:"ok"`
	Kind  string         `json:"kind"`
	Data  any            `json:"data,omitempty"`
	Error *errorResponse `json:"error,omitempty"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type planAction struct {
	Operation   string `json:"operation"`
	Description string `json:"description"`
}

type planData struct {
	Actions []planAction `json:"actions"`
}

type commandResult struct {
	Command string `json:"command"`
	Output  string `json:"output,omitempty"`
}

type cliError struct {
	code string
	data any
	err  error
}

func (e *cliError) Error() string {
	return e.err.Error()
}

func (e *cliError) Unwrap() error {
	return e.err
}

var responseWritten bool

func wantsJSON(_ *cobra.Command) bool {
	return jsonOutput
}

func action(operation, description string) planAction {
	return planAction{Operation: operation, Description: description}
}

func writeJSON(cmd *cobra.Command, data any) error {
	return writeResponse(cmd.OutOrStdout(), response{OK: true, Kind: "result", Data: data})
}

func writePlan(cmd *cobra.Command, actions ...planAction) error {
	if jsonOutput {
		return writeResponse(cmd.OutOrStdout(), response{OK: true, Kind: "plan", Data: planData{Actions: actions}})
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Dry run; no changes made:")
	for _, action := range actions {
		fmt.Fprintln(cmd.OutOrStdout(), "- "+action.Description)
	}
	return nil
}

func writeCommandResult(writer io.Writer, command, output string) error {
	return writeResponse(writer, response{
		OK:   true,
		Kind: "result",
		Data: commandResult{Command: command, Output: strings.TrimSpace(output)},
	})
}

func writeErrorResponse(writer io.Writer, err error) error {
	payload := response{
		OK:   false,
		Kind: "error",
		Error: &errorResponse{
			Code:    errorCode(err),
			Message: err.Error(),
		},
	}
	var cliErr *cliError
	if errors.As(err, &cliErr) {
		payload.Data = cliErr.data
	} else if output := strings.TrimSpace(terminalcmd.CapturedOutput()); output != "" {
		payload.Data = commandResult{Command: executedCommand, Output: output}
	}
	return writeResponse(writer, payload)
}

func writeResponse(writer io.Writer, payload response) error {
	responseWritten = true
	return json.NewEncoder(writer).Encode(payload)
}

func errorCode(err error) string {
	var cliErr *cliError
	if errors.As(err, &cliErr) {
		return cliErr.code
	}
	if errors.Is(err, config.ErrConfigMissing) {
		return "CONFIG_NOT_FOUND"
	}
	if errors.Is(err, config.ErrConfigInvalid) {
		return "CONFIG_INVALID"
	}
	if errors.Is(err, os.ErrPermission) {
		return "PERMISSION_DENIED"
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return "EXTERNAL_COMMAND_FAILED"
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "required") || strings.Contains(message, "unknown flag") || strings.Contains(message, "accepts ") {
		return "INVALID_ARGUMENT"
	}
	return "COMMAND_FAILED"
}
