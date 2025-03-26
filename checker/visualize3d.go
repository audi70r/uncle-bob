package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Generate3DVisualization creates a 3D visualization of the dependency structure as a building
func Generate3DVisualization(packageMap map[string]PackageInfo, packageLevels [][]string, outputPath string) error {
	// Create the data structure for the template
	buildingData := prepareVisualizationData(packageMap, packageLevels)

	// Create output directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// Execute the template
	tmpl, err := template.New("3dvisualization").Parse(threejsTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, buildingData)
}

// BuildingData represents the data needed for the 3D visualization
type BuildingData struct {
	Title         string
	Floors        []Floor
	Dependencies  []Dependency
	FloorLabels   []string
	HasViolations bool
	MaxPackages   int
	TotalPackages int
}

// Floor represents a single level in the building
type Floor struct {
	Level    int
	Label    string
	Packages []Package
}

// Package represents a single package in the visualization
type Package struct {
	ID           string
	Name         string
	ShortName    string
	Level        int
	X            float64
	Z            float64
	Width        float64
	Depth        float64
	Color        string
	IsEntryPoint bool
	IsUtility    bool
}

// Dependency represents a connection between packages
type Dependency struct {
	FromID      string
	ToID        string
	FromLevel   int
	ToLevel     int
	IsViolation bool
}

// prepareVisualizationData converts the package data into a format for the 3D visualization
func prepareVisualizationData(packageMap map[string]PackageInfo, packageLevels [][]string) BuildingData {
	var data BuildingData
	data.Title = "Uncle Bob - Clean Architecture 3D Visualization"

	// Set floor labels based on Clean Architecture layers but in reverse order
	// Utility packages should be at the bottom floors (deeper levels)
	data.FloorLabels = make([]string, len(packageLevels))
	for i := 0; i < len(packageLevels) && i < len(CleanArchitectureLayers); i++ {
		// Reverse the order to match Clean Architecture principles
		// Level 0 (bottom floor) = Enterprise Business Rules (innermost layer)
		// Level N (top floor) = Frameworks & Drivers (outermost layer)
		floorIndex := len(packageLevels) - i - 1
		data.FloorLabels[floorIndex] = CleanArchitectureLayers[i]
	}

	// Count total packages for sizing
	totalPackages := 0
	maxPackagesPerFloor := 0

	for _, packages := range packageLevels {
		count := len(packages)
		totalPackages += count
		if count > maxPackagesPerFloor {
			maxPackagesPerFloor = count
		}
	}

	data.TotalPackages = totalPackages
	data.MaxPackages = maxPackagesPerFloor

	// Create floors with packages - reversed order
	// Highest level (utility packages) should be at the bottom floor (level 0)
	data.Floors = make([]Floor, len(packageLevels))
	packageIDMap := make(map[string]string) // Maps package path to ID

	for i, packages := range packageLevels {
		// Reverse floor index - level 0 is now the bottom floor (highest dependency level)
		reversedFloorIndex := len(packageLevels) - i - 1

		// Get the appropriate label for this floor
		floorLabel := ""
		if reversedFloorIndex < len(packageLevels) && reversedFloorIndex < len(CleanArchitectureLayers) {
			floorLabel = CleanArchitectureLayers[reversedFloorIndex]
		} else {
			// For levels beyond Clean Architecture definitions
			floorLabel = fmt.Sprintf("Level %d", reversedFloorIndex)
		}

		floor := Floor{
			Level:    reversedFloorIndex, // Use reversed index
			Label:    floorLabel,
			Packages: make([]Package, 0, len(packages)),
		}

		// Better package layout with more even distribution
		numPackages := len(packages)

		// Calculate how to arrange packages in a grid
		// We'll use a square-ish grid and determine spacing based on floor size
		gridSize := int(Sqrt(float64(numPackages)))
		if gridSize*gridSize < numPackages {
			gridSize++
		}

		// Calculate spacing to evenly distribute across a 16x16 floor
		// (floor is 20x20 but we leave some margin)
		floorSize := 16.0
		spacingX := floorSize / float64(gridSize)
		spacingZ := floorSize / float64(gridSize)

		// Start position (upper left of the grid, with margins)
		startX := -floorSize/2 + spacingX/2
		startZ := -floorSize/2 + spacingZ/2

		// Layout packages in grid pattern
		for j, pkg := range packages {
			cleanPkg := strings.Trim(pkg, `"`)
			shortName := getShortPackageName(cleanPkg)

			// Generate a unique ID for this package
			id := fmt.Sprintf("pkg_%d_%d", reversedFloorIndex, j)
			packageIDMap[pkg] = id

			// Calculate grid position
			row := j / gridSize
			col := j % gridSize

			// Calculate actual position with even spacing
			posX := startX + float64(col)*spacingX
			posZ := startZ + float64(row)*spacingZ

			// Size based on spacing but with margins
			pkgWidth := spacingX * 0.8
			pkgDepth := spacingZ * 0.8

			// Determine color based on package type
			color := "#ffffff" // white default
			isEntryPoint := isEntryPointPackage(pkg)
			isUtility := isUtilityPackage(pkg)

			if isEntryPoint {
				color = "#ff9999" // light red for entry points
			} else if isUtility {
				color = "#ffffcc" // light yellow for utilities
			}

			// Create the package
			package3d := Package{
				ID:           id,
				Name:         cleanPkg,
				ShortName:    shortName,
				Level:        reversedFloorIndex, // Use reversed index
				X:            posX,
				Z:            posZ,
				Width:        pkgWidth,
				Depth:        pkgDepth,
				Color:        color,
				IsEntryPoint: isEntryPoint,
				IsUtility:    isUtility,
			}

			floor.Packages = append(floor.Packages, package3d)
		}

		data.Floors[reversedFloorIndex] = floor
	}

	// Create dependencies (connecting lines)
	for pkg, info := range packageMap {
		fromID, exists := packageIDMap[pkg]
		if !exists {
			continue
		}

		// Get package level but adjust for reversed order
		fromLevel := findPackageLevel(pkg, packageLevels)
		reversedFromLevel := len(packageLevels) - fromLevel - 1

		for _, imp := range info.Imports {
			toID, exists := packageIDMap[imp]
			if !exists {
				continue
			}

			// Get imported package level but adjust for reversed order
			toLevel := findPackageLevel(imp, packageLevels)
			reversedToLevel := len(packageLevels) - toLevel - 1

			// Check if this is a violation
			// In the reversed visualization, a violation is when a lower floor (higher number)
			// imports from a higher floor (lower number) - opposite of the original rule
			isViolation := false
			if !isEntryPointPackage(pkg) && reversedFromLevel >= reversedToLevel {
				isViolation = true
				data.HasViolations = true
			}

			// Add the dependency
			dependency := Dependency{
				FromID:      fromID,
				ToID:        toID,
				FromLevel:   reversedFromLevel, // Use reversed level
				ToLevel:     reversedToLevel,   // Use reversed level
				IsViolation: isViolation,
			}

			data.Dependencies = append(data.Dependencies, dependency)
		}
	}

	return data
}

