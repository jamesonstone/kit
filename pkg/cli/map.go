package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jamesonstone/kit/internal/config"
	"github.com/jamesonstone/kit/internal/feature"
)

var mapContext bool

var mapJSON bool

var mapAll bool

var mapCmd = &cobra.Command{
	Use:   "map [feature]",
	Short: "Show a read-only map of canonical Kit documents",
	Long: `Render the current document hierarchy and explicit feature relationships.

Without a feature argument, opens the interactive feature selector.
With a feature argument, shows a focused map for that feature.
Use --all to show the full project map.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMap,
}

func init() {
	mapCmd.Flags().BoolVar(&mapContext, "context", false, "show a focused reference read plan for the selected feature")
	mapCmd.Flags().BoolVar(&mapJSON, "json", false, "output the document map as JSON")
	mapCmd.Flags().BoolVar(&mapAll, "all", false, "show the full project map")
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

	if mapAll && len(args) > 0 {
		return fmt.Errorf("--all cannot be used with a feature argument")
	}
	if mapAll && mapContext {
		return fmt.Errorf("--all cannot be used with --context")
	}

	var selectedFeature *feature.Feature
	if len(args) == 0 && !mapAll {
		promptOut := cmd.OutOrStdout()
		if mapJSON {
			promptOut = cmd.ErrOrStderr()
		}
		selectedFeature, err = selectFeatureForMap(cfg.SpecsPath(projectRoot), cfg, os.Stdin, promptOut)
		if err != nil {
			return err
		}
	} else if len(args) == 1 {
		selectedFeature, err = loadFeatureWithState(cfg.SpecsPath(projectRoot), cfg, args[0])
		if err != nil {
			return err
		}
	}

	projectMap, err := feature.BuildProjectMap(projectRoot, cfg)
	if err != nil {
		return fmt.Errorf("failed to build document map: %w", err)
	}

	if mapAll {
		if mapJSON {
			return outputMapJSON(cmd.OutOrStdout(), projectMap)
		}
		return outputProjectMap(cmd.OutOrStdout(), projectMap)
	}

	return outputSelectedFeatureMap(cmd.OutOrStdout(), projectMap, selectedFeature)
}

func selectFeatureForMap(specsDir string, cfg *config.Config, input io.Reader, output io.Writer) (*feature.Feature, error) {
	features, err := feature.ListFeaturesWithState(specsDir, cfg)
	if err != nil {
		return nil, err
	}
	if len(features) == 0 {
		return nil, fmt.Errorf("no features found\n\nRun 'kit spec <feature>' to start a v2 feature, or 'kit legacy brainstorm' for staged migration work")
	}

	printSelectionHeaderTo(output, "Select a feature to map:")
	for i, f := range features {
		pausedSuffix := ""
		if f.Paused {
			pausedSuffix = ", paused"
		}
		_, _ = fmt.Fprintf(output, "  [%d] %s (%s%s)\n", i+1, f.DirName, f.Phase, pausedSuffix)
	}
	_, _ = fmt.Fprintln(output)
	_, _ = fmt.Fprint(output, selectionPrompt(output))

	reader := bufio.NewReader(input)
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)
	if selection == "" {
		return nil, fmt.Errorf("no feature selected; run 'kit map --all' for the full project map or pass a feature name")
	}

	num, err := strconv.Atoi(selection)
	if err != nil || num < 1 || num > len(features) {
		return nil, fmt.Errorf("invalid selection: %s", selection)
	}

	selected := features[num-1]
	return &selected, nil
}

func outputSelectedFeatureMap(out io.Writer, projectMap *feature.ProjectMap, feat *feature.Feature) error {
	if feat == nil {
		return fmt.Errorf("feature is required")
	}

	for _, featureMap := range projectMap.Features {
		if featureMap.Feature.DirName == feat.DirName {
			if mapContext {
				if mapJSON {
					return outputContextPlanJSON(out, featureMap)
				}
				return outputFeatureContextPlan(out, featureMap)
			}
			if mapJSON {
				return outputMapJSON(out, featureMap)
			}
			return outputFeatureMap(out, projectMap.GlobalDocuments, featureMap, filterMapWarnings(projectMap.Warnings, feat.DirName))
		}
	}

	return fmt.Errorf("feature '%s' not found in project map", feat.DirName)
}

func outputProjectMap(out io.Writer, projectMap *feature.ProjectMap) error {
	glyphs := selectMapGlyphs()
	style := styleForWriter(out)

	_, _ = fmt.Fprintln(out, mapTitle(style, "Kit Map"))
	_, _ = fmt.Fprintln(out)
	outputGlobalDocs(out, style, glyphs, projectMap.GlobalDocuments)
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.label("Lifecycle"))
	_, _ = fmt.Fprintf(out, "%s %s\n", mapTreePrefix(style, glyphs.Last), mapLifecycleLine(style))
	_, _ = fmt.Fprintln(out)
	outputFeatureDocKey(out, style, glyphs)
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.label("Feature Graph"))
	if len(projectMap.Features) == 0 {
		_, _ = fmt.Fprintf(out, "%s %s\n", mapTreePrefix(style, glyphs.Last), mapMutedIfEnabled(style, "none"))
	} else {
		for i, featureMap := range projectMap.Features {
			if i > 0 {
				_, _ = fmt.Fprintln(out)
			}
			renderFeatureCard(out, style, glyphs, featureMap)
			renderProjectEdges(out, style, glyphs, featureMap.Outgoing)
			renderReferenceLinks(out, style, glyphs, featureMap.References, "no reference links")
		}
	}

	outputMapWarnings(out, style, glyphs, projectMap.Warnings)

	return nil
}

func outputFeatureMap(out io.Writer, globalDocs []feature.MapDocument, featureMap feature.FeatureMap, warnings []feature.MapWarning) error {
	glyphs := selectMapGlyphs()
	style := styleForWriter(out)

	_, _ = fmt.Fprintf(out, "%s: %s\n\n", mapTitle(style, "Kit Map"), mapFeatureName(style, featureMap.Feature.DirName))
	outputGlobalDocs(out, style, glyphs, globalDocs)
	_, _ = fmt.Fprintln(out)
	outputFeatureDocKey(out, style, glyphs)
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.label("Feature Focus"))
	renderFeatureCard(out, style, glyphs, featureMap)

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.label("Incoming Relationships"))
	renderIncomingEdges(out, style, glyphs, featureMap)

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.label("Outgoing Relationships"))
	renderProjectEdges(out, style, glyphs, featureMap.Outgoing)

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, style.label("Reference Links"))
	renderReferenceLinks(out, style, glyphs, featureMap.References, "none")

	outputMapWarnings(out, style, glyphs, warnings)

	return nil
}
