package cmd

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

// TestMermaidJSDocsExamples validates that examples from the official Mermaid.js
// documentation (https://mermaid.js.org) can be parsed and rendered.
func TestMermaidJSDocsExamples(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// === Flowchart examples from https://mermaid.js.org/syntax/flowchart.html ===
		{
			name: "flowchart/basic_TD",
			input: `flowchart TD
    A[Start] --> B[Process]
    B --> C{Decision}
    C -->|Yes| D[End]
    C -->|No| B`,
		},
		{
			name: "flowchart/LR_with_shapes",
			input: `flowchart LR
    A([Stadium]) --> B((Circle))
    B --> C{Diamond}
    C --> D>Flag]`,
		},
		{
			name: "flowchart/dotted_and_thick_links",
			input: `flowchart TD
    A-.Dotted link.->B
    B ==> C
    C -->|Text| D`,
		},
		{
			name: "flowchart/subgraph",
			input: `flowchart TD
    subgraph cluster["Subgraph Title"]
        A[Node 1]
        B[Node 2]
    end
    A --> B`,
		},
		{
			name: "flowchart/chained_links",
			input: `flowchart TD
    A --> B --> C --> D`,
		},
		{
			name: "flowchart/node_with_text",
			input: `flowchart LR
    id1["This is text in the box"]
    id2(Round edges)
    id1 --> id2`,
		},

		// === Sequence diagram examples from https://mermaid.js.org/syntax/sequenceDiagram.html ===
		{
			name: "sequence/loop_and_notes",
			input: `sequenceDiagram
    participant Alice
    participant Bob
    Alice->>John: Hello John, how are you?
    loop HealthCheck
        John->>John: Fight against hypochondria
    end
    Note right of John: Rational thoughts!
    John-->>Alice: Great!
    John->>Bob: How about you?
    Bob-->>John: Jolly good!`,
		},
		{
			name: "sequence/actors",
			input: `sequenceDiagram
    actor Alice
    actor Bob
    Alice->>Bob: Hi Bob!
    Bob->>Alice: Hi Alice!`,
		},
		{
			name: "sequence/alt_else",
			input: `sequenceDiagram
    participant User
    participant Server
    User->>Server: Request data
    alt Successful response
        Server->>User: Return data
    else Error
        Server->>User: Return error message
    end`,
		},
		{
			name: "sequence/par_blocks",
			input: `sequenceDiagram
    participant A
    participant B
    par Action 1
        A->>B: Task 1
    and Action 2
        A->>B: Task 2
    end`,
		},
		{
			name: "sequence/opt_block",
			input: `sequenceDiagram
    participant Client
    participant Service
    opt Optional interaction
        Client->>Service: Optional request
    end`,
		},
		{
			name: "sequence/critical_block",
			input: `sequenceDiagram
    participant Client
    participant Server
    critical Establish connection
        Client->>Server: Connect
    option On success
        Server->>Client: Connected
    option On failure
        Server->>Client: Error
    end`,
		},
		{
			name: "sequence/activation",
			input: `sequenceDiagram
    participant A
    participant B
    A->>+B: Activate B
    B->>-A: Deactivate`,
		},
		{
			name: "sequence/notes",
			input: `sequenceDiagram
    participant A
    participant B
    Note left of A: This is a note
    Note right of B: Another note
    Note over A,B: Spanning note`,
		},
		{
			name: "sequence/cross_and_async",
			input: `sequenceDiagram
    participant A
    participant B
    A-xB: Solid line with cross
    A--xB: Dotted line with cross
    A-)B: Async message
    A--)B: Dotted async`,
		},

		// === Class diagram examples from https://mermaid.js.org/syntax/classDiagram.html ===
		{
			name: "class/animal_hierarchy",
			input: `classDiagram
    class Animal
    class Duck
    class Fish
    class Zebra
    Animal <|-- Duck
    Animal <|-- Fish
    Animal <|-- Zebra
    Animal : +int age
    Animal : +String gender
    Animal: +isMammal()
    Animal: +mate()
    class Duck{
      +String beakColor
      +swim()
      +quack()
    }
    class Fish{
      -int sizeInFeet
      -canEat()
    }
    class Zebra{
      +bool is_wild
      +run()
    }`,
		},
		{
			name: "class/bank_account",
			input: `classDiagram
    class BankAccount
    BankAccount : +String owner
    BankAccount : +Decimal balance
    BankAccount : +deposit(amount)
    BankAccount : +withdraw(amount)`,
		},
		{
			name: "class/stereotypes",
			input: `classDiagram
    class Interface1 {
      <<Interface>>
    }
    class Service1 {
      <<Service>>
    }
    class Enum1 {
      <<Enumeration>>
    }`,
		},
		{
			name: "class/visibility",
			input: `classDiagram
    class Person{
      +String name
      #int age
      -String ssn
      ~String address
    }`,
		},

		// === State diagram examples from https://mermaid.js.org/syntax/stateDiagram.html ===
		{
			name: "state/basic_transitions",
			input: `stateDiagram-v2
    [*] --> Still
    Still --> Moving
    Moving --> Still
    Moving --> Crash
    Crash --> [*]`,
		},
		{
			name: "state/composite",
			input: `stateDiagram-v2
    state Composite {
        [*] --> SubState1
        SubState1 --> SubState2
        SubState2 --> [*]
    }`,
		},
		{
			name: "state/nested_composite",
			input: `stateDiagram-v2
    state NestedComposite {
        state Inner {
            [*] --> InnerState
            InnerState --> [*]
        }
    }`,
		},
		{
			name: "state/transition_label",
			input: `stateDiagram-v2
    StateA --> StateB: Transition Label`,
		},

		// === ER diagram examples from https://mermaid.js.org/syntax/entityRelationshipDiagram.html ===
		{
			name: "er/order_example",
			input: `erDiagram
    CUSTOMER ||--o{ ORDER : places
    ORDER ||--|{ LINE-ITEM : contains
    CUSTOMER }|..|{ DELIVERY-ADDRESS : uses`,
		},
		{
			name: "er/with_attributes",
			input: `erDiagram
    CUSTOMER ||--o{ ORDER : places
    CUSTOMER {
        string name
        string custNumber
        string sector
    }
    ORDER ||--|{ LINE-ITEM : contains
    ORDER {
        int orderNumber
        string deliveryAddress
    }
    LINE-ITEM {
        string productCode
        int quantity
        float pricePerUnit
    }`,
		},
		{
			name: "er/named_driver",
			input: `erDiagram
    CAR ||--o{ NAMED-DRIVER : allows
    CAR {
        string registrationNumber
        string make
        string model
    }
    PERSON ||--o{ NAMED-DRIVER : is
    PERSON {
        string firstName
        string lastName
        int age
    }`,
		},
		{
			name: "er/pk_fk_attributes",
			input: `erDiagram
    CAR ||--o{ NAMED-DRIVER : allows
    CAR {
        string registrationNumber PK
        string make
        string model
    }
    PERSON ||--o{ NAMED-DRIVER : is
    PERSON {
        string driversLicense PK
        string firstName
        string lastName
        int age
    }
    NAMED-DRIVER {
        string carRegistrationNumber PK, FK
        string driverLicence PK, FK
    }`,
		},

		// === Gantt examples from https://mermaid.js.org/syntax/gantt.html ===
		{
			name: "gantt/basic",
			input: `gantt
    title A Gantt Diagram
    dateFormat YYYY-MM-DD
    section Section
    A task           :a1, 2014-01-01, 30d
    Another task     :after a1, 20d`,
		},
		{
			name: "gantt/with_milestones",
			input: `gantt
    dateFormat YYYY-MM-DD
    section Milestones
    Kickoff          :crit, milestone, m1, 2014-01-07, 0d
    Phase 1          :p1, 2014-01-07, 30d`,
		},
		{
			name: "gantt/dependencies",
			input: `gantt
    dateFormat YYYY-MM-DD
    section Development
    Design           :des1, 2014-01-06, 5d
    Implementation   :impl1, after des1, 10d
    Testing          :test1, after impl1, 5d`,
		},

		// === Pie chart examples from https://mermaid.js.org/syntax/pie.html ===
		{
			name: "pie/pets_adopted",
			input: `pie title Pets adopted by volunteers
    "Dogs" : 386
    "Cats" : 85
    "Rats" : 15`,
		},
		{
			name: "pie/show_data",
			input: `pie showData
    title Key elements in Product X
    "Calcium" : 42.96
    "Potassium" : 50.05
    "Magnesium" : 10.01
    "Iron" :  5`,
		},

		// === Mindmap examples from https://mermaid.js.org/syntax/mindmap.html ===
		{
			name: "mindmap/comprehensive",
			input: `mindmap
  root((mindmap))
    Origins
      Long history
      Popularisation
        British popular psychology author Tony Buzan
    Research
      On effectiveness and features
      On Automatic creation
        Uses
            Creative techniques
            Strategic planning
            Argument mapping
    Tools
      Pen and paper
      Mermaid`,
		},
		{
			name: "mindmap/shapes",
			input: `mindmap
    root
        id1[I am a square]
        id2(I am a rounded square)
        id3((I am a circle))`,
		},

		// === Timeline examples from https://mermaid.js.org/syntax/timeline.html ===
		{
			name: "timeline/social_media",
			input: `timeline
    title History of Social Media Platform
    2002 : LinkedIn
    2004 : Facebook
         : Google
    2005 : YouTube
    2006 : Twitter`,
		},
		{
			name: "timeline/industrial_revolution",
			input: `timeline
    title Timeline of Industrial Revolution
    section 17th-20th century
        Industry 1.0 : Machinery, Water power, Steam power
        Industry 2.0 : Electricity, Internal combustion engine, Mass production
        Industry 3.0 : Electronics, Computers, Automation
    section 21st century
        Industry 4.0 : Internet, Robotics, Internet of Things
        Industry 5.0 : Artificial intelligence, Big data, 3D printing`,
		},

		// === GitGraph examples from https://mermaid.js.org/syntax/gitgraph.html ===
		{
			name: "gitgraph/basic",
			input: `gitGraph
    commit
    commit
    commit`,
		},
		{
			name: "gitgraph/custom_ids",
			input: `gitGraph
    commit id: "initial"
    commit id: "feature"
    commit id: "bugfix"`,
		},
		{
			name: "gitgraph/commit_types",
			input: `gitGraph
    commit type: NORMAL
    commit type: REVERSE
    commit type: HIGHLIGHT`,
		},
		{
			name: "gitgraph/branching_and_merge",
			input: `gitGraph
    commit
    commit
    branch develop
    commit
    commit
    commit
    checkout main
    merge develop
    commit
    commit`,
		},
		{
			name: "gitgraph/tags",
			input: `gitGraph
    commit tag: "v1.0"
    commit tag: "v1.1"
    commit tag: "v2.0"`,
		},
		{
			name: "gitgraph/cherry_pick",
			input: `gitGraph
    commit id: "initial"
    branch feature
    commit id: "feature1"
    checkout main
    cherry-pick id: "feature1"`,
		},

		// === Journey examples from https://mermaid.js.org/syntax/userJourney.html ===
		{
			name: "journey/working_day",
			input: `journey
    title My working day
    section Go to work
      Make tea: 5: Me
      Go upstairs: 3: Me
      Do work: 1: Me, Cat
    section Go home
      Go downstairs: 5: Me
      Sit down: 5: Me`,
		},

		// === Quadrant chart examples from https://mermaid.js.org/syntax/quadrantChart.html ===
		{
			name: "quadrant/campaigns",
			input: `quadrantChart
    title Reach and engagement of campaigns
    x-axis Low Reach --> High Reach
    y-axis Low Engagement --> High Engagement
    quadrant-1 We should expand
    quadrant-2 Need to promote
    quadrant-3 Re-evaluate
    quadrant-4 May be improved
    Campaign A: [0.3, 0.6]
    Campaign B: [0.45, 0.23]
    Campaign C: [0.57, 0.69]
    Campaign D: [0.78, 0.34]
    Campaign E: [0.40, 0.34]
    Campaign F: [0.35, 0.78]`,
		},
		{
			name: "quadrant/priority_matrix",
			input: `quadrantChart
    x-axis Urgent --> Not Urgent
    y-axis Not Important --> Important
    quadrant-1 Plan
    quadrant-2 Do
    quadrant-3 Delegate
    quadrant-4 Delete`,
		},

		// === XY chart examples from https://mermaid.js.org/syntax/xyChart.html ===
		{
			name: "xychart/sales_revenue",
			input: `xychart-beta
    title "Sales Revenue"
    x-axis [jan, feb, mar, apr, may, jun, jul, aug, sep, oct, nov, dec]
    y-axis "Revenue (in $)" 4000 --> 11000
    bar [5000, 6000, 7500, 8200, 9500, 10500, 11000, 10200, 9200, 8500, 7000, 6000]
    line [5000, 6000, 7500, 8200, 9500, 10500, 11000, 10200, 9200, 8500, 7000, 6000]`,
		},
		{
			name: "xychart/simplest",
			input: `xychart-beta
    line [1.3, 0.6, 2.4, 0.34]`,
		},

		// === C4 examples from https://mermaid.js.org/syntax/c4.html ===
		{
			name: "c4/context_banking",
			input: `C4Context
    title System Context diagram for Internet Banking System
    Person(customerA, "Banking Customer A", "A customer of the bank, with personal bank accounts.")
    System(SystemAA, "Internet Banking System", "Allows customers to view information about their bank accounts, and make payments.")
    System_Ext(SystemC, "E-mail system", "The internal Microsoft Exchange e-mail system.")
    SystemDb_Ext(SystemE, "Mainframe Banking System", "Stores all of the core banking information.")
    Rel(customerA, SystemAA, "Uses")
    Rel(SystemAA, SystemE, "Uses")
    Rel(SystemAA, SystemC, "Sends e-mails", "SMTP")
    Rel(SystemC, customerA, "Sends e-mails to")`,
		},
		{
			name: "c4/container_banking",
			input: `C4Container
    title Container diagram for Internet Banking System
    Person(customer, "Customer", "A customer of the bank")
    Container_Boundary(c1, "Internet Banking") {
        Container(spa, "Single-Page App", "JavaScript, Angular", "Provides banking functionality")
        Container(web_app, "Web Application", "Java, Spring MVC", "Delivers static content")
        ContainerDb(database, "Database", "SQL Database", "Stores user information")
    }
    System_Ext(banking_system, "Mainframe Banking System", "Core banking information")
    Rel(customer, web_app, "Uses", "HTTPS")
    Rel(web_app, spa, "Delivers")
    Rel(spa, database, "Reads/writes")
    Rel(database, banking_system, "Uses")`,
		},
		{
			name: "c4/component",
			input: `C4Component
    title Component diagram for Internet Banking System - API Application
    Container(spa, "Single Page Application", "javascript and angular", "Provides banking functionality.")
    ContainerDb(db, "Database", "Relational Database Schema", "Stores user information.")
    Container_Boundary(api, "API Application") {
        Component(sign, "Sign In Controller", "MVC Rest Controller", "Allows users to sign in")
        Component(accounts, "Accounts Summary Controller", "MVC Rest Controller", "Provides account summaries")
    }
    Rel(spa, sign, "Uses", "JSON/HTTPS")
    Rel(spa, accounts, "Uses", "JSON/HTTPS")
    Rel(sign, db, "Read & write to", "JDBC")`,
		},

		// === Requirement diagram examples from https://mermaid.js.org/syntax/requirementDiagram.html ===
		{
			name: "requirement/basic",
			input: `requirementDiagram
    requirement test_req {
        id: 1
        text: the test text.
        risk: high
        verifymethod: test
    }
    element test_entity {
        type: simulation
    }
    test_entity - satisfies -> test_req`,
		},
		{
			name: "requirement/comprehensive",
			input: `requirementDiagram
    requirement test_req {
        id: 1
        text: the test text.
        risk: high
        verifymethod: test
    }
    functionalRequirement test_req2 {
        id: 1.1
        text: the second test text.
        risk: low
        verifymethod: inspection
    }
    performanceRequirement test_req3 {
        id: 1.2
        text: the third test text.
        risk: medium
        verifymethod: demonstration
    }
    element test_entity {
        type: simulation
    }
    element test_entity2 {
        type: word doc
        docRef: reqs/test_entity
    }
    test_entity - satisfies -> test_req2
    test_req - traces -> test_req2
    test_req - contains -> test_req3`,
		},

		// === Block diagram examples from https://mermaid.js.org/syntax/block.html ===
		// Note: our parser handles one block per line; multi-block-per-line and edge
		// syntax from the docs are not yet supported.
		{
			name: "block/with_labels",
			input: `block-beta
    columns 2
    A["Block A"]
    B["Block B"]
    C["Block C"]
    D["Block D"]`,
		},
		{
			name: "block/nested",
			input: `block-beta
    columns 1
    block:ID
        A
        B["A wide one in the middle"]
        C
    end
    D`,
		},
		{
			name: "block/shapes",
			input: `block-beta
    id1("Rounded")
    id2("Stadium")
    id3("Subroutine")
    id4("Database")
    id5("Circle")`,
		},

		// === Sankey examples from https://mermaid.js.org/syntax/sankey.html ===
		{
			name: "sankey/basic",
			input: `sankey-beta
Electricity grid,Over generation / exports,104.453
Electricity grid,Heating and cooling - homes,113.726
Electricity grid,H2 conversion,27.14`,
		},
		{
			name: "sankey/multi_source",
			input: `sankey-beta
Agricultural waste,Bio-conversion,124.729
Bio-conversion,Liquid,0.597
Bio-conversion,Losses,26.862
Bio-conversion,Solid,280.322
Bio-conversion,Gas,81.144`,
		},

		// === Packet diagram examples from https://mermaid.js.org/syntax/packet.html ===
		{
			name: "packet/tcp",
			input: `packet-beta
0-15: "Source Port"
16-31: "Destination Port"
32-63: "Sequence Number"
64-95: "Acknowledgment Number"
96-99: "Data Offset"
100-105: "Reserved"
106: "URG"
107: "ACK"
108: "PSH"
109: "RST"
110: "SYN"
111: "FIN"
112-127: "Window"
128-143: "Checksum"
144-159: "Urgent Pointer"
160-191: "(Options and Padding)"
192-255: "Data (variable length)"`,
		},

		// === Kanban examples from https://mermaid.js.org/syntax/kanban.html ===
		{
			name: "kanban/complete_board",
			input: `kanban
Todo
    task1[Create Documentation]
    task2[Create Blog about the new diagram]
InProgress
    task3[Create renderer so that it works in all cases]
Done
    task4[Define grammar]
    task5[Define data]`,
		},

		// === Architecture examples from https://mermaid.js.org/syntax/architecture.html ===
		{
			name: "architecture/basic_api",
			input: `architecture-beta
    group api(cloud)[API]
    service db(database)[Database] in api
    service disk1(disk)[Storage] in api
    service disk2(disk)[Storage] in api
    service server(server)[Server] in api
    db:L -- R:server
    disk1:T -- B:server
    disk2:T -- B:db`,
		},
		{
			name: "architecture/junctions",
			input: `architecture-beta
    service left_disk(disk)[Disk]
    service top_disk(disk)[Disk]
    service bottom_disk(disk)[Disk]
    service top_gateway(internet)[Gateway]
    service bottom_gateway(internet)[Gateway]
    junction junctionCenter
    junction junctionRight
    left_disk:R -- L:junctionCenter
    top_disk:B -- T:junctionCenter
    bottom_disk:T -- B:junctionCenter
    junctionCenter:R -- L:junctionRight
    top_gateway:B -- T:junctionRight
    bottom_gateway:T -- B:junctionRight`,
		},

		// === ZenUML examples (adapted from https://mermaid.js.org/syntax/zenuml.html) ===
		// Note: our parser supports Type.method() syntax, not the A->B: message arrow syntax.
		{
			name: "zenuml/sync_message",
			input: `zenuml
Client client
Server server
server.process(data)`,
		},
		{
			name: "zenuml/async_block",
			input: `zenuml
Client client
API api
API.handleRequest() {
    return ok
}`,
		},
		{
			name: "zenuml/nested_calls",
			input: `zenuml
Client client
Server server
Database db
server.handleRequest(data)
db.query(sql)`,
		},
	}

	config := diagram.NewTestConfig(true, "cli")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := RenderDiagram(tt.input, config)
			if err != nil {
				t.Fatalf("RenderDiagram error: %v", err)
			}
			if len(strings.TrimSpace(output)) == 0 {
				t.Error("RenderDiagram produced empty output")
			}
		})
	}
}
