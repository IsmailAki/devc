package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/IsmailAki/devc/pkg/types"
	"github.com/spf13/cobra"
)

var (
	listAll  bool
	listJSON bool
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List development containers",
	Aliases: []string{"ls", "ps"},
	Run:     runList,
}

func init() {
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "Show all containers including stopped")
	listCmd.Flags().BoolVarP(&listJSON, "json", "j", false, "Output in JSON format")
	rootCmd.AddCommand(listCmd)
}

type containerInfo struct {
	Name     string
	State    *types.ContainerState
	Metadata *types.ContainerMetadata
}

func runList(cmd *cobra.Command, args []string) {
	filtered, err := loadContainerInfos(listAll)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing containers: %v\n", err)
		os.Exit(1)
	}

	if len(filtered) == 0 && listAll {
		fmt.Println("No dev containers found")
		return
	}

	if len(filtered) == 0 {
		fmt.Println("No running dev containers (use --all to see all)")
		return
	}

	if listJSON {
		printJSON(filtered)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tPORT\tFEATURES\tCREATED")
	for _, c := range filtered {
		features := ""
		if len(c.Metadata.Features) > 0 {
			features = c.Metadata.Features[0]
			if len(c.Metadata.Features) > 1 {
				features += fmt.Sprintf(" (+%d)", len(c.Metadata.Features)-1)
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
			c.Name,
			c.State.Status,
			c.State.SSHPort,
			features,
			c.State.CreatedAt.Format("2006-01-02 15:04"),
		)
	}
	w.Flush()
}

func printJSON(containers []containerInfo) {
	payload := make([]map[string]interface{}, 0, len(containers))
	for _, c := range containers {
		payload = append(payload, map[string]interface{}{
			"name":       c.Name,
			"status":     c.State.Status,
			"port":       c.State.SSHPort,
			"features":   c.Metadata.Features,
			"created_at": c.State.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}
