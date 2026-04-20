(function() {
  const form = document.getElementById('calculate-form');
  const resultsDiv = document.getElementById('results');
  const errorDiv = document.getElementById('error-message');
  
  let chart2d = null;

  async function loadMaterials() {
    const token = sessionStorage.getItem('flowmodel_token');
    if (!token) return;
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
          { label: 'Температура (°C)', data: profile.map(p => p.temperature), borderColor: '#ff3c8e', borderWidth: 3, pointRadius: 0, tension: 0.3, yAxisID: 'y' },
          { label: 'Вязкость (Па·с)', data: profile.map(p => p.viscosity), borderColor: '#c8ff00', borderWidth: 3, pointRadius: 0, tension: 0.3, yAxisID: 'y1' }
        ]
      },
      options: {
        responsive: true, maintainAspectRatio: false, animation: false,
        plugins: { legend: { labels: { font: { family: 'Courier New' } } } },
        scales: {
          x: { title: { display: true, text: 'Длина (м)' }, grid: { color: '#1a1a1a' } },
          y: { type: 'linear', position: 'left', title: { display: true, text: 'Температура (°C)' }, grid: { color: '#1a1a1a' } },
          y1: { type: 'linear', position: 'right', title: { display: true, text: 'Вязкость (Па·с)' }, grid: { drawOnChartArea: false }, reverse: true }
        }
      }
    });
  }

  function calculateProductivity(w, h, vu, eta0, eta) {
    const safeEta = Number(eta) || 1;
    return ((w * h * vu) / 2) * (eta0 / safeEta);
  }

  function interpolateViscosityByTemperature(profile, temperature) {
    const points = profile
      .filter(p => Number.isFinite(p.temperature) && Number.isFinite(p.viscosity))
      .slice()
      .sort((a, b) => a.temperature - b.temperature);

    if (!points.length) return 1;
    if (temperature <= points[0].temperature) return points[0].viscosity;
    if (temperature >= points[points.length - 1].temperature) return points[points.length - 1].viscosity;

    for (let i = 1; i < points.length; i++) {
      const left = points[i - 1];
      const right = points[i];
      if (temperature <= right.temperature) {
        const span = right.temperature - left.temperature || 1;
        const ratio = (temperature - left.temperature) / span;
        return left.viscosity + ratio * (right.viscosity - left.viscosity);
      }
    }

    return points[points.length - 1].viscosity;
  }

  function render3DChart(profile, input) {
    const container = document.getElementById('chart3d');
    container.innerHTML = '';
    
    const width = container.clientWidth || 800;
    const height = 500;

    const scene = new THREE.Scene();
    scene.background = new THREE.Color('#f2ede8');

    const camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 1000);
    camera.position.set(1.8, 1.8, 2.4);
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
    controls.enableZoom = true;
    controls.target.set(0.5, 0.5, 0.5);

    scene.add(new THREE.AmbientLight(0xffffff, 0.6));
    const dirLight1 = new THREE.DirectionalLight(0xffffff, 0.8);
    dirLight1.position.set(1, 2, 1);
    scene.add(dirLight1);
    const dirLight2 = new THREE.DirectionalLight(0xffffff, 0.4);
    dirLight2.position.set(-1, 1, -1);
    scene.add(dirLight2);

    // Нормализация
    const w = Number(input.w) || 0;
    const h = Number(input.h) || 0;
    const baseVu = Math.max(Number(input.vu) || 0, 0.0001);
    const baseTu = Number(input.tu) || 0;
    const eta0 = Number(profile[0]?.viscosity) || 1;

    const tempValues = profile.map(p => Number(p.temperature)).filter(Number.isFinite);
    const profileMinTu = Math.min(...tempValues, baseTu);
    const profileMaxTu = Math.max(...tempValues, baseTu);
    const minTu = profileMinTu === profileMaxTu ? baseTu - 10 : profileMinTu;
    const maxTu = profileMinTu === profileMaxTu ? baseTu + 10 : profileMaxTu;
    const minVu = Math.max(baseVu * 0.5, 0.0001);
    const maxVu = Math.max(baseVu * 1.5, minVu + 0.0001);

    const vuSteps = 24;
    const tuSteps = 24;
    const grid = [];
    let minQ = Infinity;
    let maxQ = -Infinity;

    for (let i = 0; i < vuSteps; i++) {
      const vu = minVu + (i / (vuSteps - 1)) * (maxVu - minVu);
      const row = [];
      for (let j = 0; j < tuSteps; j++) {
        const tu = minTu + (j / (tuSteps - 1)) * (maxTu - minTu);
        const eta = interpolateViscosityByTemperature(profile, tu);
        const q = calculateProductivity(w, h, vu, eta0, eta);
        row.push({ vu, tu, q });
        minQ = Math.min(minQ, q);
        maxQ = Math.max(maxQ, q);
      }
      grid.push(row);
    }
    
    const vertices = [];
    const colors = [];
    const indices = [];

    for (let i = 0; i < vuSteps; i++) {
      for (let j = 0; j < tuSteps; j++) {
        const point = grid[i][j];
        const x = (point.vu - minVu) / (maxVu - minVu || 1);
        const y = (point.tu - minTu) / (maxTu - minTu || 1);
        const z = (point.q - minQ) / (maxQ - minQ || 1);
        const color = new THREE.Color().setHSL(0.68 - z * 0.55, 0.85, 0.52);
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
      vertexColors: true,
      transparent: true,
      opacity: 0.82,
      side: THREE.DoubleSide
    });
    
    const mesh = new THREE.Mesh(geom, mat);
    scene.add(mesh);

    const currentEta = interpolateViscosityByTemperature(profile, baseTu);
    const currentQ = calculateProductivity(w, h, baseVu, eta0, currentEta);
    const marker = new THREE.Mesh(
      new THREE.SphereGeometry(0.035, 24, 24),
      new THREE.MeshPhongMaterial({ color: 0xff3c8e, emissive: 0x220011 })
    );
    marker.position.set(
      (baseVu - minVu) / (maxVu - minVu || 1),
      (baseTu - minTu) / (maxTu - minTu || 1),
      (currentQ - minQ) / (maxQ - minQ || 1)
    );
    scene.add(marker);

    // Оси
    const axesHelper = new THREE.AxesHelper(1.5);
    scene.add(axesHelper);

    // Функция метки
    function createLabel(text, color, position, fontSize = 18) {
      const canvas = document.createElement('canvas');
      canvas.width = 128;
      canvas.height = 48;
      const ctx = canvas.getContext('2d');
      ctx.fillStyle = color;
      ctx.font = `Bold ${fontSize}px 'Courier New'`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(text, 64, 24);
      
      const texture = new THREE.CanvasTexture(canvas);
      const material = new THREE.SpriteMaterial({ map: texture });
      const sprite = new THREE.Sprite(material);
      sprite.position.copy(position);
      sprite.scale.set(0.4, 0.15, 1);
      scene.add(sprite);
    }

    createLabel('Vu (m/s)', '#3366cc', new THREE.Vector3(1.3, -0.1, -0.1), 18);
    createLabel('Tu (C)', '#ff3c8e', new THREE.Vector3(-0.1, 1.3, -0.1), 18);
    createLabel('Q (m3/s)', '#c8ff00', new THREE.Vector3(-0.1, -0.1, 1.3), 18);

    [0, 0.25, 0.5, 0.75, 1.0].forEach(t => {
      createLabel((minVu + t * (maxVu - minVu)).toFixed(2), '#3366cc', new THREE.Vector3(t, -0.12, -0.12), 12);
    });

    [0, 0.25, 0.5, 0.75, 1.0].forEach(t => {
      createLabel((minTu + t * (maxTu - minTu)).toFixed(0), '#ff3c8e', new THREE.Vector3(-0.15, t, -0.15), 12);
    });

    [0, 0.25, 0.5, 0.75, 1.0].forEach(t => {
      const val = minQ + t * (maxQ - minQ);
      createLabel(val.toExponential(2), '#c8ff00', new THREE.Vector3(-0.15, -0.15, t), 12);
    });

    function animate() {
      requestAnimationFrame(animate);
      controls.update();
      renderer.render(scene, camera);
    }
    animate();

    window.addEventListener('resize', () => {
      const w = container.clientWidth;
      const h = container.clientHeight;
      renderer.setSize(w, h);
      camera.aspect = w / h;
      camera.updateProjectionMatrix();
    });
  }

  if (form) {
    if (sessionStorage.getItem('flowmodel_token')) loadMaterials();
    
    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      errorDiv.textContent = '';
      
      const token = sessionStorage.getItem('flowmodel_token');
      if (!token) { window.location.href = '/login'; return; }
      
      const formData = new FormData(form);
      const data = {
        w: parseFloat(formData.get('w')), h: parseFloat(formData.get('h')), l: parseFloat(formData.get('l')),
        vu: parseFloat(formData.get('vu')), tu: parseFloat(formData.get('tu')),
        material_id: parseInt(formData.get('material_id')), t0: parseFloat(formData.get('t0')), steps: parseInt(formData.get('steps'))
      };

      try {
        const result = await FlowModelAPI.client.calculation.calculate(data);
        const eta0 = Number(result.profile?.[0]?.viscosity) || 1;
        const eta = Number(result.viscosity) || Number(result.profile?.[result.profile.length - 1]?.viscosity) || eta0;
        const productivity = calculateProductivity(data.w, data.h, data.vu, eta0, eta);
        document.getElementById('productivity').textContent = productivity.toFixed(6);
        document.getElementById('temperature').textContent = result.temperature.toFixed(1);
        document.getElementById('viscosity').textContent = result.viscosity.toFixed(1);
        resultsDiv.style.display = 'block';
        render2DChart(result.profile);
        render3DChart(result.profile, data);
      } catch (e) {
        errorDiv.textContent = `Ошибка: ${e.message}`;
        if (e instanceof FlowModelAPI.AuthError) window.location.href = '/login';
      }
    });
  }
})();
