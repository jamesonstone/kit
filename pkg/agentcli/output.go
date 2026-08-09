package agentcli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/jamesonstone/kit/internal/registry"
)

func newJSONEncoder(writer io.Writer) *json.Encoder {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder
}

func writePlanHuman(writer io.Writer, plan registry.Plan, applied bool) error {
	verb := "planned"
	if applied {
		verb = "applied"
	}
	if _, err := fmt.Fprintf(writer, "%s: %s (%d change(s), %d artifact(s))\n", verb, plan.State, len(plan.Changes), len(plan.Artifacts)); err != nil {
		return err
	}
	for _, artifact := range plan.Artifacts {
		if _, err := fmt.Fprintf(writer, "  %-10s %-34s %-16s %s\n", artifact.Kind, artifact.Slug, artifact.State, artifact.Action); err != nil {
			return err
		}
	}
	for _, diagnostic := range plan.Diagnostics {
		if _, err := fmt.Fprintln(writer, "  attention:", diagnostic); err != nil {
			return err
		}
	}
	for _, action := range plan.NextActions {
		if _, err := fmt.Fprintln(writer, "  next:", action); err != nil {
			return err
		}
	}
	return nil
}

func sortedRecords(records []registry.ArtifactRecord, kind string) []registry.ArtifactRecord {
	var result []registry.ArtifactRecord
	for _, record := range records {
		if kind == "" || record.Kind == kind {
			result = append(result, record)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return registry.ArtifactKey(result[i].Kind, result[i].Slug) < registry.ArtifactKey(result[j].Kind, result[j].Slug)
	})
	return result
}
