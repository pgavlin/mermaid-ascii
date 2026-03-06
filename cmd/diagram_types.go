package cmd

import (
	"fmt"

	"github.com/pgavlin/mermaid-ascii/pkg/architecture"
	"github.com/pgavlin/mermaid-ascii/pkg/blockdiagram"
	"github.com/pgavlin/mermaid-ascii/pkg/c4"
	"github.com/pgavlin/mermaid-ascii/pkg/classdiagram"
	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
	"github.com/pgavlin/mermaid-ascii/pkg/erdiagram"
	"github.com/pgavlin/mermaid-ascii/pkg/gantt"
	"github.com/pgavlin/mermaid-ascii/pkg/gitgraph"
	"github.com/pgavlin/mermaid-ascii/pkg/journey"
	"github.com/pgavlin/mermaid-ascii/pkg/kanban"
	"github.com/pgavlin/mermaid-ascii/pkg/mindmap"
	"github.com/pgavlin/mermaid-ascii/pkg/packet"
	"github.com/pgavlin/mermaid-ascii/pkg/piechart"
	"github.com/pgavlin/mermaid-ascii/pkg/quadrant"
	"github.com/pgavlin/mermaid-ascii/pkg/requirement"
	"github.com/pgavlin/mermaid-ascii/pkg/sankey"
	"github.com/pgavlin/mermaid-ascii/pkg/statediagram"
	"github.com/pgavlin/mermaid-ascii/pkg/timeline"
	"github.com/pgavlin/mermaid-ascii/pkg/xychart"
	"github.com/pgavlin/mermaid-ascii/pkg/zenuml"
)