// Sqrt is a simple square root function to avoid importing math
func Sqrt(x float64) float64 {
	z := x / 2.0
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

// The template for the 3D visualization HTML file
const threejsTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>U_ARCH :: CLEAN ARCHITECTURE VISUALIZER</title>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/three.js/r128/three.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/three@0.128.0/examples/js/controls/OrbitControls.min.js"></script>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Orbitron:wght@400;700&display=swap');
        
        body { 
            margin: 0;
            overflow: hidden;
            font-family: 'Orbitron', sans-serif;
            background: #0a0a0f;
        }
        
        /* Cyberpunk UI elements */
        #info {
            position: absolute;
            top: 15px;
            width: 100%;
            text-align: center;
            color: #00ffc6;
            text-shadow: 0 0 10px rgba(0, 255, 198, 0.7);
            z-index: 100;
            pointer-events: none;
            letter-spacing: 1px;
        }
        
        #info h1 {
            margin: 0;
            font-size: 28px;
        }
        
        #package-info {
            position: absolute;
            bottom: 10px;
            left: 10px;
            background-color: rgba(10,10,20,0.8);
            color: #d1f7ff;
            padding: 15px;
            border-radius: 3px;
            max-width: 300px;
            font-size: 12px;
            display: none;
            box-shadow: 0 0 15px rgba(0,255,255,0.3), inset 0 0 10px rgba(0,255,255,0.1);
            border: 1px solid rgba(0,255,255,0.3);
        }
        
        #controls {
            position: absolute;
            top: 10px;
            right: 10px;
            background-color: rgba(10,10,30,0.7);
            color: #00ffc6;
            padding: 15px;
            border-radius: 3px;
            max-width: 200px;
            box-shadow: 0 0 15px rgba(0,255,255,0.3), inset 0 0 10px rgba(0,255,255,0.1);
            border: 1px solid rgba(0,255,255,0.3);
        }
        
        /* Removed drag handle styles */
        
        button {
            margin: 5px;
            padding: 7px 12px;
            background: rgba(0,255,198,0.2);
            color: #00ffc6;
            border: 1px solid #00ffc6;
            border-radius: 2px;
            cursor: pointer;
            font-family: 'Orbitron', sans-serif;
            letter-spacing: 1px;
            text-transform: uppercase;
            font-size: 11px;
            transition: all 0.2s ease;
            text-shadow: 0 0 5px rgba(0,255,198,0.7);
            box-shadow: 0 0 10px rgba(0,255,198,0.3);
        }
        
        button:hover {
            background: rgba(0,255,198,0.3);
            box-shadow: 0 0 15px rgba(0,255,198,0.5);
        }
        
        .legend {
            position: absolute;
            top: 10px;
            left: 10px;
            background-color: rgba(10,10,30,0.7);
            color: #d1f7ff;
            padding: 15px;
            border-radius: 3px;
            max-width: 250px;
            font-size: 12px;
            box-shadow: 0 0 15px rgba(0,255,255,0.3), inset 0 0 10px rgba(0,255,255,0.1);
            border: 1px solid rgba(0,255,255,0.3);
        }
        
        .legend h3 {
            margin-top: 0;
            color: #00ffc6;
            letter-spacing: 1px;
        }
        
        .legend-item {
            display: flex;
            align-items: center;
            margin: 7px 0;
        }
        
        .color-box {
            width: 15px;
            height: 15px;
            margin-right: 10px;
            border: 1px solid rgba(209,247,255,0.5);
            box-shadow: 0 0 5px rgba(0,255,255,0.3);
        }
    </style>
