package architecture

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const ArchitectureBetaKeyword = "architecture-beta"

var (
	// service name(icon)[label]
	serviceRegex = regexp.MustCompile(`^\s*service\s+(\w+)(?:\(([^)]*)\))?(?:\[([^\]]*)\])?\s*$`)

	// group name(icon)[label] {
	groupStartRegex = regexp.MustCompile(`^\s*group\s+(\w+)(?:\(([^)]*)\))?(?:\[([^\]]*)\])?\s*\{\s*$`)

	// junction name
	junctionRegex = regexp.MustCompile(`^\s*junction\s+(\w+)\s*$`)

	// service1:edge_pos -- service2:edge_pos  or service1:edge_pos --> service2:edge_pos
	connectionRegex = regexp.MustCompile(`^\s*(\w+)(?::(\w+))?\s*(-->|--)\s*(\w+)(?::(\w+))?\s*$`)

	// end of group
	groupEndRegex = regexp.MustCompile(`^\s*\}\s*$`)
)

// Service represents a service/node in the architecture diagram.
type Service struct {
	ID    string
	Icon  string
	Label string
}

// Group represents a group of services.
type Group struct {
	ID       string
	Icon     string
	Label    string
	Services []*Service
	Groups   []*Group
}

// Connection represents a connection between two services.
type Connection struct {
	From      string
	FromEdge  string
	To        string
	ToEdge    string
	Directed  bool // true for -->, false for --
}

// ArchitectureDiagram represents a parsed architecture diagram.
type ArchitectureDiagram struct {
	Services    []*Service
	Groups      []*Group
	Connections []*Connection
}

// IsArchitectureDiagram returns true if the input starts with architecture-beta keyword.
func IsArchitectureDiagram(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return strings.HasPrefix(trimmed, ArchitectureBetaKeyword)
	}
	return false
}

// Parse parses an architecture diagram.
func Parse(input string) (*ArchitectureDiagram, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if !strings.HasPrefix(strings.TrimSpace(lines[0]), ArchitectureBetaKeyword) {
		return nil, fmt.Errorf("expected %q keyword", ArchitectureBetaKeyword)
	}

	d := &ArchitectureDiagram{}
	_, err := parseArchLines(d, lines[1:], nil)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func parseArchLines(d *ArchitectureDiagram, lines []string, currentGroup *Group) (int, error) {
	i := 0
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			i++
			continue
		}

		// End of group
		if groupEndRegex.MatchString(trimmed) {
			return i + 1, nil
		}

		// Group start
		if m := groupStartRegex.FindStringSubmatch(trimmed); m != nil {
			g := &Group{
				ID:    m[1],
				Icon:  m[2],
				Label: m[3],
			}
			if g.Label == "" {
				g.Label = g.ID
			}
			i++
			consumed, err := parseArchLines(d, lines[i:], g)
			if err != nil {
				return 0, err
			}
			i += consumed
			if currentGroup != nil {
				currentGroup.Groups = append(currentGroup.Groups, g)
			} else {
				d.Groups = append(d.Groups, g)
			}
			continue
		}

		// Junction
		if m := junctionRegex.FindStringSubmatch(trimmed); m != nil {
			svc := &Service{ID: m[1], Label: m[1]}
			if currentGroup != nil {
				currentGroup.Services = append(currentGroup.Services, svc)
			} else {
				d.Services = append(d.Services, svc)
			}
			i++
			continue
		}

		// Service
		if m := serviceRegex.FindStringSubmatch(trimmed); m != nil {
			svc := &Service{
				ID:    m[1],
				Icon:  m[2],
				Label: m[3],
			}
			if svc.Label == "" {
				svc.Label = svc.ID
			}
			if currentGroup != nil {
				currentGroup.Services = append(currentGroup.Services, svc)
			} else {
				d.Services = append(d.Services, svc)
			}
			i++
			continue
		}

		// Connection
		if m := connectionRegex.FindStringSubmatch(trimmed); m != nil {
			conn := &Connection{
				From:     m[1],
				FromEdge: m[2],
				To:       m[4],
				ToEdge:   m[5],
				Directed: m[3] == "-->",
			}
			d.Connections = append(d.Connections, conn)
			i++
			continue
		}

		// Skip unrecognized lines
		i++
	}
	return i, nil
}
