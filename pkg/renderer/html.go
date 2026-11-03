package renderer

import (
	"encoding/json"
	"fmt"
	"html"

	"github.com/alainairom/github-contribution-city/pkg/city"
)

// HTMLRenderer renders a complete, responsive WebGL / Three.js 3D interactive viewer
type HTMLRenderer struct{}

// NewHTMLRenderer creates a new HTMLRenderer
func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{}
}

// RenderInteractive3D generates a self-contained HTML page with a Three.js 3D city scene
func (r *HTMLRenderer) RenderInteractive3D(scene *city.CityScene, themeName string) string {
	theme := GetTheme(themeName)

	sceneJSON, _ := json.Marshal(scene)
	themeJSON, _ := json.Marshal(theme)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s's GitHub Contribution City (3D)</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background: %s;
      color: %s;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      overflow: hidden;
      width: 100vw;
      height: 100vh;
      display: flex;
      flex-direction: column;
    }
    #navbar {
      background: rgba(13, 17, 23, 0.85);
      backdrop-filter: blur(10px);
      padding: 12px 24px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      border-bottom: 1px solid #30363d;
      z-index: 100;
    }
    .brand {
      display: flex;
      align-items: center;
      gap: 12px;
      font-size: 1.15rem;
      font-weight: 700;
      color: %s;
    }
    .brand span {
      background: #238636;
      color: #fff;
      font-size: 0.75rem;
      padding: 2px 8px;
      border-radius: 12px;
    }
    .controls {
      display: flex;
      gap: 12px;
      align-items: center;
    }
    .btn {
      background: #21262d;
      color: #c9d1d9;
      border: 1px solid #30363d;
      padding: 6px 14px;
      border-radius: 6px;
      cursor: pointer;
      font-size: 0.85rem;
      font-weight: 500;
      transition: all 0.2s;
      text-decoration: none;
    }
    .btn:hover { background: #30363d; color: #fff; }
    .btn-primary { background: #238636; color: #fff; border-color: rgba(240, 246, 252, 0.1); }
    .btn-primary:hover { background: #2ea043; }
    #viewport {
      flex: 1;
      position: relative;
      width: 100%%;
      height: 100%%;
    }
    #canvas-container {
      width: 100%%;
      height: 100%%;
    }
    #overlay-stats {
      position: absolute;
      bottom: 24px;
      left: 24px;
      background: rgba(22, 27, 34, 0.85);
      backdrop-filter: blur(8px);
      border: 1px solid #30363d;
      padding: 16px 20px;
      border-radius: 8px;
      pointer-events: auto;
      max-width: 320px;
      box-shadow: 0 8px 24px rgba(0,0,0,0.5);
    }
    .stat-row {
      display: flex;
      justify-content: space-between;
      margin-bottom: 6px;
      font-size: 0.85rem;
    }
    .stat-row .label { color: #8b949e; }
    .stat-row .val { font-weight: 600; color: #f0f6fc; }
    #tooltip {
      position: absolute;
      display: none;
      background: rgba(13, 17, 23, 0.95);
      border: 1px solid #58a6ff;
      color: #c9d1d9;
      padding: 10px 14px;
      border-radius: 6px;
      font-size: 0.8rem;
      pointer-events: none;
      z-index: 1000;
      box-shadow: 0 4px 12px rgba(0,0,0,0.4);
    }
    #legend {
      position: absolute;
      bottom: 24px;
      right: 24px;
      background: rgba(22, 27, 34, 0.85);
      backdrop-filter: blur(8px);
      border: 1px solid #30363d;
      padding: 12px 16px;
      border-radius: 8px;
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 0.75rem;
      color: #8b949e;
    }
    .legend-box {
      width: 12px;
      height: 12px;
      border-radius: 2px;
    }
  </style>
  <!-- Three.js and OrbitControls from CDN (with reliable local fallback) -->
  <script src="https://cdnjs.cloudflare.com/ajax/libs/three.js/r128/three.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/three@0.128.0/examples/js/controls/OrbitControls.js"></script>
