package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateDotGraph creates a Graphviz DOT representation of the dependencies
func GenerateDotGraph(packageMap map[string]PackageInfo, packageLevels [][]string) string {
	// Initialize DOT format string
	dotGraph := "digraph DependencyGraph {\n"
	dotGraph += "  // Graph styling\n"
	dotGraph += "  graph [rankdir=BT, fontname=\"Arial\", splines=ortho, ranksep=1.5];\n" // Bottom to top layout to match Clean Architecture
	dotGraph += "  node [shape=box, style=filled, fontname=\"Arial\", fontsize=11];\n"
	dotGraph += "  edge [fontname=\"Arial\", fontsize=9];\n\n"

	// Create subgraphs for each level in reverse order (utility packages at bottom)
	for i := len(packageLevels) - 1; i >= 0; i-- {
		packages := packageLevels[i]

		// The display level (0 is now the deepest level - Enterprise Business Rules)
		displayLevel := len(packageLevels) - i - 1

		// Get layer info if available
		layerInfo := ""
		if displayLevel < len(CleanArchitectureLayers) {
			layerInfo = CleanArchitectureLayers[displayLevel]
		}

		// Start subgraph for this level
		dotGraph += fmt.Sprintf("  // Level %d: %s\n", displayLevel, layerInfo)
		dotGraph += fmt.Sprintf("  subgraph cluster_level_%d {\n", displayLevel)
		dotGraph += fmt.Sprintf("    label=\"Level %d: %s\";\n", displayLevel, layerInfo)
		dotGraph += "    style=filled;\n"

		// Color based on level
		switch displayLevel {
		case 0:
			dotGraph += "    color=\"#f7e6ff\";\n" // Light purple - Enterprise Business Rules
		case 1:
			dotGraph += "    color=\"#fff2e6\";\n" // Light orange - Application Business Rules
		case 2:
			dotGraph += "    color=\"#e6ffe6\";\n" // Light green - Interface Adapters
		case 3:
			dotGraph += "    color=\"#e6f7ff\";\n" // Light blue - Frameworks & Drivers
		default:
			dotGraph += "    color=\"#f2f2f2\";\n" // Light gray
		}

		// Add each package as a node
		for _, pkg := range packages {
			cleanPkg := strings.Trim(pkg, `"`)
			// Create a safe node ID
			nodeID := sanitizeNodeID(cleanPkg)

			// Set node color/style based on package type
			if isEntryPointPackage(pkg) {
				dotGraph += fmt.Sprintf("    %s [label=\"%s\", fillcolor=\"#ffcccc\", tooltip=\"Entry Point\"];\n",
					nodeID, getShortPackageName(cleanPkg))
			} else if isUtilityPackage(pkg) {
				dotGraph += fmt.Sprintf("    %s [label=\"%s\", fillcolor=\"#ffffcc\", tooltip=\"Utility\"];\n",
					nodeID, getShortPackageName(cleanPkg))
			} else {
				dotGraph += fmt.Sprintf("    %s [label=\"%s\", fillcolor=\"white\"];\n",
					nodeID, getShortPackageName(cleanPkg))
			}
		}

		dotGraph += "  }\n\n"
	}

	// Add edges for dependencies
	dotGraph += "  // Dependencies\n"
	for pkg, info := range packageMap {
		fromNode := sanitizeNodeID(strings.Trim(pkg, `"`))

		// Add edges for each import
		for _, imp := range info.Imports {
			toNode := sanitizeNodeID(strings.Trim(imp, `"`))

			// Check if this is a violation (same or higher level import)
			// With reversed levels, we need to use the original logic but adjust the level comparison
			pkgLevel := findPackageLevel(pkg, packageLevels)
			impLevel := findPackageLevel(imp, packageLevels)

			// Using the original violation check - higher numbers mean higher layers
			if pkgLevel <= impLevel && !isEntryPointPackage(pkg) {
				// This is a violation - show in red
				dotGraph += fmt.Sprintf("  %s -> %s [color=\"red\", penwidth=2.0, tooltip=\"Violation\"];\n",
					fromNode, toNode)
			} else {
				// This is a valid dependency - show in blue
				dotGraph += fmt.Sprintf("  %s -> %s [color=\"blue\"];\n",
					fromNode, toNode)
			}
		}
	}

	// Close the graph
	dotGraph += "}\n"

	return dotGraph
}

