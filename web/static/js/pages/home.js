(function() {
  const form = document.getElementById('calculate-form');
  const surfaceForm = document.getElementById('surface-form');
  const resultsDiv = document.getElementById('results');
  const errorDiv = document.getElementById('error-message');
  
  let chart2d = null;
  
  let lastBaseInput = null; 
  let lastBaseViscosity = null;

  async function checkAuth() {
    try {
      await FlowModelAPI.client.auth.me();
      return true;
    } catch (e) {
      window.location.href = '/login';
      return false;
    }
  }

  async function loadMaterials() {
    try {
      const materials = await FlowModelAPI.client.materials.list();
      const select = document.querySelector('select[name="material_id"]');
      if (select && materials.length) {
        select.innerHTML = materials.map(m => `<option value="${m.id}">${m.name}</option>`).join('');
      }
    } catch (e) {
      console.error('Ошибка загрузки материалов:', e);
    }
  }

  function renderProfileTable(profile) {
    const tbody = document.querySelector('#profile-table tbody');
    if (!tbody) return;
    
    tbody.innerHTML = profile.map(p => `
      <tr>
        <td>${p.x.toFixed(4)}</td>
        <td>${p.temperature.toFixed(1)}</td>
        <td>${p.viscosity.toFixed(1)}</td>
      </tr>
    `).join('');
  }

  function renderSurfaceTable(points) {
    const tbody = document.querySelector('#surface-table tbody');
    if (!tbody) return;
    
    tbody.innerHTML = points.map(p => `
      <tr>
        <td>${p.vu.toFixed(2)}</td>
        <td>${p.tu.toFixed(1)}</td>
        <td>${p.productivity.toFixed(2)}</td>
        <td>${p.viscosity.toFixed(1)}</td>
      </tr>
    `).join('');
  }

  function render2DChart(profile) {
    const canvas = document.getElementById('chart2d');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    canvas.style.width = '100%';
    canvas.style.height = '300px';
    canvas.width = canvas.clientWidth || 800;
    canvas.height = 300;
    
    if (chart2d) chart2d.destroy();

    chart2d = new Chart(ctx, {
      type: 'line',
      data: {
        labels: profile.map(p => p.x.toFixed(2)),
        datasets: [
          { 
            label: 'Температура, °C', 
            data: profile.map(p => p.temperature), 
            borderColor: '#ff3c8e', 
            borderWidth: 3, 
            pointRadius: 0, 
            tension: 0.3, 
            yAxisID: 'y' 
          },
          { 
            label: 'Вязкость, Па·с', 
            data: profile.map(p => p.viscosity), 
            borderColor: '#c8ff00', 
            borderWidth: 3, 
            pointRadius: 0, 
            tension: 0.3, 
            yAxisID: 'y1' 
          }
        ]
      },
      options: {
        responsive: true, 
        maintainAspectRatio: false, 
        animation: false,
        interaction: { mode: 'index', intersect: false },
        plugins: { 
          legend: { labels: { font: { family: 'Courier New' }, usePointStyle: true } }
        },
        scales: {
          x: { title: { display: true, text: 'Координата канала, м' } },
          y: { type: 'linear', position: 'left', title: { display: true, text: 'Температура, °C' } },
          y1: { type: 'linear', position: 'right', title: { display: true, text: 'Вязкость, Па·с' }, grid: { drawOnChartArea: false } }
        }
      }
    });
  }

  function render3DChart(surfaceData, baseInput, baseViscosity) {
    const container = document.getElementById('chart3d');
    container.innerHTML = '';
    
    if (!surfaceData || !surfaceData.points || surfaceData.points.length === 0) {
      container.innerHTML = '<p>Не удалось загрузить данные для 3D графика</p>';
      return;
    }

    const width = container.clientWidth || 800;
    const height = 500;

    const scene = new THREE.Scene();
    scene.background = new THREE.Color('#f2ede8');

    const camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 1000);
    camera.position.set(1.8, 1.5, 2.2);
    camera.lookAt(0.5, 0.5, 0.5);

    const renderer = new THREE.WebGLRenderer({ antialias: true });
    renderer.setSize(width, height);
    renderer.setPixelRatio(1);
    container.appendChild(renderer.domElement);

    const controls = new THREE.OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.05;
    controls.autoRotate = true;
    controls.autoRotateSpeed = 0.8;

    scene.add(new THREE.AmbientLight(0xffffff, 0.6));
    const dirLight = new THREE.DirectionalLight(0xffffff, 0.8);
    dirLight.position.set(1, 2, 1);
    scene.add(dirLight);

    const points = surfaceData.points;
    const vuSet = [...new Set(points.map(p => p.vu))].sort((a,b)=>a-b);
    const tuSet = [...new Set(points.map(p => p.tu))].sort((a,b)=>a-b);
    const vuSteps = vuSet.length;
    const tuSteps = tuSet.length;

    const minVu = vuSet[0];
    const maxVu = vuSet[vuSteps - 1];
    const minTu = tuSet[0];
    const maxTu = tuSet[tuSteps - 1];

    let minEta = Infinity;
    let maxEta = -Infinity;
    points.forEach(p => {
      minEta = Math.min(minEta, p.viscosity);
      maxEta = Math.max(maxEta, p.viscosity);
    });

    // Обновляем визуальную легенду (цветовую шкалу)
    const legendEl = document.getElementById('chart3d-legend');
    if (legendEl) {
      legendEl.style.display = 'flex';
      document.getElementById('legend-max').textContent = maxEta.toFixed(1);
      document.getElementById('legend-min').textContent = minEta.toFixed(1);
    }

    const vertices = [];
    const colors = [];
    const indices = [];

    for (let i = 0; i < vuSteps; i++) {
      for (let j = 0; j < tuSteps; j++) {
        const point = points[i * tuSteps + j];
        const x = (point.vu - minVu) / (maxVu - minVu || 1);
        const y = (point.viscosity - minEta) / (maxEta - minEta || 1);
        const z = (point.tu - minTu) / (maxTu - minTu || 1);
        
        // HSL(0.66, x, y) = Синий. HSL(0, x, y) = Красный.
        const color = new THREE.Color().setHSL(0.66 * (1 - y), 0.85, 0.52);

        vertices.push(x, y, z);
        colors.push(color.r, color.g, color.b);
      }
    }

    for (let i = 0; i < vuSteps - 1; i++) {
      for (let j = 0; j < tuSteps - 1; j++) {
        const a = i * tuSteps + j;
        const b = (i + 1) * tuSteps + j;
        const c = (i + 1) * tuSteps + j + 1;
        const d = i * tuSteps + j + 1;
        indices.push(a, b, d, b, c, d);
      }
    }
    
    const geom = new THREE.BufferGeometry();
    geom.setAttribute('position', new THREE.Float32BufferAttribute(vertices, 3));
    geom.setAttribute('color', new THREE.Float32BufferAttribute(colors, 3));
    geom.setIndex(indices);
    geom.computeVertexNormals();
    
    const mat = new THREE.MeshPhongMaterial({
      vertexColors: true, transparent: true, opacity: 0.85, side: THREE.DoubleSide
    });
    scene.add(new THREE.Mesh(geom, mat));

    const markerX = (baseInput.vu - minVu) / (maxVu - minVu || 1);
    const markerY = (baseViscosity - minEta) / (maxEta - minEta || 1);
    const markerZ = (baseInput.tu - minTu) / (maxTu - minTu || 1);

    if (markerX >= 0 && markerX <= 1 && markerZ >= 0 && markerZ <= 1) {
      const marker = new THREE.Mesh(
        new THREE.SphereGeometry(0.04, 24, 24),
        new THREE.MeshPhongMaterial({ color: 0x000000, emissive: 0x222222 })
      );
      marker.position.set(markerX, markerY, markerZ);
      scene.add(marker);
    }
    
    scene.add(new THREE.AxesHelper(1.5));

    function createLabel(text, color, position, fontSize = 16) {
      const canvas = document.createElement('canvas');
      canvas.width = 160; canvas.height = 48;
      const ctx = canvas.getContext('2d');
      ctx.fillStyle = color; ctx.font = `Bold ${fontSize}px 'Courier New'`;
      ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
      ctx.fillText(text, 80, 24);
      const sprite = new THREE.Sprite(new THREE.SpriteMaterial({ map: new THREE.CanvasTexture(canvas) }));
      sprite.position.copy(position);
      sprite.scale.set(0.5, 0.15, 1);
      scene.add(sprite);
    }

    createLabel('Vu (м/с)', '#3366cc', new THREE.Vector3(1.3, -0.1, -0.1));
    createLabel('Вязкость (Па·с)', '#c8ff00', new THREE.Vector3(-0.1, 1.3, -0.1));
    createLabel('Tu (°C)', '#ff3c8e', new THREE.Vector3(-0.1, -0.1, 1.3));

    [0, 1].forEach(t => createLabel((minVu + t * (maxVu - minVu)).toFixed(2), '#3366cc', new THREE.Vector3(t, -0.12, -0.12), 12));
    [0, 1].forEach(t => createLabel((minEta + t * (maxEta - minEta)).toFixed(0), '#c8ff00', new THREE.Vector3(-0.15, t, -0.15), 12));
    [0, 1].forEach(t => createLabel((minTu + t * (maxTu - minTu)).toFixed(0), '#ff3c8e', new THREE.Vector3(-0.15, -0.15, t), 12));

    function animate() { requestAnimationFrame(animate); controls.update(); renderer.render(scene, camera); }
    animate();

    window.addEventListener('resize', () => {
      renderer.setSize(container.clientWidth, container.clientHeight);
      camera.aspect = container.clientWidth / container.clientHeight;
      camera.updateProjectionMatrix();
    });
  }

  async function loadAndRenderSurface(surfacePayload, baseInput, baseViscosity) {
    try {
      const surfaceResult = await FlowModelAPI.client.calculation.surface(surfacePayload);
      render3DChart(surfaceResult, baseInput, baseViscosity);
      renderSurfaceTable(surfaceResult.points);
    } catch (e) {
      console.error('Ошибка при отрисовке поверхности:', e);
      document.getElementById('chart3d').innerHTML = '<p style="color:red;">Ошибка загрузки данных 3D графика.</p>';
    }
  }

  if (form) {
    checkAuth().then(() => {
      loadMaterials();
    });
    
    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      errorDiv.textContent = '';
      
      const isAuth = await checkAuth();
      if (!isAuth) return;
      
      const formData = new FormData(form);
      const data = {
        w: parseFloat(formData.get('w')), 
        h: parseFloat(formData.get('h')), 
        l: parseFloat(formData.get('l')),
        vu: parseFloat(formData.get('vu')), 
        tu: parseFloat(formData.get('tu')),
        material_id: parseInt(formData.get('material_id')), 
        steps: parseInt(formData.get('steps')) || 100
      };

      try {
        const result = await FlowModelAPI.client.calculation.calculate(data);
        
        lastBaseInput = data;
        lastBaseViscosity = result.viscosity;

        if (result.metrics) {
          document.getElementById('metric-time').textContent = result.metrics.calc_time_ms || 0;
          document.getElementById('metric-memory').textContent = result.metrics.memory_used_bytes || 0;
          document.getElementById('metric-ops').textContent = result.metrics.operations_count || result.metrics.operation_counts || 0;
        }

        document.getElementById('productivity').textContent = result.productivity.toFixed(4) + ' кг/ч';
        document.getElementById('temperature').textContent = result.temperature.toFixed(1);
        document.getElementById('viscosity').textContent = result.viscosity.toFixed(1);
        
        resultsDiv.style.display = 'block';
        renderProfileTable(result.profile);
        render2DChart(result.profile);

        document.getElementById('surf_vu_min').value = Math.max(data.vu * 0.5, 0.01).toFixed(2);
        document.getElementById('surf_vu_max').value = (data.vu * 1.5).toFixed(2);
        document.getElementById('surf_tu_min').value = Math.max(data.tu - 20, 0).toFixed(0);
        document.getElementById('surf_tu_max').value = (data.tu + 20).toFixed(0);

        const surfacePayload = {
          ...data,
          vu_min: parseFloat(document.getElementById('surf_vu_min').value),
          vu_max: parseFloat(document.getElementById('surf_vu_max').value),
          vu_steps: 24,
          tu_min: parseFloat(document.getElementById('surf_tu_min').value),
          tu_max: parseFloat(document.getElementById('surf_tu_max').value),
          tu_steps: 24
        };
        
        await loadAndRenderSurface(surfacePayload, lastBaseInput, lastBaseViscosity);

      } catch (e) {
        errorDiv.textContent = `Ошибка: ${e.message}`;
        if (e instanceof FlowModelAPI.AuthError) window.location.href = '/login';
      }
    });

    if (surfaceForm) {
      surfaceForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        if (!lastBaseInput) return; 

        const vu_min = parseFloat(document.getElementById('surf_vu_min').value);
        const vu_max = parseFloat(document.getElementById('surf_vu_max').value);
        const tu_min = parseFloat(document.getElementById('surf_tu_min').value);
        const tu_max = parseFloat(document.getElementById('surf_tu_max').value);

        if (vu_max <= vu_min || tu_max <= tu_min) {
          alert("ОШИБКА: Значения MAX должны быть строго больше значений MIN");
          return;
        }

        const surfacePayload = {
          ...lastBaseInput,
          vu_min: vu_min,
          vu_max: vu_max,
          vu_steps: 24,
          tu_min: tu_min,
          tu_max: tu_max,
          tu_steps: 24
        };

        await loadAndRenderSurface(surfacePayload, lastBaseInput, lastBaseViscosity);
      });
    }
  }
})();