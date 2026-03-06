package requirement

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const RequirementDiagramKeyword = "requirementDiagram"

var (
	// requirement name {
	requirementStartRegex = regexp.MustCompile(`^\s*(requirement|functionalRequirement|interfaceRequirement|performanceRequirement|physicalRequirement|designConstraint)\s+(.+?)\s*\{\s*$`)

	// element name {
	elementStartRegex = regexp.MustCompile(`^\s*element\s+(.+?)\s*\{\s*$`)

	// key: value inside a block
	kvRegex = regexp.MustCompile(`^\s*(\w+)\s*:\s*(.+?)\s*$`)

	// closing brace
	closeBraceRegex = regexp.MustCompile(`^\s*\}\s*$`)

	// name - relationship -> name2
	relationshipRegex = regexp.MustCompile(`^\s*(.+?)\s+-\s+(traces|copies|derives|satisfies|verifies|refines|contains)\s+->\s+(.+?)\s*$`)
)

// Requirement represents a requirement node.
type Requirement struct {
	Name         string
	Type         string // "requirement", "functionalRequirement", etc.
	ID           string
	Text         string
	Risk         string
	VerifyMethod string
}

// ReqElement represents an element node.
type ReqElement struct {
	Name   string
	Type   string
	DocRef string
}

// ReqRelationship represents a relationship between requirements/elements.
type ReqRelationship struct {
	Source string
	Target string
	Type   string // "satisfies", "traces", "copies", "derives", "verifies", "refines", "contains"
}

// RequirementDiagram represents a parsed requirement diagram.
type RequirementDiagram struct {
	Requirements  []*Requirement
	Elements      []*ReqElement
	Relationships []*ReqRelationship
}

// IsRequirementDiagram returns true if the input starts with requirementDiagram keyword.
func IsRequirementDiagram(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return strings.HasPrefix(trimmed, RequirementDiagramKeyword)
	}
	return false
}

// Parse parses a requirement diagram.
func Parse(input string) (*RequirementDiagram, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if !strings.HasPrefix(strings.TrimSpace(lines[0]), RequirementDiagramKeyword) {
		return nil, fmt.Errorf("expected %q keyword", RequirementDiagramKeyword)
	}

	d := &RequirementDiagram{}

	i := 1
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			i++
			continue
		}

		// Requirement block
		if m := requirementStartRegex.FindStringSubmatch(trimmed); m != nil {
			req := &Requirement{Type: m[1], Name: strings.Trim(m[2], `"`)}
			i++
			for i < len(lines) {
				inner := strings.TrimSpace(lines[i])
				if closeBraceRegex.MatchString(inner) {
					i++
					break
				}
				if kv := kvRegex.FindStringSubmatch(inner); kv != nil {
					switch kv[1] {
					case "id":
						req.ID = strings.Trim(kv[2], `"`)
					case "text":
						req.Text = strings.Trim(kv[2], `"`)
					case "risk":
						req.Risk = strings.Trim(kv[2], `"`)
					case "verifymethod":
						req.VerifyMethod = strings.Trim(kv[2], `"`)
					}
				}
				i++
			}
			d.Requirements = append(d.Requirements, req)
			continue
		}

		// Element block
		if m := elementStartRegex.FindStringSubmatch(trimmed); m != nil {
			elem := &ReqElement{Name: strings.Trim(m[1], `"`)}
			i++
			for i < len(lines) {
				inner := strings.TrimSpace(lines[i])
				if closeBraceRegex.MatchString(inner) {
					i++
					break
				}
				if kv := kvRegex.FindStringSubmatch(inner); kv != nil {
					switch kv[1] {
					case "type":
						elem.Type = strings.Trim(kv[2], `"`)
					case "docref":
						elem.DocRef = strings.Trim(kv[2], `"`)
					case "docRef":
						elem.DocRef = strings.Trim(kv[2], `"`)
					}
				}
				i++
			}
			d.Elements = append(d.Elements, elem)
			continue
		}

		// Relationship
		if m := relationshipRegex.FindStringSubmatch(trimmed); m != nil {
			rel := &ReqRelationship{
				Source: strings.Trim(m[1], `"`),
				Type:   m[2],
				Target: strings.Trim(m[3], `"`),
			}
			d.Relationships = append(d.Relationships, rel)
			i++
			continue
		}

		i++
	}

	return d, nil
}
