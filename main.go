package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"golang.org/x/mod/modfile"
)

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

var (
	upgrader   = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	targetPath string
)

func main() {
	flag.StringVar(&targetPath, "path", ".", "Path to analyze")
	port := flag.String("port", "8080", "Server port")
	flag.Parse()

	// Process positional arguments (overrides flags)
	args := flag.Args()
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Validate and fix empty path
	if targetPath == "" {
		targetPath = "."
		fmt.Println("⚠️ Empty path provided, defaulting to current directory")
	}

	// Validate port
	if portNum, err := strconv.Atoi(*port); err != nil || portNum < 1 || portNum > 65535 {
		fmt.Printf("⚠️ Invalid port '%s', defaulting to 8084\n", *port)
		*port = "8084"
	}

	// Check if target path exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		fmt.Printf("❌ Path '%s' does not exist\n", targetPath)
		os.Exit(1)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/ws", websocketHandler)

	fmt.Printf("🎨 Analyzing: %s\n", targetPath)
	fmt.Printf("🌐 Visualizer: http://localhost:%s\n", *port)

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Send initial graph on connection
	graph, err := analyzeProject(targetPath)
	if err != nil {
		conn.WriteJSON(map[string]interface{}{"error": err.Error()})
		return
	}

	conn.WriteJSON(map[string]interface{}{"graph": graph})

	// Keep connection alive
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func analyzeProject(projectPath string) (*Graph, error) {
	graph := &Graph{Nodes: []Node{}, Edges: []Edge{}}
	nodeMap := make(map[string]*Node)
	moduleToImporter := make(map[string][]string) // track which packages import each module
	directModules := make(map[string]bool)        // track direct vs indirect modules
	usedModules := make(map[string]bool)          // track modules that are actually imported

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
				// Don't add nodes yet - let imports drive the connections
			}
		}
	}

	// Parse Go files
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		// Skip vendor folder unless explicitly included
		skipVendor := strings.Contains(path, "vendor/")
		if err != nil || !strings.HasSuffix(path, ".go") || skipVendor {
			return err
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil
		}

		relPath := strings.TrimPrefix(filepath.Dir(path), projectPath+"/")
		if relPath == "." || relPath == "" {
			relPath = "root"
		}

		packageID := "pkg:" + relPath
		// Use directory name for label to avoid confusion with main module
		displayName := filepath.Base(relPath)
		if displayName == "root" {
			displayName = "main"
		}
		addNode(graph, nodeMap, packageID, displayName, "package", 0)

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
					// Mark this module as actually used
					usedModules[rootModule] = true

					// Add module node if not exists
					addNode(graph, nodeMap, rootModule, rootModule, "external", 2)

					// Track that this package imports this module
					moduleToImporter[rootModule] = append(moduleToImporter[rootModule], packageID)

					// If import path exactly matches the module root, connect directly to module
					if importPath == rootModule {
						addEdge(graph, packageID, rootModule)
					} else {
						// Create separate import node for sub-packages
						importID := "import:" + importPath
						// Use full import path for external dependencies, not just base name
						importLabel := importPath
						// If it's too long, show module + last part
						if len(importPath) > 40 {
							parts := strings.Split(importPath, "/")
							if len(parts) > 2 {
								// Show first part (module) + last part
								importLabel = parts[0] + "/.../" + parts[len(parts)-1]
							}
						}
						addNode(graph, nodeMap, importID, importLabel, "external", 1)
						addEdge(graph, packageID, importID)

						// Connect import to its root module
						addEdge(graph, importID, rootModule)
					}
				}
			}
		}

		return nil
	})

	// ONLY connect modules that are actually used in imports
	for modulePath := range usedModules {
		if directModules[modulePath] {
			// Direct dependency that's actually imported - connect to main
			addEdge(graph, mainModule, modulePath)
		} else {
			// Indirect dependency that's actually imported - find best parent
			connected := false
			for directModule := range directModules {
				if directModules[directModule] && usedModules[directModule] {
					// Check if this indirect module is likely a sub-dependency
					if strings.HasPrefix(modulePath, strings.Split(directModule, "/")[0]) ||
						strings.Contains(modulePath, strings.Split(directModule, "/")[0]) {
						addEdge(graph, directModule, modulePath)
						connected = true
						break
					}
				}
			}
			// If we can't find a good parent, don't connect it to avoid orphans
			if !connected {
				// Remove the orphaned module to avoid yellow dots
				for i, node := range graph.Nodes {
					if node.ID == modulePath {
						// Remove node
						graph.Nodes = append(graph.Nodes[:i], graph.Nodes[i+1:]...)
						delete(nodeMap, modulePath)
						break
					}
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
	// Prevent duplicate edges
	for _, edge := range graph.Edges {
		if edge.Source == source && edge.Target == target {
			return
		}
	}
	graph.Edges = append(graph.Edges, Edge{Source: source, Target: target})
}
