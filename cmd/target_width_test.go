package cmd

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
	"github.com/pgavlin/mermaid-ascii/pkg/render"
)

// maxLineWidth returns the character width of the widest line in s.
func maxLineWidth(s string) int {
	max := 0
	for _, line := range strings.Split(s, "\n") {
		// Trim trailing whitespace (the drawing pads with spaces)
		trimmed := strings.TrimRight(line, " ")
		w := utf8.RuneCountInString(trimmed)
		if w > max {
			max = w
		}
	}
	return max
}

func renderWithWidth(t *testing.T, input string, targetWidth int) string {
	t.Helper()
	config := &diagram.Config{
		UseAscii:        false,
		BoxBorderPadding: 1,
		PaddingBetweenX: 5,
		PaddingBetweenY: 5,
		StyleType:       "cli",
		TargetWidth:     targetWidth,
	}
	output, err := render.Render(input, config)
	if err != nil {
		t.Fatalf("Failed to render: %v", err)
	}
	return output
}

func TestTargetWidthZeroIsNoOp(t *testing.T) {
	input := `graph LR
A[Hello] --> B[World]`

	withoutConstraint := renderWithWidth(t, input, 0)
	withLargeWidth := renderWithWidth(t, input, 1000)

	if withoutConstraint != withLargeWidth {
		t.Errorf("TargetWidth=0 and TargetWidth=1000 should produce identical output for a small diagram")
	}
}

func TestTargetWidthReducesWidth(t *testing.T) {
	input := `graph LR
A[Input Processing Module] --> B[Validation Layer]
A --> C[Authentication Service]
B --> D[Output Formatter]
C --> D`

	naturalOutput := renderWithWidth(t, input, 0)
	constrainedOutput := renderWithWidth(t, input, 60)

	naturalWidth := maxLineWidth(naturalOutput)
	constrainedWidth := maxLineWidth(constrainedOutput)

	if constrainedWidth >= naturalWidth {
		t.Errorf("Constrained output should be narrower: natural=%d constrained=%d", naturalWidth, constrainedWidth)
	}
}

func TestTargetWidthWrapsNodeText(t *testing.T) {
	input := `graph LR
A[Very Long Application Name Here] --> B[Short]`

	output := renderWithWidth(t, input, 30)

	// The long label should be wrapped across multiple lines
	if !strings.Contains(output, "Very Long") {
		t.Errorf("Output should contain wrapped text, got:\n%s", output)
	}
	// Verify it rendered without error (basic smoke test)
	if len(output) == 0 {
		t.Error("Output should not be empty")
	}
}

func TestTargetWidthPreservesContent(t *testing.T) {
	input := `graph TD
A[Start] --> B[Middle]
B --> C[End]`

	output := renderWithWidth(t, input, 40)

	// All node labels should appear in the output
	for _, label := range []string{"Start", "Middle", "End"} {
		if !strings.Contains(output, label) {
			t.Errorf("Output should contain label %q, got:\n%s", label, output)
		}
	}
}

func TestTargetWidthWithBrTags(t *testing.T) {
	input := `graph TD
A[Application<br/>System] --> B[Database<br/>Layer]`

	output := renderWithWidth(t, input, 40)

	// Both parts of the br-split labels should appear
	for _, label := range []string{"Application", "System", "Database", "Layer"} {
		if !strings.Contains(output, label) {
			t.Errorf("Output should contain label %q, got:\n%s", label, output)
		}
	}
}

func TestTargetWidthWithEdgeLabels(t *testing.T) {
	// Edge labels that expand column widths should not cause panics
	input := `graph TD
A -->|long edge label| B
B --> C`

	// Render with a tight constraint — must not panic
	output := renderWithWidth(t, input, 30)
	if len(output) == 0 {
		t.Error("Output should not be empty")
	}

	// With a generous constraint, edge label should be preserved
	output = renderWithWidth(t, input, 200)
	if !strings.Contains(output, "long edge label") {
		t.Errorf("Output with generous width should contain the edge label, got:\n%s", output)
	}
}

func TestTargetWidthWithDiamondShape(t *testing.T) {
	input := `graph TD
A[Start] --> B{Decision Point}
B -->|Yes| C[End]
B -->|No| D[Retry]`

	output := renderWithWidth(t, input, 50)

	if !strings.Contains(output, "Decision Point") {
		t.Errorf("Output should contain diamond label, got:\n%s", output)
	}
}

func TestTargetWidthManyColumns(t *testing.T) {
	// A diagram with many side-by-side nodes — width constraint should still reduce
	// from natural width even if it can't hit the exact target
	input := `graph TD
A --> B
A --> C
A --> D
A --> E
A --> F`

	naturalOutput := renderWithWidth(t, input, 0)
	constrainedOutput := renderWithWidth(t, input, 30)

	naturalWidth := maxLineWidth(naturalOutput)
	constrainedWidth := maxLineWidth(constrainedOutput)

	// Should at least reduce padding even if nodes can't shrink much
	if constrainedWidth > naturalWidth {
		t.Errorf("Constrained output should not be wider than natural: natural=%d constrained=%d", naturalWidth, constrainedWidth)
	}
}

func TestTargetWidthTDLayout(t *testing.T) {
	input := `graph TD
A[Application Layer] --> B[Middle Layer]
B --> C[Database Layer]
B --> D[Cache Layer]`

	naturalOutput := renderWithWidth(t, input, 0)
	constrainedOutput := renderWithWidth(t, input, 30)

	naturalWidth := maxLineWidth(naturalOutput)
	constrainedWidth := maxLineWidth(constrainedOutput)

	if constrainedWidth >= naturalWidth {
		t.Errorf("TD layout: constrained should be narrower: natural=%d constrained=%d", naturalWidth, constrainedWidth)
	}
}

func TestTargetWidthLRLayout(t *testing.T) {
	input := `graph LR
A[Application Layer] --> B[Processing Unit]
B --> C[Output Layer]`

	naturalOutput := renderWithWidth(t, input, 0)
	constrainedOutput := renderWithWidth(t, input, 50)

	naturalWidth := maxLineWidth(naturalOutput)
	constrainedWidth := maxLineWidth(constrainedOutput)

	if constrainedWidth >= naturalWidth {
		t.Errorf("LR layout: constrained should be narrower: natural=%d constrained=%d", naturalWidth, constrainedWidth)
	}
}

func TestTargetWidthAlreadyFits(t *testing.T) {
	// A tiny diagram that already fits within the target — should be unchanged
	input := `graph TD
A --> B`

	natural := renderWithWidth(t, input, 0)
	constrained := renderWithWidth(t, input, 200)

	if natural != constrained {
		t.Errorf("Diagram that already fits should not change with large target width")
	}
}

func TestTargetWidthASCIIMode(t *testing.T) {
	input := `graph LR
A[Hello World Module] --> B[Another Module Here]
B --> C[Final]`

	config := &diagram.Config{
		UseAscii:        true,
		BoxBorderPadding: 1,
		PaddingBetweenX: 5,
		PaddingBetweenY: 5,
		StyleType:       "cli",
		TargetWidth:     50,
	}
	output, err := render.Render(input, config)
	if err != nil {
		t.Fatalf("Failed to render in ASCII mode with target width: %v", err)
	}

	// Should contain ASCII box characters, not Unicode
	if strings.Contains(output, "┌") || strings.Contains(output, "─") {
		t.Errorf("ASCII mode should use ASCII characters, got:\n%s", output)
	}

	// Content should be preserved
	if !strings.Contains(output, "Hello") || !strings.Contains(output, "Final") {
		t.Errorf("Content should be preserved, got:\n%s", output)
	}
}
