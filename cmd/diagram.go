package cmd

import (
	"fmt"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
	"github.com/pgavlin/mermaid-ascii/pkg/sequence"
)

// DiagramDetector can detect and create a specific diagram type.
type DiagramDetector struct {
	Detect func(input string) bool
	Create func() diagram.Diagram
}

// diagramRegistry holds all registered diagram detectors, checked in order.
// Graph is always the fallback and should not be in this list.
var diagramRegistry []DiagramDetector

func init() {
	// Register built-in diagram types
	RegisterDiagram(DiagramDetector{
		Detect: sequence.IsSequenceDiagram,
		Create: func() diagram.Diagram { return &SequenceDiagram{} },
	})
}

// RegisterDiagram adds a diagram detector to the registry.
// Detectors are checked in registration order; graph is always the fallback.
func RegisterDiagram(d DiagramDetector) {
	diagramRegistry = append(diagramRegistry, d)
}

func DiagramFactory(input string) (diagram.Diagram, error) {
	input = strings.TrimSpace(input)

	for _, detector := range diagramRegistry {
		if detector.Detect(input) {
			return detector.Create(), nil
		}
	}

	// Graph is the default fallback
	return &GraphDiagram{}, nil
}

type SequenceDiagram struct {
	parsed *sequence.SequenceDiagram
}

func (sd *SequenceDiagram) Parse(input string) error {
	parsed, err := sequence.Parse(input)
	if err != nil {
		return err
	}
	sd.parsed = parsed
	return nil
}

func (sd *SequenceDiagram) Render(config *diagram.Config) (string, error) {
	if sd.parsed == nil {
		return "", fmt.Errorf("sequence diagram not parsed: call Parse() before Render()")
	}
	return sequence.Render(sd.parsed, config)
}

func (sd *SequenceDiagram) Type() string {
	return "sequence"
}

type GraphDiagram struct {
	properties *graphProperties
}

func (gd *GraphDiagram) Parse(input string) error {
	properties, err := mermaidFileToMap(input, "cli")
	if err != nil {
		return err
	}
	gd.properties = properties
	return nil
}

func (gd *GraphDiagram) Render(config *diagram.Config) (string, error) {
	if gd.properties == nil {
		return "", fmt.Errorf("graph diagram not parsed: call Parse() before Render()")
	}

	if config == nil {
		config = diagram.DefaultConfig()
	}

	styleType := config.StyleType
	if styleType == "" {
		styleType = "cli"
	}
	gd.properties.styleType = styleType
	gd.properties.useAscii = config.UseAscii
	gd.properties.boxBorderPadding = config.BoxBorderPadding
	gd.properties.showCoords = config.ShowCoords
	if config.PaddingBetweenX > 0 {
		gd.properties.paddingX = config.PaddingBetweenX
	}
	if config.PaddingBetweenY > 0 {
		gd.properties.paddingY = config.PaddingBetweenY
	}

	return drawMap(gd.properties), nil
}

func (gd *GraphDiagram) Type() string {
	return "graph"
}
