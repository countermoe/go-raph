package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// Copy the structs from main
type Node struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	VX    float64 `json:"vx"`
	VY    float64 `json:"vy"`
	Type  string  `json:"type"`
	Depth int     `json:"depth"`
}

type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

func analyzeProject(projectPath string) (*Graph, error) {
	graph := &Graph{Nodes: []Node{}, Edges: []Edge{}}
	nodeMap := make(map[string]*Node)
	directModules := make(map[string]bool)

	// Parse go.mod if exists
	modPath := filepath.Join(projectPath, "go.mod")
	var mainModule string
	availableModules := make(map[string]bool)

	if data, err := os.ReadFile(modPath); err == nil {
		if modFile, err := modfile.Parse("go.mod", data, nil); err == nil {
			mainModule = modFile.Module.Mod.Path

			// Add main module
			addNode(graph, nodeMap, mainModule, mainModule, "main", 0)

			// Track available external modules
			for _, req := range modFile.Require {
				availableModules[req.Mod.Path] = true
				directModules[req.Mod.Path] = !req.Indirect
			}
		}
	}

	// Parse Go files
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") || strings.Contains(path, "vendor/") || strings.Contains(path, "test/") {
			return err
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil
		}

		packageName := file.Name.Name
		relPath := strings.TrimPrefix(filepath.Dir(path), projectPath+"/")
		if relPath == "." {
			relPath = packageName
		}

		packageID := "pkg:" + relPath
		addNode(graph, nodeMap, packageID, packageName, "package", 0)

		// Process imports
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)

			// Skip standard library
			if !strings.Contains(importPath, ".") {
				continue
			}

			if strings.HasPrefix(importPath, mainModule) {
				// Internal import
				importID := "import:" + importPath
				addNode(graph, nodeMap, importID, filepath.Base(importPath), "internal", 0)
				addEdge(graph, packageID, importID)
			} else {
				// External import - find the best matching module (longest prefix)
				var rootModule string
				var maxLength int
				for modulePath := range availableModules {
					if strings.HasPrefix(importPath, modulePath) && len(modulePath) > maxLength {
						rootModule = modulePath
						maxLength = len(modulePath)
					}
				}

				if rootModule != "" {
					// Add module node if not exists
					addNode(graph, nodeMap, rootModule, rootModule, "external", 2)

					importID := "import:" + importPath
					addNode(graph, nodeMap, importID, filepath.Base(importPath), "external", 1)
					addEdge(graph, packageID, importID)

					// Connect import to its root module
					addEdge(graph, importID, rootModule)
				}
			}
		}

		return nil
	})

	// Connect main module to direct dependencies, and handle indirect dependency chains
	for modulePath := range availableModules {
		if directModules[modulePath] {
			// Direct dependency - connect to main
			addNode(graph, nodeMap, modulePath, modulePath, "external", 2)
			addEdge(graph, mainModule, modulePath)
		} else {
			// Indirect dependency - find what it depends on
			addNode(graph, nodeMap, modulePath, modulePath, "external", 3)

			// Connect indirect dependencies to their likely parents
			for directModule := range directModules {
				if directModules[directModule] && strings.Contains(modulePath, strings.Split(directModule, "/")[0]) {
					// Likely dependency relationship
					addEdge(graph, directModule, modulePath)
					break
				}
			}
		}
	}

	return graph, err
}

func addNode(graph *Graph, nodeMap map[string]*Node, id, label, nodeType string, depth int) {
	if _, exists := nodeMap[id]; !exists {
		node := Node{
			ID:    id,
			Label: label,
			Type:  nodeType,
			Depth: depth,
		}
		graph.Nodes = append(graph.Nodes, node)
		nodeMap[id] = &graph.Nodes[len(graph.Nodes)-1]
	}
}

func addEdge(graph *Graph, source, target string) {
	graph.Edges = append(graph.Edges, Edge{Source: source, Target: target})
}

func main() {
	fmt.Println("🧪 Testing go-raph dependency analysis...")

	graph, err := analyzeProject("..")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("=== DEPENDENCY GRAPH ANALYSIS ===\n")
	fmt.Printf("Nodes: %d, Edges: %d\n\n", len(graph.Nodes), len(graph.Edges))

	// Check for expected nodes
	nodeTypes := make(map[string][]string)
	for _, node := range graph.Nodes {
		nodeTypes[node.Type] = append(nodeTypes[node.Type], node.ID)
	}

	fmt.Printf("Node Types:\n")
	for nodeType, nodes := range nodeTypes {
		fmt.Printf("  %s (%d): %v\n", nodeType, len(nodes), nodes)
	}

	// Check edge connections
	fmt.Printf("\nEdge Connections:\n")
	edgeMap := make(map[string][]string)
	for _, edge := range graph.Edges {
		edgeMap[edge.Source] = append(edgeMap[edge.Source], edge.Target)
	}

	for source, targets := range edgeMap {
		fmt.Printf("  %s -> %v\n", source, targets)
	}

	// Verify websocket and golang.org/x/net relationship
	hasWebsocket := false
	hasGolangNet := false
	websocketConnections := []string{}

	for _, node := range graph.Nodes {
		if node.ID == "github.com/gorilla/websocket" {
			hasWebsocket = true
		}
		if node.ID == "golang.org/x/net" {
			hasGolangNet = true
		}
	}

	for _, edge := range graph.Edges {
		if edge.Source == "github.com/gorilla/websocket" {
			websocketConnections = append(websocketConnections, edge.Target)
		}
	}

	fmt.Printf("\nDependency Chain Analysis:\n")
	fmt.Printf("  Has websocket: %v\n", hasWebsocket)
	fmt.Printf("  Has golang.org/x/net: %v\n", hasGolangNet)
	fmt.Printf("  Websocket connections: %v\n", websocketConnections)

	// Check that golang.org/x/net is connected to websocket if both exist
	if hasWebsocket && hasGolangNet {
		connected := false
		for _, target := range websocketConnections {
			if target == "golang.org/x/net" {
				connected = true
				break
			}
		}
		if !connected {
			fmt.Printf("  WARNING: golang.org/x/net is not connected to websocket\n")
		} else {
			fmt.Printf("  ✓ golang.org/x/net properly connected to websocket\n")
		}
	}

	// Print full graph JSON
	fmt.Printf("\n%s\n", strings.Repeat("=", 50))
	data, _ := json.MarshalIndent(graph, "", "  ")
	fmt.Printf("GRAPH JSON:\n%s\n", string(data))
}
