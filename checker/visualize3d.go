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
        
        .drag-handle {
            cursor: move;
            position: absolute;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            color: #00ffc6;
            font-size: 24px;
            pointer-events: all;
            z-index: 100;
            background-color: rgba(10,10,30,0.7);
            border-radius: 50%;
            width: 80px;
            height: 80px;
            text-align: center;
            line-height: 28px;
            user-select: none;
            display: block;
            padding-top: 12px;
            box-shadow: 0 0 20px rgba(0,255,255,0.5);
            border: 2px solid rgba(0,255,255,0.4);
            animation: pulse 2s infinite;
        }
        
        @keyframes pulse {
            0% { transform: translate(-50%, -50%) scale(1); box-shadow: 0 0 15px rgba(0,255,255,0.5); }
            50% { transform: translate(-50%, -50%) scale(1.1); box-shadow: 0 0 25px rgba(0,255,255,0.7); }
            100% { transform: translate(-50%, -50%) scale(1); box-shadow: 0 0 15px rgba(0,255,255,0.5); }
        }
        
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
        <p>LEFT-CLICK: PAN | RIGHT-CLICK: ROTATE | SCROLL: ZOOM | CENTER HANDLE: DRAG BUILDING</p>
    </div>
    
    <div id="dragHandle" class="drag-handle">
        <span style="font-size: 28px;">⬌</span><br>
        <span style="font-size: 12px;">MOVE</span>
    </div>
    
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
        <button id="toggleViolationsOnly">Show Violations Only</button>
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

            // Create camera
            camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 1000);
            camera.position.set(20, 20, 20);
            camera.lookAt(0, 0, 0);

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
            directionalLight.position.set(10, 20, 10);
            directionalLight.castShadow = true;
            scene.add(directionalLight);
            
            // Add colored spotlights for cyberpunk effect
            const spotLight1 = new THREE.SpotLight(0xff00ff, 0.8); // Magenta spotlight
            spotLight1.position.set(-15, 30, -15);
            spotLight1.angle = Math.PI / 6;
            spotLight1.penumbra = 0.3;
            scene.add(spotLight1);
            
            const spotLight2 = new THREE.SpotLight(0x00ffff, 0.8); // Cyan spotlight
            spotLight2.position.set(15, 30, 15);
            spotLight2.angle = Math.PI / 6;
            spotLight2.penumbra = 0.3;
            scene.add(spotLight2);

            // Create building foundation
            createFoundation();

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
        function createFoundation() {
            const foundationGeometry = new THREE.BoxGeometry(
                30, 1, 30
            );
            const foundationMaterial = new THREE.MeshStandardMaterial({ 
                color: 0x2d00f7, // Deep electric blue
                roughness: 0.3,
                metalness: 0.8,
                emissive: 0x100060,
                emissiveIntensity: 0.4
            });
            const foundation = new THREE.Mesh(foundationGeometry, foundationMaterial);
            foundation.position.y = floorBaseY - 1;
            foundation.receiveShadow = true;
            scene.add(foundation);
            
            // Add foundation text
            addFloorLabel("Clean Architecture Building", 0, floorBaseY - 0.5);
        }
        
        // Add text label above each floor
        function addFloorLabel(text, level, y) {
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
            textMesh.position.set(0, y, -12);
            textMesh.rotation.x = -Math.PI / 8; // Tilt slightly for better visibility
            
            // Add to scene
            scene.add(textMesh);
        }

        // Create the building with floors and packages
        function createBuilding() {
            scene.add(floorGroup);
            
            buildingData.floors.forEach((floor, index) => {
                const floorY = floorBaseY + index * normalFloorSpacing;
                
                // Create floor
                const floorGeometry = new THREE.BoxGeometry(20, floorHeight, 20);
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
                floorMesh.position.y = floorY;
                floorMesh.receiveShadow = true;
                floorGroup.add(floorMesh);
                
                // Add floor label
                const labelText = floor.label || "Level " + index;
                addFloorLabel(labelText, index, floorY + floorHeight/2 + 0.5);
                
                // Create packages on this floor
                floor.packages.forEach(pkg => {
                    createPackage(pkg, floorY);
                });
            });
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
        function createPackage(pkg, floorY) {
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
            
            // Position using the pre-calculated positions
            mesh.position.set(
                pkg.x, 
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
            updateDependencyVisibility();
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
        
        // Setup drag functionality for moving the entire scene
        function setupDragFunctionality() {
            const dragHandle = document.getElementById('dragHandle');
            let isDragging = false;
            let previousMousePosition = { x: 0, y: 0 };
            
            // Offset for the entire scene
            floorGroup.userData.offset = { x: 0, y: 0, z: 0 };
            
            // Mouse down event
            dragHandle.addEventListener('mousedown', function(e) {
                isDragging = true;
                previousMousePosition = {
                    x: e.clientX,
                    y: e.clientY
                };
                e.preventDefault();
                
                // Temporarily disable orbit controls while dragging
                controls.enabled = false;
            });
            
            // Mouse move event
            document.addEventListener('mousemove', function(e) {
                if (!isDragging) return;
                
                // Calculate movement delta
                const deltaMove = {
                    x: e.clientX - previousMousePosition.x,
                    y: e.clientY - previousMousePosition.y
                };
                
                // Update position based on mouse movement
                const sensitivity = 0.02; // Adjust sensitivity as needed
                floorGroup.position.x += deltaMove.x * sensitivity;
                floorGroup.position.z += deltaMove.y * sensitivity;
                
                // Store the offset for reset functionality
                floorGroup.userData.offset.x = floorGroup.position.x;
                floorGroup.userData.offset.z = floorGroup.position.z;
                
                // Update previous mouse position
                previousMousePosition = {
                    x: e.clientX,
                    y: e.clientY
                };
                
                e.preventDefault();
            });
            
            // Mouse up event
            document.addEventListener('mouseup', function() {
                if (isDragging) {
                    isDragging = false;
                    // Re-enable orbit controls
                    controls.enabled = true;
                }
            });
            
            // Mouse leave event
            document.addEventListener('mouseleave', function() {
                if (isDragging) {
                    isDragging = false;
                    // Re-enable orbit controls
                    controls.enabled = true;
                }
            });
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
            floorGroup.children.forEach(child => {
                // Floor meshes are BoxGeometry with height = floorHeight
                if (child.geometry && child.geometry.type === 'BoxGeometry' && 
                    child.geometry.parameters.height === floorHeight) {
                    floorMeshes.push({
                        mesh: child,
                        startY: child.position.y
                    });
                }
            });
            
            // Find floor labels
            scene.children.forEach(child => {
                // Floor labels are PlaneGeometry with specific properties
                if (child.geometry && child.geometry.type === 'PlaneGeometry' &&
                    child.position.z === -12) {
                    floorLabels.push({
                        mesh: child,
                        startY: child.position.y
                    });
                }
            });
            
            // Animation function
            function updatePositions() {
                const elapsedTime = Date.now() - startTime;
                const progress = Math.min(elapsedTime / duration, 1);
                // Easing function for smoother animation
                const easedProgress = 1 - Math.pow(1 - progress, 3);
                
                // Update each floor's position and all its packages
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
                    
                    // Move the floor mesh
                    if (index < floorMeshes.length) {
                        const floorMesh = floorMeshes[index];
                        const yOffset = (targetY - (floorBaseY + index * normalFloorSpacing));
                        floorMesh.mesh.position.y = floorMesh.startY + yOffset;
                        
                        // Also move the floor label if found
                        const labelIndex = floorLabels.findIndex(label => 
                            Math.abs(label.startY - (floorMesh.startY + floorHeight/2 + 0.5)) < 0.1);
                        
                        if (labelIndex !== -1) {
                            floorLabels[labelIndex].mesh.position.y = 
                                floorLabels[labelIndex].startY + yOffset;
                        }
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
            camera.position.set(20, 20, 20);
            camera.lookAt(0, 0, 0);
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
