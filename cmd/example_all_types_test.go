package cmd

import (
	"fmt"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
	"github.com/pgavlin/mermaid-ascii/pkg/render"
)

// ExampleRender_allTypes demonstrates rendering one diagram of each
// supported type through the render.Render entry point.
func ExampleRender_allTypes() {
	config := diagram.NewTestConfig(true, "cli")

	diagrams := []struct {
		name  string
		input string
	}{
		{"graph", `graph LR
    A[Start] --> B[End]`},

		{"sequence", `sequenceDiagram
    Alice->>Bob: Hello
    Bob-->>Alice: Hi`},

		{"classDiagram", `classDiagram
    class Animal {
        +String name
        +makeSound()
    }
    class Dog {
        +fetch()
    }
    Animal <|-- Dog`},

		{"stateDiagram", `stateDiagram-v2
    [*] --> Active
    Active --> Inactive`},

		{"erDiagram", `erDiagram
    CUSTOMER ||--o{ ORDER : places`},

		{"gantt", `gantt
    dateFormat YYYY-MM-DD
    title Project Plan
    section Phase 1
    Design :a1, 2024-01-01, 7d
    Build  :a2, after a1, 14d`},

		{"pie", `pie title Pets
    "Dogs" : 40
    "Cats" : 30
    "Fish" : 20`},

		{"mindmap", `mindmap
    root((Project))
        Planning
            Requirements
            Design
        Development
            Frontend
            Backend`},

		{"timeline", `timeline
    title History
    2020 : Founded
    2021 : Series A
    2022 : Launch`},

		{"gitGraph", `gitGraph
    commit
    commit
    branch develop
    commit
    checkout main
    merge develop
    commit`},

		{"journey", `journey
    title My Day
    section Morning
        Wake up: 3: Me
        Breakfast: 5: Me`},

		{"quadrantChart", `quadrantChart
    title Skills
    x-axis Low --> High
    y-axis Low --> High
    Go: [0.8, 0.9]
    Python: [0.7, 0.6]`},

		{"xychart-beta", `xychart-beta
    title Sales
    x-axis [Q1, Q2, Q3, Q4]
    y-axis "Revenue"
    bar [10, 20, 15, 25]`},

		{"C4Context", `C4Context
    Person(user, "User", "A person")
    System(app, "App", "The application")
    Rel(user, app, "Uses")`},

		{"requirementDiagram", `requirementDiagram
    requirement test_req {
        id: 1
        text: the system shall do something
    }
    element test_entity {
        type: simulation
    }
    test_entity - satisfies -> test_req`},

		{"block-beta", `block-beta
    columns 3
    A["Alpha"]
    B["Beta"]
    C["Gamma"]
    D["Delta"]
    E["Epsilon"]
    F["Zeta"]`},

		{"sankey-beta", `sankey-beta
Source,Target,25
Source,Other,15
Other,Final,10`},

		{"packet-beta", `packet-beta
    0-7: "Header"
    8-15: "Payload"
    16-23: "Checksum"
    24-31: "Footer"`},

		{"kanban", `kanban
Todo
    Task 1
    Task 2
Doing
    Task 3
Done
    Task 4`},

		{"architecture-beta", `architecture-beta
    service api(server)[API]
    service db(database)[Database]
    api --> db`},

		{"zenuml", `zenuml
    Alice.hello() {
        Bob.hi()
    }`},
	}

	for _, d := range diagrams {
		output, err := render.Render(d.input, config)
		if err != nil {
			fmt.Printf("--- %s ---\nERROR: %v\n\n", d.name, err)
			continue
		}
		fmt.Printf("--- %s ---\n%s\n", d.name, strings.TrimRight(output, "\n")+"\n")
	}

	// Output:
	// --- graph ---
	// +-------+     +-----+
	// |       |     |     |
	// | Start |---->| End |
	// |       |     |     |
	// +-------+     +-----+
	//
	// --- sequence ---
	// +-------+     +-----+
	// | Alice |     | Bob |
	// +---+---+     +--+--+
	//     |            |
	//     | Hello      |
	//     +----------->|
	//     |            |
	//     | Hi         |
	//     |<...........+
	//     |            |
	//
	// --- classDiagram ---
	// +--------------+
	// |    Animal    |
	// +--------------+
	// | +String name |
	// | +makeSound() |
	// +--------------+
	//
	// +----------+
	// |   Dog    |
	// +----------+
	// | +fetch() |
	// +----------+
	//
	// Animal <|-- Dog
	//
	// --- stateDiagram ---
	//     (*)
	//       |
	//       v
	// +----------+
	// |  Active  |
	// +----------+
	//       |
	//       v
	// +----------+
	// | Inactive |
	// +----------+
	//
	// --- erDiagram ---
	// +------------+
	// |  CUSTOMER  |
	// +------------+
	// |            |
	// +------------+
	//
	// +------------+
	// |   ORDER    |
	// +------------+
	// |            |
	// +------------+
	//
	// CUSTOMER ||-->o ORDER : places
	//
	// --- gantt ---
	//                                Project Plan
	//
	//               01/01          01/06          01/11          01/16          01/22
	//              +------------------------------------------------------------+
	// Phase 1
	// Design       |####################........................................|
	// Build        |....................########################################|
	//              +------------------------------------------------------------+
	//
	// --- pie ---
	//                           Pets
	//
	// Dogs |#################                       |  44.4%
	// Cats |=============                           |  33.3%
	// Fish |++++++++                                |  22.2%
	//
	// --- mindmap ---
	// root((Project))
	// `-- Planning
	//     |-- Requirements
	//     |   `-- Design
	//     `-- Development
	//         `-- Frontend
	//             `-- Backend
	//
	// --- timeline ---
	// History
	//
	//   2020  *-- +------------+
	//            | Founded    |
	//            +------------+
	//   2021  *-- +------------+
	//            | Series A   |
	//            +------------+
	//   2022  *-- +------------+
	//            | Launch     |
	//            +------------+
	//
	//         |
	//         +--------------
	//
	// --- gitGraph ---
	// maindevelop
	// *         c0
	// |
	// *         c1
	// |
	// *         c2
	// |/
	// *         m3
	// |
	// *         c4
	//
	// --- journey ---
	//                     My Day
	//
	// Task       | Score | Satisfaction
	// -----------+-------+----------------------
	// Morning    |       |
	// -----------+-------+----------------------
	// Wake up    |  3 😐 | ############........ Me
	// Breakfast  |  5 😊 | #################### Me
	// -----------+-------+----------------------
	//
	// --- quadrantChart ---
	//                       Skills
	//
	//   High |
	//                            |
	//                            |
	//                            |          *
	//                            |
	//                            |
	//                            |
	//                            |
	//                            |
	//                            |      *
	//                            |
	//        --------------------+-------------------
	//                            |
	//                            |
	//                            |
	//                            |
	//                            |
	//                            |
	//                            |
	//                            |
	//                            |
	//    Low +----------------------------------------
	//         Low                                 High
	//
	//   * Go (0.80, 0.90)
	//   * Python (0.70, 0.60)
	//
	// --- xychart-beta ---
	//                          Sales
	//
	//   25 |                                    ############
	//   23 |                                    ############
	//   21 |                                    ############
	//   20 |            ############            ############
	//   18 |            ############            ############
	//   16 |            ############            ############
	//   14 |            ####################################
	//   12 |            ####################################
	//   11 |            ####################################
	//    9 |################################################
	//    7 |################################################
	//    5 |################################################
	//    4 |################################################
	//    2 |################################################
	//    0 |################################################
	//      +------------------------------------------------
	//            Q1          Q2          Q3          Q4
	//
	// --- C4Context ---
	// +--------------+
	// |  <<person>>  |
	// |     User     |
	// |   A person   |
	// +--------------+
	//
	// +-------------------+
	// |    <<system>>     |
	// |        App        |
	// |  The application  |
	// +-------------------+
	//
	// user --> app : Uses
	//
	// --- requirementDiagram ---
	// +---------------------------------------+
	// |            <<requirement>>            |
	// |               test_req                |
	// |                 Id: 1                 |
	// |  Text: the system shall do something  |
	// +---------------------------------------+
	//
	// +--------------------+
	// |    <<element>>     |
	// |    test_entity     |
	// |  Type: simulation  |
	// +--------------------+
	//
	// test_entity --> test_req [satisfies]
	//
	// --- block-beta ---
	// +----------+  +----------+  +----------+
	// |  Alpha   |  |   Beta   |  |  Gamma   |
	// +----------+  +----------+  +----------+
	// +----------+  +----------+  +----------+
	// |  Delta   |  | Epsilon  |  |   Zeta   |
	// +----------+  +----------+  +----------+
	//
	// --- sankey-beta ---
	// Source ######################################## --> Target (25)
	// Source ######################## --> Other (15)
	// Other  ################ --> Final (10)
	//
	// --- packet-beta ---
	//   0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31
	// +-------------------------------+-------------------------------+-------------------------------+-------------------------------+
	// |            Header             |            Payload            |           Checksum            |            Footer             |
	// +-------------------------------+-------------------------------+-------------------------------+-------------------------------+
	//
	// --- kanban ---
	// +--------------------+--------------------+--------------------+
	// |        Todo        |       Doing        |        Done        |
	// +--------------------+--------------------+--------------------+
	// | Task 1             | Task 3             | Task 4             |
	// | Task 2             |                    |                    |
	// +--------------------+--------------------+--------------------+
	//
	// --- architecture-beta ---
	// +-------+
	// |  API  |
	// +-------+
	//
	// +------------+
	// |  Database  |
	// +------------+
	//
	// api --> db
	//
	// --- zenuml ---
	// +-------+     +-----+
	// | Alice |     | Bob |
	// +---+---+     +--+--+
	//     |            |
	//     | (async) hello()
	//     +---         |
	//     |            |
	//     | hi()       |
	//     +----------->|
	//     |            |
}
