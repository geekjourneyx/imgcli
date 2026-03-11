package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
)

type SuccessResponse struct {
	OK      bool   `json:"ok"`
	Command string `json:"command"`
	Data    any    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Error    string `json:"error"`
	Code     string `json:"code"`
	ExitCode int    `json:"exit_code"`
}

func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func PrintSuccess(w io.Writer, command string, data any) error {
	return WriteJSON(w, SuccessResponse{
		OK:      true,
		Command: command,
		Data:    data,
	})
}

func PrintError(w io.Writer, err error) {
	appErr := apperr.From(err)
	_ = WriteJSON(w, ErrorResponse{
		Error:    appErr.Message,
		Code:     appErr.Code,
		ExitCode: appErr.ExitCode,
	})
}

func ValidateJSONEnabled(jsonOutput bool) error {
	if !jsonOutput {
		return fmt.Errorf("plain-text output is not supported in this build")
	}
	return nil
}