func init() {
	// Register all diagram types. Graph is the fallback and not registered here.
	// Sequence is registered in diagram.go's init().

	RegisterDiagram(DiagramDetector{
		Detect: classdiagram.IsClassDiagram,
		Create: func() diagram.Diagram { return &classDiagramWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: statediagram.IsStateDiagram,
		Create: func() diagram.Diagram { return &stateDiagramWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: erdiagram.IsERDiagram,
		Create: func() diagram.Diagram { return &erDiagramWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: gantt.IsGanttDiagram,
		Create: func() diagram.Diagram { return &ganttDiagramWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: piechart.IsPieChart,
		Create: func() diagram.Diagram { return &pieChartWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: mindmap.IsMindmapDiagram,
		Create: func() diagram.Diagram { return &mindmapWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: timeline.IsTimelineDiagram,
		Create: func() diagram.Diagram { return &timelineWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: gitgraph.IsGitGraph,
		Create: func() diagram.Diagram { return &gitgraphWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: journey.IsJourneyDiagram,
		Create: func() diagram.Diagram { return &journeyWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: quadrant.IsQuadrantChart,
		Create: func() diagram.Diagram { return &quadrantWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: xychart.IsXYChart,
		Create: func() diagram.Diagram { return &xychartWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: c4.IsC4Diagram,
		Create: func() diagram.Diagram { return &c4Wrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: requirement.IsRequirementDiagram,
		Create: func() diagram.Diagram { return &requirementWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: blockdiagram.IsBlockDiagram,
		Create: func() diagram.Diagram { return &blockDiagramWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: sankey.IsSankeyDiagram,
		Create: func() diagram.Diagram { return &sankeyWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: packet.IsPacketDiagram,
		Create: func() diagram.Diagram { return &packetWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: kanban.IsKanbanBoard,
		Create: func() diagram.Diagram { return &kanbanWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: architecture.IsArchitectureDiagram,
		Create: func() diagram.Diagram { return &architectureWrapper{} },
	})
	RegisterDiagram(DiagramDetector{
		Detect: zenuml.IsZenUML,
		Create: func() diagram.Diagram { return &zenumlWrapper{} },
	})
}

// --- Class Diagram ---

type classDiagramWrapper struct {
	parsed *classdiagram.ClassDiagram
}

func (w *classDiagramWrapper) Parse(input string) error {
	p, err := classdiagram.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *classDiagramWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return classdiagram.Render(w.parsed, config)
}

func (w *classDiagramWrapper) Type() string { return "classDiagram" }

// --- State Diagram ---

type stateDiagramWrapper struct {
	parsed *statediagram.StateDiagram
}

func (w *stateDiagramWrapper) Parse(input string) error {
	p, err := statediagram.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *stateDiagramWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return statediagram.Render(w.parsed, config)
}

func (w *stateDiagramWrapper) Type() string { return "stateDiagram" }

// --- ER Diagram ---

type erDiagramWrapper struct {
	parsed *erdiagram.ERDiagram
}

func (w *erDiagramWrapper) Parse(input string) error {
	p, err := erdiagram.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *erDiagramWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return erdiagram.Render(w.parsed, config)
}

func (w *erDiagramWrapper) Type() string { return "erDiagram" }

// --- Gantt ---

type ganttDiagramWrapper struct {
	parsed *gantt.GanttDiagram
}

func (w *ganttDiagramWrapper) Parse(input string) error {
	p, err := gantt.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *ganttDiagramWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return gantt.Render(w.parsed, config)
}

func (w *ganttDiagramWrapper) Type() string { return "gantt" }

// --- Pie Chart ---

type pieChartWrapper struct {
	parsed *piechart.PieChart
}

func (w *pieChartWrapper) Parse(input string) error {
	p, err := piechart.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *pieChartWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return piechart.Render(w.parsed, config)
}

func (w *pieChartWrapper) Type() string { return "pie" }

// --- Mindmap ---

type mindmapWrapper struct {
	parsed *mindmap.MindmapDiagram
}

func (w *mindmapWrapper) Parse(input string) error {
	p, err := mindmap.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *mindmapWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return mindmap.Render(w.parsed, config)
}

func (w *mindmapWrapper) Type() string { return "mindmap" }

// --- Timeline ---

type timelineWrapper struct {
	parsed *timeline.TimelineDiagram
}

func (w *timelineWrapper) Parse(input string) error {
	p, err := timeline.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *timelineWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return timeline.Render(w.parsed, config)
}

func (w *timelineWrapper) Type() string { return "timeline" }

// --- Git Graph ---

type gitgraphWrapper struct {
	parsed *gitgraph.GitGraph
}

func (w *gitgraphWrapper) Parse(input string) error {
	p, err := gitgraph.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *gitgraphWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return gitgraph.Render(w.parsed, config)
}

func (w *gitgraphWrapper) Type() string { return "gitGraph" }

// --- Journey ---

type journeyWrapper struct {
	parsed *journey.JourneyDiagram
}

func (w *journeyWrapper) Parse(input string) error {
	p, err := journey.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *journeyWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return journey.Render(w.parsed, config)
}

func (w *journeyWrapper) Type() string { return "journey" }

// --- Quadrant ---

type quadrantWrapper struct {
	parsed *quadrant.QuadrantChart
}

func (w *quadrantWrapper) Parse(input string) error {
	p, err := quadrant.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *quadrantWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return quadrant.Render(w.parsed, config)
}

func (w *quadrantWrapper) Type() string { return "quadrantChart" }

// --- XY Chart ---

type xychartWrapper struct {
	parsed *xychart.XYChart
}

func (w *xychartWrapper) Parse(input string) error {
	p, err := xychart.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *xychartWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return xychart.Render(w.parsed, config)
}

func (w *xychartWrapper) Type() string { return "xychart-beta" }

// --- C4 ---

type c4Wrapper struct {
	parsed *c4.C4Diagram
}

func (w *c4Wrapper) Parse(input string) error {
	p, err := c4.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *c4Wrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return c4.Render(w.parsed, config)
}

func (w *c4Wrapper) Type() string { return "C4Context" }

// --- Requirement ---

type requirementWrapper struct {
	parsed *requirement.RequirementDiagram
}

func (w *requirementWrapper) Parse(input string) error {
	p, err := requirement.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *requirementWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return requirement.Render(w.parsed, config)
}

func (w *requirementWrapper) Type() string { return "requirementDiagram" }

// --- Block Diagram ---

type blockDiagramWrapper struct {
	parsed *blockdiagram.BlockDiagram
}

func (w *blockDiagramWrapper) Parse(input string) error {
	p, err := blockdiagram.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *blockDiagramWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return blockdiagram.Render(w.parsed, config)
}

func (w *blockDiagramWrapper) Type() string { return "block-beta" }

// --- Sankey ---

type sankeyWrapper struct {
	parsed *sankey.SankeyDiagram
}

func (w *sankeyWrapper) Parse(input string) error {
	p, err := sankey.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *sankeyWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return sankey.Render(w.parsed, config)
}

func (w *sankeyWrapper) Type() string { return "sankey-beta" }

// --- Packet ---

type packetWrapper struct {
	parsed *packet.PacketDiagram
}

func (w *packetWrapper) Parse(input string) error {
	p, err := packet.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *packetWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return packet.Render(w.parsed, config)
}

func (w *packetWrapper) Type() string { return "packet-beta" }

// --- Kanban ---

type kanbanWrapper struct {
	parsed *kanban.KanbanBoard
}

func (w *kanbanWrapper) Parse(input string) error {
	p, err := kanban.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *kanbanWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return kanban.Render(w.parsed, config)
}

func (w *kanbanWrapper) Type() string { return "kanban" }

// --- Architecture ---

type architectureWrapper struct {
	parsed *architecture.ArchitectureDiagram
}

func (w *architectureWrapper) Parse(input string) error {
	p, err := architecture.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *architectureWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return architecture.Render(w.parsed, config)
}

func (w *architectureWrapper) Type() string { return "architecture-beta" }

// --- ZenUML ---

type zenumlWrapper struct {
	parsed *zenuml.ZenUMLDiagram
}

func (w *zenumlWrapper) Parse(input string) error {
	p, err := zenuml.Parse(input)
	if err != nil {
		return err
	}
	w.parsed = p
	return nil
}

func (w *zenumlWrapper) Render(config *diagram.Config) (string, error) {
	if w.parsed == nil {
		return "", fmt.Errorf("not parsed")
	}
	return zenuml.Render(w.parsed, config)
}

func (w *zenumlWrapper) Type() string { return "zenuml" }