// GenerateHTMLReport generates an HTML visualization of the dependencies
func GenerateHTMLReport(packageMap map[string]PackageInfo, packageLevels [][]string, outputPath string) error {
	// Generate DOT representation
	dotGraph := GenerateDotGraph(packageMap, packageLevels)

	// Create HTML with embedded visualization using Viz.js
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Uncle Bob - Clean Architecture Report</title>
    <script src="https://unpkg.com/viz.js@2.1.2/viz.js"></script>
    <script src="https://unpkg.com/viz.js@2.1.2/full.render.js"></script>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            background-color: #f8f9fa;
        }
        h1 {
            color: #333;
            border-bottom: 2px solid #ddd;
            padding-bottom: 10px;
        }
        #graph {
            border: 1px solid #ddd;
            background-color: white;
            padding: 20px;
            margin: 20px 0;
            overflow: auto;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .info-box {
            background-color: #e7f3fe;
            border-left: 6px solid #2196F3;
            margin: 10px 0;
            padding: 10px;
        }
        .clean-arch {
            display: flex;
            flex-direction: column;
            margin: 20px 0;
        }
        .layer {
            padding: 15px;
            margin: 5px 0;
            color: #333;
            font-weight: bold;
        }
        .layer-0 {
            background-color: #e6f7ff;
        }
        .layer-1 {
            background-color: #e6ffe6;
        }
        .layer-2 {
            background-color: #fff2e6;
        }
        .layer-3 {
            background-color: #f7e6ff;
        }
    </style>
</head>
<body>
    <h1>Uncle Bob - Clean Architecture Visualization</h1>
    
    <div class="info-box">
        <h3>Clean Architecture Layers</h3>
        <div class="clean-arch">
            <div class="layer layer-3">Level 3: Frameworks & Drivers (outermost layer)</div>
            <div class="layer layer-2">Level 2: Interface Adapters</div>
            <div class="layer layer-1">Level 1: Application Business Rules</div>
            <div class="layer layer-0">Level 0: Enterprise Business Rules (innermost layer)</div>
        </div>
    </div>
    
    <h2>Dependency Graph</h2>
    <p>Blue arrows represent valid dependencies (inward), red arrows represent violations (same level or outward).</p>
    <div id="graph"></div>
    
    <script>
        // Use Viz.js to render the graph
        const viz = new Viz();
        
        // DOT graph definition
        const dot = ` + "`" + dotGraph + "`" + `;
        
        // Render the graph
        viz.renderSVGElement(dot)
            .then(element => {
                document.getElementById('graph').appendChild(element);
            })
            .catch(error => {
                document.getElementById('graph').innerHTML = 
                    '<p>Error rendering graph: ' + error.message + '</p>';
            });
    </script>
</body>
</html>`

	// Create output directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// Write the HTML to file
	return os.WriteFile(outputPath, []byte(html), 0644)
}

// findPackageLevel finds which level a package belongs to
func findPackageLevel(pkg string, packageLevels [][]string) int {
	for lvl, packages := range packageLevels {
		if contains(packages, pkg) {
			return lvl
		}
	}
	return -1
}

// sanitizeNodeID creates a safe ID for DOT format
func sanitizeNodeID(pkg string) string {
	// Replace slashes with underscores
	id := strings.ReplaceAll(pkg, "/", "_")
	// Replace dots with underscores
	id = strings.ReplaceAll(id, ".", "_")
	// Replace dashes with underscores
	id = strings.ReplaceAll(id, "-", "_")
	// Add prefix to ensure it doesn't start with a number
	return "pkg_" + id
}

// getShortPackageName returns just the last part of the package name for display
func getShortPackageName(pkg string) string {
	// Get the last component of the path
	parts := strings.Split(pkg, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return pkg
}
