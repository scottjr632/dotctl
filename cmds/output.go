package cmds

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type response struct {
	OK   bool `json:"ok"`
	Data any  `json:"data"`
}

func addJSONFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "output machine-readable JSON")
}

func wantsJSON(cmd *cobra.Command) bool {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	return jsonOutput
}

func writeJSON(cmd *cobra.Command, data any) error {
	return writeJSONResponse(cmd, true, data)
}

func writeJSONResponse(cmd *cobra.Command, ok bool, data any) error {
	return json.NewEncoder(cmd.OutOrStdout()).Encode(response{OK: ok, Data: data})
}

func writePlan(cmd *cobra.Command, actions ...string) error {
	fmt.Fprintln(cmd.OutOrStdout(), "Dry run; no changes made:")
	for _, action := range actions {
		fmt.Fprintln(cmd.OutOrStdout(), "- "+action)
	}
	return nil
}
