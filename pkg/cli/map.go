package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

var mapCmd = &cobra.Command{
	Use:   "map [feature]",
	Short: "Show a read-only map of canonical Kit documents",
	Long: `Render the current document hierarchy and explicit feature relationships.

Without a feature argument, shows the full project map.
With a feature argument, shows a focused map for that feature.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMap,
}

func init() {
	rootCmd.AddCommand(mapCmd)
}

func runMap(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	projectMap, err := feature.BuildProjectMap(projectRoot, cfg)
	if err != nil {
		return fmt.Errorf("failed to build document map: %w", err)
	}

	if len(args) == 0 {
		return outputProjectMap(cmd.OutOrStdout(), projectMap)
	}

	feat, err := loadFeatureWithState(cfg.SpecsPath(projectRoot), cfg, args[0])
	if err != nil {
		return err
	}

	for _, featureMap := range projectMap.Features {
		if featureMap.Feature.DirName == feat.DirName {
			return outputFeatureMap(cmd.OutOrStdout(), projectMap.GlobalDocuments, featureMap)
		}
	}

	return fmt.Errorf("feature '%s' not found in project map", feat.DirName)
}

func outputProjectMap(out io.Writer, projectMap *feature.ProjectMap) error {
	_, _ = fmt.Fprintln(out, "🗺️ Kit Map")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Global Docs")
	for _, doc := range projectMap.GlobalDocuments {
		_, _ = fmt.Fprintf(out, "- %s\n", formatMapDocument(doc))
	}

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Lifecycle")
	_, _ = fmt.Fprintln(out, "- CONSTITUTION.md -> BRAINSTORM.md (optional) -> SPEC.md -> PLAN.md -> TASKS.md -> implement -> reflect")

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Features")
	if len(projectMap.Features) == 0 {
		_, _ = fmt.Fprintln(out, "- none")
		return nil
	}

	for _, featureMap := range projectMap.Features {
		_, _ = fmt.Fprintf(out, "- %s [phase: %s] [paused: %s]\n", featureMap.Feature.DirName, featureMap.Feature.Phase, mapYesNo(featureMap.Feature.Paused))
		_, _ = fmt.Fprintln(out, "  documents:")
		for _, doc := range featureMap.Documents {
			_, _ = fmt.Fprintf(out, "  - %s\n", formatMapDocument(doc))
		}
		_, _ = fmt.Fprintln(out, "  relationships:")
		if len(featureMap.Outgoing) == 0 {
			_, _ = fmt.Fprintln(out, "  - none")
			continue
		}
		for _, edge := range featureMap.Outgoing {
			_, _ = fmt.Fprintf(out, "  - %s\n", formatOutgoingEdge(edge))
		}
	}

	return nil
}

func outputFeatureMap(out io.Writer, globalDocs []feature.MapDocument, featureMap feature.FeatureMap) error {
	_, _ = fmt.Fprintf(out, "🗺️ Kit Map: %s\n\n", featureMap.Feature.DirName)
	_, _ = fmt.Fprintln(out, "Global Docs")
	for _, doc := range globalDocs {
		_, _ = fmt.Fprintf(out, "- %s\n", formatMapDocument(doc))
	}

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "Feature\n- id: %s\n- phase: %s\n- paused: %s\n", featureMap.Feature.DirName, featureMap.Feature.Phase, mapYesNo(featureMap.Feature.Paused))
	_, _ = fmt.Fprintln(out, "- documents:")
	for _, doc := range featureMap.Documents {
		_, _ = fmt.Fprintf(out, "  - %s\n", formatMapDocument(doc))
	}

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Outgoing Relationships")
	if len(featureMap.Outgoing) == 0 {
		_, _ = fmt.Fprintln(out, "- none")
	} else {
		for _, edge := range featureMap.Outgoing {
			_, _ = fmt.Fprintf(out, "- %s\n", formatOutgoingEdge(edge))
		}
	}

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Incoming Relationships")
	if len(featureMap.Incoming) == 0 {
		_, _ = fmt.Fprintln(out, "- none")
		return nil
	}

	for _, edge := range featureMap.Incoming {
		_, _ = fmt.Fprintf(out, "- %s\n", formatIncomingEdge(edge))
	}

	return nil
}

func formatMapDocument(doc feature.MapDocument) string {
	return fmt.Sprintf("%s [%s, %s] via %s", doc.Path, requiredOptional(doc.Required), presentMissing(doc.Exists), doc.ManagedBy)
}

func formatOutgoingEdge(edge feature.RelationshipEdge) string {
	return fmt.Sprintf("%s %s -> %s [%s]", edge.SourceDoc, edge.Type, edge.TargetFeatureID, resolvedLabel(edge.Resolved))
}

func formatIncomingEdge(edge feature.RelationshipEdge) string {
	return fmt.Sprintf("%s via %s %s -> %s [%s]", edge.SourceFeatureID, edge.SourceDoc, edge.Type, edge.TargetFeatureID, resolvedLabel(edge.Resolved))
}

func requiredOptional(required bool) string {
	if required {
		return "required"
	}
	return "optional"
}

func presentMissing(exists bool) string {
	if exists {
		return "present"
	}
	return "missing"
}

func resolvedLabel(resolved bool) string {
	if resolved {
		return "resolved"
	}
	return "unresolved"
}

func mapYesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