</head>
<body>
  <div id="navbar">
    <div class="brand">
      🌆 <span>3D</span> %s's GitHub Contribution City
    </div>
    <div class="controls">
      <a href="/api/svg?user=%s&theme=%s" class="btn" target="_blank">📥 Export SVG Skyline</a>
      <a href="/api/iso?user=%s&theme=%s" class="btn" target="_blank">📐 Export 3D SVG</a>
      <button class="btn" onclick="toggleAutoRotate()">🔄 Rotate: <span id="rot-state">ON</span></button>
      <button class="btn" onclick="resetCamera()">🎯 Reset View</button>
      <button class="btn btn-primary" onclick="exportData()">💾 JSON Data</button>
    </div>
  </div>

  <div id="viewport">
    <div id="canvas-container"></div>
    
    <div id="overlay-stats">
      <h3 style="font-size: 1rem; margin-bottom: 10px; color: #58a6ff;">City Demographics</h3>
      <div class="stat-row"><span class="label">Mayor:</span><span class="val">%s</span></div>
      <div class="stat-row"><span class="label">Total Contributions:</span><span class="val">%d</span></div>
      <div class="stat-row"><span class="label">City Blocks (Weeks):</span><span class="val">%d</span></div>
      <div class="stat-row"><span class="label">Tallest Tower:</span><span class="val">%d contributions</span></div>
      <div class="stat-row"><span class="label">Avg Weekly Activity:</span><span class="val">%.1f</span></div>
      <div class="stat-row"><span class="label">Active Theme:</span><span class="val">%s</span></div>
    </div>

    <div id="legend">
      <span>Less</span>
      <div class="legend-box" style="background: %s;"></div>
      <div class="legend-box" style="background: %s;"></div>
      <div class="legend-box" style="background: %s;"></div>
      <div class="legend-box" style="background: %s;"></div>
      <div class="legend-box" style="background: %s;"></div>
      <span>More</span>
    </div>

    <div id="tooltip"></div>
  </div>

  <script>
    const sceneData = %s;
    const themeData = %s;

    let scene, camera, renderer, controls;
    let buildingMeshes = [];
    let raycaster, mouse;
    let autoRotate = true;

    function init() {
      const container = document.getElementById('canvas-container');
      const width = container.clientWidth;
      const height = container.clientHeight;

      // Three.js Scene
      scene = new THREE.Scene();
      scene.background = new THREE.Color(themeData.background);
      scene.fog = new THREE.FogExp2(themeData.background, 0.015);

      // Camera
      camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 1000);
      camera.position.set(30, 25, 45);

      // Renderer
      renderer = new THREE.WebGLRenderer({ antialias: true });
      renderer.setSize(width, height);
      renderer.setPixelRatio(window.devicePixelRatio);
      renderer.shadowMap.enabled = true;
      renderer.shadowMap.type = THREE.PCFSoftShadowMap;
      container.appendChild(renderer.domElement);

      // Controls
      controls = new THREE.OrbitControls(camera, renderer.domElement);
      controls.enableDamping = true;
      controls.dampingFactor = 0.05;
      controls.maxPolarAngle = Math.PI / 2 - 0.05; // Don't go below ground
      controls.target.set(sceneData.buildings_3d.length * 0.9, 4, 0);

      // Lights
      const ambientLight = new THREE.AmbientLight(0xffffff, 0.4);
      scene.add(ambientLight);

      const dirLight = new THREE.DirectionalLight(0xffffff, 0.8);
      dirLight.position.set(40, 60, 40);
      dirLight.castShadow = true;
      dirLight.shadow.mapSize.width = 2048;
      dirLight.shadow.mapSize.height = 2048;
      dirLight.shadow.camera.near = 0.5;
      dirLight.shadow.camera.far = 200;
      const d = 50;
      dirLight.shadow.camera.left = -d;
      dirLight.shadow.camera.right = d;
      dirLight.shadow.camera.top = d;
      dirLight.shadow.camera.bottom = -d;
      scene.add(dirLight);

      const hemiLight = new THREE.HemisphereLight(0xffffff, 0x222233, 0.3);
      hemiLight.position.set(0, 50, 0);
      scene.add(hemiLight);

      // Ground plane / Avenue
      const totalLen = sceneData.buildings_3d.length * 1.8 + 20;
      const groundGeo = new THREE.PlaneGeometry(totalLen, 30);
      const groundMat = new THREE.MeshStandardMaterial({
        color: themeData.street,
        roughness: 0.8,
        metalness: 0.2
      });
      const ground = new THREE.Mesh(groundGeo, groundMat);
      ground.rotation.x = -Math.PI / 2;
      ground.position.set(totalLen / 2 - 10, 0, 0);
      ground.receiveShadow = true;
      scene.add(ground);

      // Road Grid Lines
      const gridHelper = new THREE.GridHelper(totalLen, Math.floor(totalLen / 2), themeData.street_line, themeData.building_edge);
      gridHelper.position.set(totalLen / 2 - 10, 0.02, 0);
      scene.add(gridHelper);

      // Build 3D City Structures
      buildCity();

      // Interaction
      raycaster = new THREE.Raycaster();
      mouse = new THREE.Vector2();

      window.addEventListener('resize', onWindowResize, false);
      renderer.domElement.addEventListener('mousemove', onMouseMove, false);

      animate();
    }

    function buildCity() {
      const buildings = sceneData.buildings_3d || [];
      const windowColors = themeData.window_levels.map(c => new THREE.Color(c));

      buildings.forEach((b, idx) => {
        const bGroup = new THREE.Group();
        bGroup.position.set(b.grid_x, 0, b.grid_z);

        const bHeight = Math.max(0.8, b.height);
        const geo = new THREE.BoxGeometry(b.width, bHeight, b.depth);

        // Building activity color
        let level = 0;
        if (b.weekly_total > 0) {
          level = 1 + Math.floor((b.weekly_total / (sceneData.max_weekly_total + 1)) * 3);
          if (level > 4) level = 4;
        }

        const mat = new THREE.MeshStandardMaterial({
          color: windowColors[level] || new THREE.Color(themeData.building_body),
          roughness: 0.3,
          metalness: 0.6,
          emissive: windowColors[level],
          emissiveIntensity: level > 0 ? 0.25 : 0.05
        });

        const mesh = new THREE.Mesh(geo, mat);
        mesh.position.y = bHeight / 2;
        mesh.castShadow = true;
        mesh.receiveShadow = true;
        mesh.userData = b;

        bGroup.add(mesh);
        buildingMeshes.push(mesh);

        // Roof Trim
        const roofGeo = new THREE.BoxGeometry(b.width * 1.05, 0.2, b.depth * 1.05);
        const roofMat = new THREE.MeshStandardMaterial({ color: themeData.building_roof, roughness: 0.5 });
        const roofMesh = new THREE.Mesh(roofGeo, roofMat);
        roofMesh.position.y = bHeight + 0.1;
        roofMesh.castShadow = true;
        bGroup.add(roofMesh);

        // Busiest Week Antenna
        if (b.has_antenna) {
          const antGeo = new THREE.CylinderGeometry(0.04, 0.08, 4, 8);
          const antMat = new THREE.MeshStandardMaterial({ color: themeData.antenna, metalness: 0.9, roughness: 0.1 });
          const ant = new THREE.Mesh(antGeo, antMat);
          ant.position.y = bHeight + 2;
          bGroup.add(ant);

          // Glowing beacon light
          const beaconGeo = new THREE.SphereGeometry(0.2, 8, 8);
          const beaconMat = new THREE.MeshBasicMaterial({ color: themeData.antenna });
          const beacon = new THREE.Mesh(beaconGeo, beaconMat);
          beacon.position.y = bHeight + 4;
          bGroup.add(beacon);

          const pointLight = new THREE.PointLight(themeData.antenna, 1.5, 10);
          pointLight.position.y = bHeight + 4.1;
          bGroup.add(pointLight);
        }

        scene.add(bGroup);
      });
    }

    function onMouseMove(event) {
      const rect = renderer.domElement.getBoundingClientRect();
      mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
      mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

      raycaster.setFromCamera(mouse, camera);
      const intersects = raycaster.intersectObjects(buildingMeshes);

      const tooltip = document.getElementById('tooltip');
      if (intersects.length > 0) {
        const b = intersects[0].object.userData;
        tooltip.style.display = 'block';
        tooltip.style.left = (event.clientX + 15) + 'px';
        tooltip.style.top = (event.clientY + 15) + 'px';
        tooltip.innerHTML = '<strong>Week ' + (b.week_index + 1) + ' (' + b.first_day + ')</strong><br/>' +
          '• Contributions: <strong>' + b.weekly_total + '</strong><br/>' +
          '• Building Floors: ' + b.floors + '<br/>' +
          '• District: ' + b.district +
          (b.has_antenna ? '<br/>🌟 <em>Busiest Week Landmark!</em>' : '');
      } else {
        tooltip.style.display = 'none';
      }
    }

    function onWindowResize() {
      const container = document.getElementById('canvas-container');
      camera.aspect = container.clientWidth / container.clientHeight;
      camera.updateProjectionMatrix();
      renderer.setSize(container.clientWidth, container.clientHeight);
    }

    function toggleAutoRotate() {
      autoRotate = !autoRotate;
      document.getElementById('rot-state').textContent = autoRotate ? 'ON' : 'OFF';
    }

    function resetCamera() {
      camera.position.set(30, 25, 45);
      controls.target.set(sceneData.buildings_3d.length * 0.9, 4, 0);
      controls.update();
    }

    function exportData() {
      const blob = new Blob([JSON.stringify(sceneData, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = sceneData.username + '-contribution-city.json';
      a.click();
    }

    function animate() {
      requestAnimationFrame(animate);
      if (autoRotate) {
        controls.autoRotate = true;
        controls.autoRotateSpeed = 0.8;
      } else {
        controls.autoRotate = false;
      }
      controls.update();
      renderer.render(scene, camera);
    }

    window.onload = init;
  </script>
</body>
</html>`,
		html.EscapeString(scene.Username),
		theme.Background,
		theme.TextPrimary,
		theme.TextPrimary,
		html.EscapeString(scene.Username),
		html.EscapeString(scene.Username),
		theme.Name,
		html.EscapeString(scene.Username),
		theme.Name,
		html.EscapeString(scene.Username),
		scene.TotalContributions,
		len(scene.Buildings3D),
		scene.MaxWeeklyTotal,
		scene.AverageWeekly,
		theme.Name,
		theme.WindowLevels[0],
		theme.WindowLevels[1],
		theme.WindowLevels[2],
		theme.WindowLevels[3],
		theme.WindowLevels[4],
		string(sceneJSON),
		string(themeJSON),
	)
}