</head>
<body>
    <div id="info">
        <h1>U_ARCH :: CLEAN ARCHITECTURE VISUALIZER</h1>
        <p>LEFT-CLICK: PAN | RIGHT-CLICK: ROTATE | SCROLL: ZOOM</p>
    </div>
    
    <!-- Removed drag handle -->
    
    <div class="legend">
        <h3>Legend</h3>
        <div class="legend-item">
            <div class="color-box" style="background-color: #ff003c;"></div>
            <span>Entry Point Package</span>
        </div>
        <div class="legend-item">
            <div class="color-box" style="background-color: #ffea00;"></div>
            <span>Utility Package</span>
        </div>
        <div class="legend-item">
            <div class="color-box" style="background-color: #d1f7ff;"></div>
            <span>Regular Package</span>
        </div>
        <div class="legend-item">
            <div class="color-box" style="background-color: #00ffc6;"></div>
            <span>Valid Dependency (downward)</span>
        </div>
        <div class="legend-item">
            <div class="color-box" style="background-color: #ff003c;"></div>
            <span>Violation (same level or upward)</span>
        </div>
    </div>
    
    <div id="controls">
        <h3>View Controls</h3>
        <button id="toggleDependencies">Show Dependencies</button>
        <button id="toggleViolationsOnly">Show Violations</button>
        <button id="explodeView">Explode View</button>
        <button id="resetView">Reset View</button>
    </div>
    
    <div id="package-info"></div>
    
    <script>
        // Building data from Go template
        const buildingData = {
            floors: [
                {{- range .Floors }}
                {
                    level: {{ .Level }},
                    label: "{{ .Label }}",
                    packages: [
                        {{- range .Packages }}
                        {
                            id: "{{ .ID }}",
                            name: "{{ .Name }}",
                            shortName: "{{ .ShortName }}",
                            level: {{ .Level }},
                            x: {{ .X }},
                            z: {{ .Z }},
                            width: {{ .Width }},
                            depth: {{ .Depth }},
                            color: "{{ .Color }}",
                            isEntryPoint: {{ .IsEntryPoint }},
                            isUtility: {{ .IsUtility }}
                        },
                        {{- end }}
                    ]
                },
                {{- end }}
            ],
            dependencies: [
                {{- range .Dependencies }}
                {
                    fromId: "{{ .FromID }}",
                    toId: "{{ .ToID }}",
                    fromLevel: {{ .FromLevel }},
                    toLevel: {{ .ToLevel }},
                    isViolation: {{ .IsViolation }}
                },
                {{- end }}
            ],
            floorLabels: [{{ range .FloorLabels }}"{{ . }}",{{ end }}],
            hasViolations: {{ .HasViolations }}
        };

        // Three.js setup
        let scene, camera, renderer, controls;
        let packageObjects = {};
        let dependencyLines = [];
        let floorGroup = new THREE.Group();
        let floorHeight = 4;
        let floorBaseY = 0;
        let normalFloorSpacing = floorHeight * 1.5;
        let isExplodedView = false;
        let showDependencies = false; // Hide dependencies by default
        let showViolationsOnly = false;

        // Initialize scene
        function init() {
            // Create scene with cyberpunk dark background
            scene = new THREE.Scene();
            scene.background = new THREE.Color(0x0a0a0f); // Dark blue-black

            // Create camera with top-down perspective slightly tilted back
            camera = new THREE.PerspectiveCamera(65, window.innerWidth / window.innerHeight, 0.1, 1000);
            camera.position.set(0, 70, 40); // Position higher up, panned more toward bottom to see roof and front
            camera.lookAt(0, 0, -10);

            // Create renderer
            renderer = new THREE.WebGLRenderer({ antialias: true });
            renderer.setSize(window.innerWidth, window.innerHeight);
            renderer.shadowMap.enabled = true;
            document.body.appendChild(renderer.domElement);

            // Add orbit controls with reversed mouse buttons
            controls = new THREE.OrbitControls(camera, renderer.domElement);
            controls.enableDamping = true;
            controls.dampingFactor = 0.25;
            // Reverse mouse buttons: left = pan, right = rotate
            controls.mouseButtons = {
                LEFT: THREE.MOUSE.PAN,
                MIDDLE: THREE.MOUSE.DOLLY,
                RIGHT: THREE.MOUSE.ROTATE
            };

            // Add cyberpunk style lighting
            const ambientLight = new THREE.AmbientLight(0x0a0a1a, 0.3); // Dark blue ambient light
            scene.add(ambientLight);

            // Main directional light
            const directionalLight = new THREE.DirectionalLight(0xffffff, 0.8);
            directionalLight.position.set(0, 50, 30);
            directionalLight.castShadow = true;
            scene.add(directionalLight);
            
            // Add colored spotlights for cyberpunk effect
            const spotLight1 = new THREE.SpotLight(0xff00ff, 0.8); // Magenta spotlight for left building
            spotLight1.position.set(-25, 40, 0);
            spotLight1.angle = Math.PI / 4;
            spotLight1.penumbra = 0.3;
            scene.add(spotLight1);
            
            const spotLight2 = new THREE.SpotLight(0x00ffff, 0.8); // Cyan spotlight for right building
            spotLight2.position.set(25, 40, 0);
            spotLight2.angle = Math.PI / 4;
            spotLight2.penumbra = 0.3;
            scene.add(spotLight2);
            
            // Add rim lighting
            const rimLight1 = new THREE.DirectionalLight(0xff3677, 0.6); // Pink rim light
            rimLight1.position.set(-30, 15, -20);
            scene.add(rimLight1);
            
            const rimLight2 = new THREE.DirectionalLight(0x38ecff, 0.6); // Blue rim light
            rimLight2.position.set(30, 15, -20);
            scene.add(rimLight2);

            // Create foundations for both buildings 
            createFoundation(-25); // Left building
            createFoundation(25);  // Right building

            // Create floors with packages
            createBuilding();

            // Add dependencies (connections between packages)
            createDependencies();

            // Add event listeners
            window.addEventListener('resize', onWindowResize);
            
            // Add package highlighting
            setupRaycasting();
            
            // Setup controls
            document.getElementById('toggleDependencies').addEventListener('click', toggleDependencies);
            document.getElementById('toggleViolationsOnly').addEventListener('click', toggleViolationsOnly);
            document.getElementById('explodeView').addEventListener('click', toggleExplodeView);
            document.getElementById('resetView').addEventListener('click', resetView);
            
            // Setup drag functionality - always enabled by default
            setupDragFunctionality();

            // Start animation loop
            animate();
        }

        // Create the building foundation
        function createFoundation(offsetX = 0) {
            // Create a foundation for each building
            const foundationGeometry = new THREE.BoxGeometry(
                35, 1, 35
            );
            const foundationMaterial = new THREE.MeshStandardMaterial({ 
                color: 0x2d00f7, // Deep electric blue
                roughness: 0.3,
                metalness: 0.8,
                emissive: 0x100060,
                emissiveIntensity: 0.4
            });
            const foundation = new THREE.Mesh(foundationGeometry, foundationMaterial);
            foundation.position.set(offsetX, floorBaseY - 1, 0);
            foundation.receiveShadow = true;
            scene.add(foundation);
        }
        
        // Add text label above each floor
        function addFloorLabel(text, level, y, x = 0, z = -12) {
            // Create canvas for the text
            const canvas = document.createElement('canvas');
            const context = canvas.getContext('2d');
            canvas.width = 512;
            canvas.height = 128;
            context.fillStyle = '#ffffff';
            context.font = 'Bold 40px Arial';
            context.textAlign = 'center';
            context.fillText(text, canvas.width / 2, canvas.height / 2);
            
            // Create texture from canvas
            const texture = new THREE.CanvasTexture(canvas);
            texture.needsUpdate = true;
            
            // Create material and geometry
            const material = new THREE.MeshBasicMaterial({
                map: texture,
                transparent: true,
                side: THREE.DoubleSide
            });
            const geometry = new THREE.PlaneGeometry(10, 2.5);
            
            // Create mesh and position it
            const textMesh = new THREE.Mesh(geometry, material);
            textMesh.position.set(x, y, z);
            textMesh.rotation.x = -Math.PI / 8; // Tilt slightly for better visibility
            
            // Add to scene
            scene.add(textMesh);
        }

        // Create the buildings with floors and packages
        function createBuilding() {
            scene.add(floorGroup);
            
            // Create first building (dependency levels)
            const building1OffsetX = -25; // Position first building on the left
            
            buildingData.floors.forEach((floor, index) => {
                const floorY = floorBaseY + index * normalFloorSpacing;
                
                // Create floor
                const floorGeometry = new THREE.BoxGeometry(30, floorHeight, 30);
                const floorMaterial = new THREE.MeshStandardMaterial({ 
                    color: getFloorColor(index),
                    transparent: true,
                    opacity: 0.6,
                    roughness: 0.2,
                    metalness: 0.9,
                    emissive: getFloorColor(index),
                    emissiveIntensity: 0.3
                });
                const floorMesh = new THREE.Mesh(floorGeometry, floorMaterial);
                floorMesh.position.set(building1OffsetX, floorY, 0); // Offset X position
                floorMesh.receiveShadow = true;
                floorGroup.add(floorMesh);
                
                // Add floor label
                const labelText = floor.label || "Level " + index;
                addFloorLabel(labelText, index, floorY + floorHeight/2 + 0.5, building1OffsetX, -12);
                
                // Create packages on this floor with X offset
                floor.packages.forEach(pkg => {
                    createPackage(pkg, floorY, building1OffsetX);
                });
            });
            
            // Create second building (path structure)
            const building2OffsetX = 25; // Position second building on the right
            
            // Organize packages by path depth
            const pathDepthMap = {};
            let maxPathDepth = 0;
            
            // Group packages by path depth
            buildingData.floors.forEach(floor => {
                floor.packages.forEach(pkg => {
                    // Calculate path depth based on the number of "/" in the package path
                    // First, remove any leading protocol or domain parts like "github.com/"
                    let cleanPath = pkg.name;
                    // Count the number of path segments for depth
                    const parts = cleanPath.split('/');
                    // Calculate depth based on number of path segments
                    const pathDepth = Math.max(0, parts.length - 1); // -1 because we don't count the root
                    
                    // Skip depth 0 (root packages)
                    if (pathDepth === 0) return;
                    
                    maxPathDepth = Math.max(maxPathDepth, pathDepth);
                    
                    // Always ensure the array exists for this depth
                    if (!pathDepthMap[pathDepth]) {
                        pathDepthMap[pathDepth] = [];
                    }
                    
                    // Copy the package data with adjusted position
                    pathDepthMap[pathDepth].push({
                        ...pkg,
                        pathDepth: pathDepth
                    });
                });
            });
            
            // Create floors for path structure building with reversed order
            // Deeper paths should be at the bottom, like utilities in the dependency building
            
            // Check if we have any packages at all depths
            for (let i = 1; i <= maxPathDepth; i++) { // Start from 1 to skip root
                if (!pathDepthMap[i]) {
                    pathDepthMap[i] = [];
                }
            }
            
            // Reverse the floors - depth 1 will be at the top, maxDepth at the bottom
            for (let depth = 1; depth <= maxPathDepth; depth++) { // Start from 1 to skip root
                const packages = pathDepthMap[depth] || [];
                if (packages.length === 0) continue; // Skip empty levels
                
                // Calculate reversed position - deeper paths at the bottom
                // This maps depth 0 to the top floor and maxDepth to the bottom floor
                const reversedDepth = maxPathDepth - depth;
                const floorY = floorBaseY + reversedDepth * normalFloorSpacing;
                
                // Create floor
                const floorGeometry = new THREE.BoxGeometry(30, floorHeight, 30);
                const floorMaterial = new THREE.MeshStandardMaterial({ 
                    color: getPathColor(depth, maxPathDepth),
                    transparent: true,
                    opacity: 0.6,
                    roughness: 0.2,
                    metalness: 0.9,
                    emissive: getPathColor(depth, maxPathDepth),
                    emissiveIntensity: 0.3
                });
                const floorMesh = new THREE.Mesh(floorGeometry, floorMaterial);
                floorMesh.position.set(building2OffsetX, floorY, 0); // Offset X position
                floorMesh.receiveShadow = true;
                floorGroup.add(floorMesh);
                
                // Add floor label for path structure (showing reversed)
                // Use the actual depth number, but the position is reversed
                const labelText = "Path Depth " + depth + (depth === maxPathDepth ? " (Deepest)" : "");
                addFloorLabel(labelText, reversedDepth, floorY + floorHeight/2 + 0.5, building2OffsetX, -12);
                
                // Calculate grid layout for packages
                const numPackages = packages.length;
                const gridSize = Math.ceil(Math.sqrt(numPackages));
                const floorSize = 26.0;
                const spacingX = floorSize / gridSize * 1.2;
                const spacingZ = floorSize / gridSize * 1.2;
                const startX = -floorSize/2 + spacingX/2;
                const startZ = -floorSize/2 + spacingZ/2;
                
                // Create packages for this floor
                packages.forEach((pkg, j) => {
                    // Calculate grid position
                    const row = Math.floor(j / gridSize);
                    const col = j % gridSize;
                    
                    // Calculate actual position with even spacing
                    const posX = startX + col * spacingX;
                    const posZ = startZ + row * spacingZ;
                    
                    // Size based on spacing but with margins
                    const pkgWidth = spacingX * 0.8;
                    const pkgDepth = spacingZ * 0.8;
                    
                    // Create package with the right position
                    createPathPackage(pkg, floorY, building2OffsetX, posX, posZ, pkgWidth, pkgDepth);
                });
            }
            
            // Add connecting foundation between buildings at the same level
            const connectionGeometry = new THREE.BoxGeometry(30, 1, 5);
            const connectionMaterial = new THREE.MeshStandardMaterial({ 
                color: 0x2d00f7,
                roughness: 0.3,
                metalness: 0.8,
                emissive: 0x100060,
                emissiveIntensity: 0.2
            });
            const connection = new THREE.Mesh(connectionGeometry, connectionMaterial);
            connection.position.set(0, floorBaseY - 1, 0);
            connection.receiveShadow = true;
            scene.add(connection);
            
            // Add labels for buildings
            addBuildingLabel("Dependency Structure", building1OffsetX, floorBaseY - 3, 0);
            addBuildingLabel("Path Structure", building2OffsetX, floorBaseY - 3, 0);
        }
        
        // Get color for path depth floors
        function getPathColor(depth, maxDepth) {
            // Create a gradient from yellow (shallow) to purple (deep)
            const colors = [
                0xffdd00, // Yellow - depth 0 (shallowest paths)
                0xff9500, // Orange - depth 1
                0xff00aa, // Pink - depth 2
                0x9900ff, // Purple - depth 3
                0x4400ff  // Deep purple - depth 4+ (deepest paths)
            ];
            
            // Use direct index for first few levels, then normalize for deeper levels
            if (depth < colors.length) {
                return colors[depth];
            } else {
                // For deeper paths, use the last color
                return colors[colors.length - 1];
            }
        }
        
        // Add a label for each building
        function addBuildingLabel(text, x, y, z) {
            const canvas = document.createElement('canvas');
            const context = canvas.getContext('2d');
            canvas.width = 512;
            canvas.height = 128;
            context.fillStyle = '#00ffc6';
            context.font = 'Bold 32px Orbitron';
            context.textAlign = 'center';
            context.fillText(text, canvas.width / 2, canvas.height / 2);
            
            const texture = new THREE.CanvasTexture(canvas);
            texture.needsUpdate = true;
            
            const material = new THREE.MeshBasicMaterial({
                map: texture,
                transparent: true,
                side: THREE.DoubleSide
            });
            const geometry = new THREE.PlaneGeometry(12, 3);
            
            const textMesh = new THREE.Mesh(geometry, material);
            textMesh.position.set(x, y, z);
            
            scene.add(textMesh);
        }
        
        // Create a package for the path structure building
        function createPathPackage(pkg, floorY, buildingOffsetX, posX, posZ, width, depth) {
            const height = floorHeight * 0.7;
            const geometry = new THREE.BoxGeometry(width, height, depth);
            
            // Use the same cyberpunk color for the package
            const packageColor = pkg.isEntryPoint ? 0xff003c : (pkg.isUtility ? 0xffea00 : 0xd1f7ff);
            const material = new THREE.MeshStandardMaterial({
                color: packageColor,
                roughness: 0.2,
                metalness: 0.8,
                emissive: packageColor,
                emissiveIntensity: 0.2
            });
            
            const mesh = new THREE.Mesh(geometry, material);
            
            // Position accounting for the building offset
            mesh.position.set(
                buildingOffsetX + posX, 
                floorY + height/2, 
                posZ
            );
            
            mesh.castShadow = true;
            mesh.receiveShadow = true;
            
            // Create a unique ID for this path package
            const pathId = pkg.id + "_path";
            
            // Store mesh reference and package data 
            mesh.userData = {
                id: pathId,
                name: pkg.name,
                shortName: pkg.shortName,
                level: pkg.pathDepth,
                isEntryPoint: pkg.isEntryPoint,
                isUtility: pkg.isUtility
            };
            
            // Store this package in the packageObjects map so we can access it during explosion
            packageObjects[pathId] = {
                mesh: mesh,
                data: pkg,
                y: floorY + height/2
            };
            
            floorGroup.add(mesh);
            
            // Add text label
            const canvas = document.createElement('canvas');
            const context = canvas.getContext('2d');
            canvas.width = 128;
            canvas.height = 64;
            context.fillStyle = '#000000';
            context.font = 'Bold 20px Arial';
            context.textAlign = 'center';
            context.fillText(pkg.shortName, canvas.width / 2, canvas.height / 2);
            
            const texture = new THREE.CanvasTexture(canvas);
            const labelMaterial = new THREE.MeshBasicMaterial({
                map: texture,
                transparent: true,
                side: THREE.DoubleSide
            });
            
            const labelGeometry = new THREE.PlaneGeometry(width * 1.2, width * 0.6);
            const label = new THREE.Mesh(labelGeometry, labelMaterial);
            label.position.y = height/2 + 0.05;
            label.rotation.x = -Math.PI / 2;
            
            mesh.add(label);
        }
        
        // Get color for floor based on level
        function getFloorColor(level) {
            // Cyberpunk color scheme - neon colors on dark backgrounds
            const colors = [
                0x00fff2, // Neon cyan - Level 1
                0xff00ff, // Neon magenta - Level 2
                0x00ff00, // Neon green - Level 3
                0xff2a6d  // Neon pink/red - Level 4
            ];
            
            return level < colors.length ? colors[level] : 0xffcc00; // Default to neon yellow
        }

        // Create individual package (room in the building)
        function createPackage(pkg, floorY, buildingOffsetX = 0) {
            // Create package box
            const height = floorHeight * 0.7;
            const geometry = new THREE.BoxGeometry(pkg.width, height, pkg.depth);
            // Create a cyberpunk style material for the package
            const packageColor = pkg.isEntryPoint ? 0xff003c : (pkg.isUtility ? 0xffea00 : 0xd1f7ff);
            const material = new THREE.MeshStandardMaterial({
                color: packageColor,
                roughness: 0.2,
                metalness: 0.8,
                emissive: packageColor,
                emissiveIntensity: 0.2
            });
            
            const mesh = new THREE.Mesh(geometry, material);
            
            // Position using the pre-calculated positions, applying building offset
            mesh.position.set(
                buildingOffsetX + pkg.x, 
                floorY + height/2, 
                pkg.z
            );
            
            mesh.castShadow = true;
            mesh.receiveShadow = true;
            
            // Store mesh reference and package data
            mesh.userData = {
                id: pkg.id,
                name: pkg.name,
                shortName: pkg.shortName,
                level: pkg.level,
                isEntryPoint: pkg.isEntryPoint,
                isUtility: pkg.isUtility
            };
            
            packageObjects[pkg.id] = {
                mesh: mesh,
                data: pkg,
                y: floorY + height/2
            };
            
            floorGroup.add(mesh);
            
            // Add text label
            const canvas = document.createElement('canvas');
            const context = canvas.getContext('2d');
            canvas.width = 128;
            canvas.height = 64;
            context.fillStyle = '#000000';
            context.font = 'Bold 20px Arial';
            context.textAlign = 'center';
            context.fillText(pkg.shortName, canvas.width / 2, canvas.height / 2);
            
            const texture = new THREE.CanvasTexture(canvas);
            const labelMaterial = new THREE.MeshBasicMaterial({
                map: texture,
                transparent: true,
                side: THREE.DoubleSide
            });
            
            const labelGeometry = new THREE.PlaneGeometry(pkg.width * 1.2, pkg.width * 0.6);
            const label = new THREE.Mesh(labelGeometry, labelMaterial);
            label.position.y = height/2 + 0.05;
            label.rotation.x = -Math.PI / 2;
            
            mesh.add(label);
        }

        // Create dependencies (lines connecting packages)
        function createDependencies() {
            buildingData.dependencies.forEach(dep => {
                const fromPackage = packageObjects[dep.fromId];
                const toPackage = packageObjects[dep.toId];
                
                if (!fromPackage || !toPackage) return;
                
                // Cyberpunk dependency lines
                const lineColor = dep.isViolation ? 0xff003c : 0x00ffc6; // Neon red for violations, neon teal for valid
                const lineWidth = dep.isViolation ? 3 : 1.5;
                
                const fromPoint = fromPackage.mesh.position.clone();
                const toPoint = toPackage.mesh.position.clone();
                
                // Create line
                const material = new THREE.LineBasicMaterial({ 
                    color: lineColor,
                    linewidth: lineWidth
                });
                
                // Add a small curve to make the line more visible
                const midPoint = new THREE.Vector3().addVectors(fromPoint, toPoint).multiplyScalar(0.5);
                midPoint.y += 1.5; // Lift the midpoint to create an arc
                
                // Create a curved path
                const curve = new THREE.QuadraticBezierCurve3(
                    fromPoint,
                    midPoint,
                    toPoint
                );
                
                // Create geometry from the curve
                const points = curve.getPoints(20);
                const geometry = new THREE.BufferGeometry().setFromPoints(points);
                
                const line = new THREE.Line(geometry, material);
                line.userData = {
                    fromId: dep.fromId,
                    toId: dep.toId,
                    isViolation: dep.isViolation
                };
                
                scene.add(line);
                dependencyLines.push(line);
            });
            
            updateDependencyVisibility();
        }

        // Setup raycasting for package selection
        function setupRaycasting() {
            const raycaster = new THREE.Raycaster();
            const mouse = new THREE.Vector2();
            const packageInfoElement = document.getElementById('package-info');
            
            let selectedPackage = null;
            
            window.addEventListener('mousemove', (event) => {
                // Convert mouse position to normalized device coordinates
                mouse.x = (event.clientX / window.innerWidth) * 2 - 1;
                mouse.y = -(event.clientY / window.innerHeight) * 2 + 1;
                
                // Update raycaster
                raycaster.setFromCamera(mouse, camera);
                
                // Find intersections
                const intersects = raycaster.intersectObjects(Object.values(packageObjects).map(p => p.mesh));
                
                if (intersects.length > 0) {
                    // Found a package
                    const packageMesh = intersects[0].object;
                    
                    if (selectedPackage !== packageMesh) {
                        // Reset previous selection
                        if (selectedPackage) {
                            selectedPackage.material.emissive.setHex(0x000000);
                        }
                        
                        // Highlight new selection
                        packageMesh.material.emissive.setHex(0x333333);
                        selectedPackage = packageMesh;
                        
                        // Update info panel
                        const info = packageMesh.userData;
                        let html = "<h3>" + info.name + "</h3>" +
                                   "<p>Level: " + info.level + "</p>";
                        
                        if (info.isEntryPoint) {
                            html += '<p><strong>Entry Point Package</strong></p>';
                        }
                        
                        if (info.isUtility) {
                            html += '<p><strong>Utility Package</strong></p>';
                        }
                        
                        // Add incoming and outgoing dependencies
                        const incoming = [];
                        const outgoing = [];
                        
                        buildingData.dependencies.forEach(dep => {
                            if (dep.fromId === info.id) {
                                const target = packageObjects[dep.toId];
                                if (target) {
                                    outgoing.push({
                                        name: target.data.name,
                                        isViolation: dep.isViolation
                                    });
                                }
                            }
                            
                            if (dep.toId === info.id) {
                                const source = packageObjects[dep.fromId];
                                if (source) {
                                    incoming.push({
                                        name: source.data.name,
                                        isViolation: dep.isViolation
                                    });
                                }
                            }
                        });
                        
                        if (outgoing.length > 0) {
                            html += '<p><strong>Imports:</strong></p><ul>';
                            outgoing.forEach(dep => {
                                const style = dep.isViolation ? 'color: red;' : '';
                                html += "<li style=\"" + style + "\">" + dep.name + "</li>";
                            });
                            html += '</ul>';
                        }
                        
                        if (incoming.length > 0) {
                            html += '<p><strong>Imported by:</strong></p><ul>';
                            incoming.forEach(dep => {
                                const style = dep.isViolation ? 'color: red;' : '';
                                html += "<li style=\"" + style + "\">" + dep.name + "</li>";
                            });
                            html += '</ul>';
                        }
                        
                        packageInfoElement.innerHTML = html;
                        packageInfoElement.style.display = 'block';
                        
                        // Highlight connected dependencies
                        dependencyLines.forEach(line => {
                            if (line.userData.fromId === info.id || line.userData.toId === info.id) {
                                line.material.color.setHex(0xffff00); // Yellow for highlighted connections
                                line.material.linewidth = 2;
                            } else {
                                const color = line.userData.isViolation ? 0xff0000 : 0x0000ff;
                                line.material.color.setHex(color);
                                line.material.linewidth = line.userData.isViolation ? 2 : 1;
                            }
                        });
                    }
                } else {
                    // Reset selection
                    if (selectedPackage) {
                        selectedPackage.material.emissive.setHex(0x000000);
                        selectedPackage = null;
                        
                        // Reset dependency colors
                        dependencyLines.forEach(line => {
                            const color = line.userData.isViolation ? 0xff0000 : 0x0000ff;
                            line.material.color.setHex(color);
                            line.material.linewidth = line.userData.isViolation ? 2 : 1;
                        });
                        
                        packageInfoElement.style.display = 'none';
                    }
                }
            });
        }

        // Toggle dependency visibility
        function toggleDependencies() {
            showDependencies = !showDependencies;
            updateDependencyVisibility();
            
            // Update button text
            const button = document.getElementById('toggleDependencies');
            button.textContent = showDependencies ? 'Hide Dependencies' : 'Show Dependencies';
        }
        
        // Toggle showing only violations
        function toggleViolationsOnly() {
            showViolationsOnly = !showViolationsOnly;
            
            // If enabling violations mode, also enable dependencies if they're not already enabled
            if (showViolationsOnly && !showDependencies) {
                showDependencies = true;
                // Update dependency button text
                const depButton = document.getElementById('toggleDependencies');
                depButton.textContent = 'Hide Dependencies';
            }
            
            updateDependencyVisibility();
            
            // Update button text
            const button = document.getElementById('toggleViolationsOnly');
            button.textContent = showViolationsOnly ? 'Show All Dependencies' : 'Show Violations';
        }
        
        // Update which dependencies are visible based on current settings
        function updateDependencyVisibility() {
            dependencyLines.forEach(line => {
                if (!showDependencies) {
                    line.visible = false;
                } else if (showViolationsOnly) {
                    line.visible = line.userData.isViolation;
                } else {
                    line.visible = true;
                }
            });
        }
        
        // Setup functionality for scene manipulation
        function setupDragFunctionality() {
            // Just initialize the offset object for the floor group
            // No drag functionality needed since we're using orbit controls
            floorGroup.userData.offset = { x: 0, y: 0, z: 0 };
        }
        
        // Toggle exploded view
        function toggleExplodeView() {
            isExplodedView = !isExplodedView;
            
            const targetSpacing = isExplodedView ? normalFloorSpacing * 2.5 : normalFloorSpacing;
            
            // Animate the floor positions
            const duration = 1000; // ms
            const startTime = Date.now();
            
            // Store starting positions for all objects
            const startPositions = {};
            Object.keys(packageObjects).forEach(id => {
                startPositions[id] = {
                    y: packageObjects[id].mesh.position.y
                };
            });
            
            // Find all floor meshes and store their starting positions
            const floorMeshes = [];
            const floorLabels = [];
            
            // Get all meshes that need to be exploded
            floorGroup.children.forEach(child => {
                // Floor meshes are BoxGeometry with height = floorHeight
                if (child.geometry && child.geometry.type === 'BoxGeometry' && 
                    child.geometry.parameters.height === floorHeight) {
                    // Store the floor mesh with its x position for identification
                    floorMeshes.push({
                        mesh: child,
                        startY: child.position.y,
                        x: child.position.x, // Store X position to identify which building
                        floorIndex: Math.round((child.position.y - floorBaseY) / normalFloorSpacing) // Calculate floor index
                    });
                }
            });
            
            // Find floor labels
            scene.children.forEach(child => {
                // Floor labels are PlaneGeometry with specific properties
                if (child.geometry && child.geometry.type === 'PlaneGeometry') {
                    floorLabels.push({
                        mesh: child,
                        startY: child.position.y,
                        x: child.position.x, // Store X position to identify which building
                        z: child.position.z
                    });
                }
            });
            
            // Animation function
            function updatePositions() {
                const elapsedTime = Date.now() - startTime;
                const progress = Math.min(elapsedTime / duration, 1);
                // Easing function for smoother animation
                const easedProgress = 1 - Math.pow(1 - progress, 3);
                
                // Update positions for left building (dependency levels)
                buildingData.floors.forEach((floor, index) => {
                    const currentSpacing = normalFloorSpacing + (targetSpacing - normalFloorSpacing) * easedProgress;
                    const targetY = floorBaseY + index * currentSpacing;
                    
                    // Move packages on this floor
                    floor.packages.forEach(pkg => {
                        const packageObj = packageObjects[pkg.id];
                        if (packageObj) {
                            // Calculate how much to move based on floor index
                            const yOffset = (targetY - (floorBaseY + index * normalFloorSpacing));
                            packageObj.mesh.position.y = startPositions[pkg.id].y + yOffset;
                        }
                    });
                    
                    // Move the dependency building floor mesh
                    const buildingOffset = -25; // Left building offset
                    const leftBuildingFloors = floorMeshes.filter(floor => 
                        Math.abs(floor.x - buildingOffset) < 5 && floor.floorIndex === index);
                    
                    if (leftBuildingFloors.length > 0) {
                        const floorMesh = leftBuildingFloors[0];
                        const yOffset = (targetY - (floorBaseY + index * normalFloorSpacing));
                        floorMesh.mesh.position.y = floorMesh.startY + yOffset;
                        
                        // Also move the floor label if found
                        const labelIndex = floorLabels.findIndex(label => 
                            Math.abs(label.x - buildingOffset) < 5 && 
                            Math.abs(label.startY - (floorMesh.startY + floorHeight/2 + 0.5)) < 0.1 && 
                            label.z < 0);
                        
                        if (labelIndex !== -1) {
                            floorLabels[labelIndex].mesh.position.y = 
                                floorLabels[labelIndex].startY + yOffset;
                        }
                    }
                });
                
                // Update positions for right building (path depths)
                // The right building is now using the reversedDepth, so we need
                // to handle the explode view differently to maintain the correct ordering
                
                // Get all floor meshes of the right building
                const buildingOffset = 25; // Right building offset
                const rightBuildingFloors = floorMeshes.filter(floor => 
                    Math.abs(floor.x - buildingOffset) < 5);
                
                // Sort by current position, bottom to top
                rightBuildingFloors.sort((a, b) => a.startY - b.startY);
                
                // Apply the explode to each floor, maintaining the current order
                rightBuildingFloors.forEach((floorMesh, index) => {
                    const currentSpacing = normalFloorSpacing + (targetSpacing - normalFloorSpacing) * easedProgress;
                    // Calculate target Y position based on current index in the sorted array
                    const targetY = floorBaseY + index * currentSpacing;
                    
                    // Calculate Y offset from current to target position
                    const origY = floorBaseY + index * normalFloorSpacing;
                    const yOffset = targetY - origY;
                    
                    // Move the floor
                    floorMesh.mesh.position.y = floorMesh.startY + yOffset;
                    
                    // Find all packages on this floor
                    Object.keys(packageObjects).forEach(id => {
                        const packageObj = packageObjects[id];
                        // Check if this package is on the right building (x > 0) and on this floor
                        // Note that we're comparing to the actual floor mesh position to find matching packages
                        if (packageObj && packageObj.mesh.position.x > 0) {
                            // If the package Y position is close to this floor's original position, move it
                            if (Math.abs(startPositions[id].y - floorMesh.startY - floorHeight * 0.35) < floorHeight/2) {
                                packageObj.mesh.position.y = startPositions[id].y + yOffset;
                            }
                        }
                    });
                    
                    // Also move the floor label if found
                    const labelIndex = floorLabels.findIndex(label => 
                        Math.abs(label.x - buildingOffset) < 5 && 
                        Math.abs(label.startY - (floorMesh.startY + floorHeight/2 + 0.5)) < 0.1 &&
                        label.z < 0);
                    
                    if (labelIndex !== -1) {
                        floorLabels[labelIndex].mesh.position.y = 
                            floorLabels[labelIndex].startY + yOffset;
                    }
                });
                
                // Update dependency lines
                updateDependencyLines();
                
                if (progress < 1) {
                    requestAnimationFrame(updatePositions);
                }
            }
            
            updatePositions();
        }
        
        // Update dependency lines when package positions change
        function updateDependencyLines() {
            dependencyLines.forEach(line => {
                scene.remove(line);
            });
            
            dependencyLines = [];
            createDependencies();
        }
        
        // Reset view
        function resetView() {
            camera.position.set(0, 70, 40); // Reset to top-down perspective with slight tilt, panned to bottom
            camera.lookAt(0, 0, -10);
            controls.reset();
            
            // Reset floor group position
            floorGroup.position.x = 0;
            floorGroup.position.y = 0;
            floorGroup.position.z = 0;
            floorGroup.userData.offset = { x: 0, y: 0, z: 0 };
            
            // Reset to normal view
            if (isExplodedView) {
                toggleExplodeView();
            }
            
            // Reset dependency visibility
            showDependencies = false;
            showViolationsOnly = false;
            updateDependencyVisibility();
            
            // Update button text
            const depButton = document.getElementById('toggleDependencies');
            depButton.textContent = 'Show Dependencies';
        }

        // Handle window resize
        function onWindowResize() {
            camera.aspect = window.innerWidth / window.innerHeight;
            camera.updateProjectionMatrix();
            renderer.setSize(window.innerWidth, window.innerHeight);
        }

        // Animation loop
        function animate() {
            requestAnimationFrame(animate);
            controls.update();
            renderer.render(scene, camera);
        }

        // Start the visualization
        init();
    </script>
</body>
</html>`
