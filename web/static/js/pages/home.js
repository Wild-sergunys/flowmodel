(function() {
  const form = document.getElementById('calculate-form');
  const surfaceForm = document.getElementById('surface-form');
  const resultsDiv = document.getElementById('results');
  const errorDiv = document.getElementById('error-message');
  
  let chart2d = null;
  let lastBaseInput = null; 
  let lastBaseViscosity = null;

  const paramLabels = {
    'density': 'Плотность (кг/м³)',
    'heat_capacity': 'Теплоёмкость (Дж/(кг·°С))',
    'melting_temp': 'Темп. плавления (°С)',
    'mu0': 'Коэф. консистенции μ0 (Па·сⁿ)',
    'Ea': 'Энергия активации (Дж/моль)',
    'Tr': 'Темп. приведения (°С)',
    'n': 'Индекс течения',
    'alpha_u': 'Коэф. теплоотдачи (Вт/(м²·°С))'
  };

  async function checkAuth() {
    try {
      await FlowModelAPI.client.auth.me();
      return true;
    } catch (e) {
      window.location.href = '/login';
      return false;
    }
  }

  async function loadMaterialParameters(materialId) {
    // Создаём блок информации, если его нет
    let infoDiv = document.getElementById('material-info');
    if (!infoDiv) {
      const card = document.querySelector('.card');
      if (card) {
        infoDiv = document.createElement('div');
        infoDiv.id = 'material-info';
        infoDiv.style.cssText = 'margin-top: 20px; padding: 12px; background: var(--card-bg); border-radius: 8px; border-left: 3px solid var(--pink);';
        card.appendChild(infoDiv);
      }
    }
    
    if (!infoDiv) return;

    infoDiv.style.display = 'block';
    infoDiv.innerHTML = '<span style="color: var(--muted); font-size: 0.9em;">Загрузка данных материала...</span>';

    try {
      const [materialData, paramsData] = await Promise.all([
        FlowModelAPI.client.materials.get(materialId),
        FlowModelAPI.client.materials.parameters.list(materialId)
      ]);
      
      let html = `<div style="margin-bottom: 12px;">
        <strong>📦 Материал:</strong> ${escapeHtml(materialData.name)}<br>
        <span style="font-size: 0.85em; color: var(--muted);">${escapeHtml(materialData.description || 'Описание отсутствует')}</span>
      </div>`;

      if (paramsData && Object.keys(paramsData).length > 0) {
        const paramsHtml = Object.entries(paramsData).map(([code, value]) => {
          const label = paramLabels[code] || code;
          let formattedValue = value;
          if (typeof value === 'number') {
            if (code === 'n') formattedValue = value.toFixed(3);
            else if (code === 'Ea') formattedValue = value.toFixed(0);
            else formattedValue = value.toFixed(2);
          }
          return `<div style="display: flex; justify-content: space-between; padding: 4px 0; border-bottom: 1px dashed #eee;">
            <span style="font-weight: 500;">${escapeHtml(label)}:</span>
            <span style="font-family: monospace;">${formattedValue} ${escapeHtml(getUnit(code))}</span>
          </div>`;
        }).join('');

        html += `
          <div style="margin-top: 12px; padding-top: 8px; border-top: 1px solid #ddd;">
            <strong>📊 РЕОЛОГИЧЕСКИЕ ПАРАМЕТРЫ</strong>
            <div style="margin-top: 8px; font-size: 0.85em;">
              ${paramsHtml}
            </div>
          </div>
        `;
      } else {
        html += '<span style="color: var(--pink); font-size: 0.85em;">⚠️ Параметры материала не заданы. Используются значения по умолчанию.</span>';
      }

      infoDiv.innerHTML = html;

    } catch (e) {
      console.error('Ошибка загрузки параметров:', e);
      infoDiv.innerHTML = `<span style="color: var(--pink); font-size: 0.85em;">❌ Ошибка загрузки характеристик: ${escapeHtml(e.message)}</span>`;
    }
  }
  
  function getUnit(code) {
    const units = {
      'density': 'кг/м³',
      'heat_capacity': 'Дж/(кг·°С)',
      'melting_temp': '°С',
      'mu0': 'Па·сⁿ',
      'Ea': 'Дж/моль',
      'Tr': '°С',
      'n': '',
      'alpha_u': 'Вт/(м²·°С)'
    };
    return units[code] || '';
  }
  
  function escapeHtml(str) {
    if (!str) return '';
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  async function loadMaterials() {
    try {
      const materials = await FlowModelAPI.client.materials.list();
      const materialSelect = document.querySelector('select[name="material_id"]');
      
      if (materialSelect && materials.length) {
        materialSelect.innerHTML = materials.map(m => `<option value="${m.id}">${escapeHtml(m.name)}</option>`).join('');
        
        // Убираем старый обработчик, если был
        const newSelect = materialSelect.cloneNode(true);
        materialSelect.parentNode.replaceChild(newSelect, materialSelect);
        
        newSelect.addEventListener('change', (e) => loadMaterialParameters(e.target.value));
        loadMaterialParameters(newSelect.value);
      }
    } catch (e) {
      console.error('Ошибка загрузки списка материалов:', e);
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
    if (!container) return;
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
      const chart3d = document.getElementById('chart3d');
      if (chart3d) chart3d.innerHTML = '<p style="color:red;">Ошибка загрузки данных 3D графика.</p>';
    }
  }

  if (form) {
    checkAuth().then(() => {
      loadMaterials();
    });
    
    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      if (errorDiv) errorDiv.textContent = '';
      
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

      // Проверка на валидность
      if (isNaN(data.w) || isNaN(data.h) || isNaN(data.l) || 
          isNaN(data.vu) || isNaN(data.tu) || isNaN(data.material_id)) {
        if (errorDiv) errorDiv.textContent = 'Заполните все поля корректными числами';
        return;
      }

      try {
        const result = await FlowModelAPI.client.calculation.calculate(data);
        
        lastBaseInput = data;
        lastBaseViscosity = result.viscosity;

        if (result.metrics && resultsDiv) {
          const metricTime = document.getElementById('metric-time');
          const metricMemory = document.getElementById('metric-memory');
          const metricOps = document.getElementById('metric-ops');
          if (metricTime) metricTime.textContent = result.metrics.calc_time_ms || 0;
          if (metricMemory) metricMemory.textContent = result.metrics.memory_used_bytes || 0;
          if (metricOps) metricOps.textContent = result.metrics.operations_count || result.metrics.operation_counts || 0;
        }

        const productivityEl = document.getElementById('productivity');
        const temperatureEl = document.getElementById('temperature');
        const viscosityEl = document.getElementById('viscosity');
        if (productivityEl) productivityEl.textContent = result.productivity.toFixed(4) + ' кг/ч';
        if (temperatureEl) temperatureEl.textContent = result.temperature.toFixed(1);
        if (viscosityEl) viscosityEl.textContent = result.viscosity.toFixed(1);
        
        if (resultsDiv) resultsDiv.style.display = 'block';
        renderProfileTable(result.profile);
        render2DChart(result.profile);

        // Устанавливаем значения для surface формы
        const surfVuMin = document.getElementById('surf_vu_min');
        const surfVuMax = document.getElementById('surf_vu_max');
        const surfTuMin = document.getElementById('surf_tu_min');
        const surfTuMax = document.getElementById('surf_tu_max');
        
        if (surfVuMin) surfVuMin.value = Math.max(data.vu * 0.5, 0.01).toFixed(2);
        if (surfVuMax) surfVuMax.value = (data.vu * 1.5).toFixed(2);
        if (surfTuMin) surfTuMin.value = Math.max(data.tu - 20, 0).toFixed(0);
        if (surfTuMax) surfTuMax.value = (data.tu + 20).toFixed(0);
        
        const vuStepsEl = document.getElementById('surf_vu_steps');
        const tuStepsEl = document.getElementById('surf_tu_steps');
        
        if (vuStepsEl) vuStepsEl.value = 24;
        if (tuStepsEl) tuStepsEl.value = 24;

        const surfacePayload = {
          ...data,
          vu_min: surfVuMin ? parseFloat(surfVuMin.value) : data.vu * 0.5,
          vu_max: surfVuMax ? parseFloat(surfVuMax.value) : data.vu * 1.5,
          vu_steps: vuStepsEl ? parseInt(vuStepsEl.value) : 24,
          tu_min: surfTuMin ? parseFloat(surfTuMin.value) : data.tu - 20,
          tu_max: surfTuMax ? parseFloat(surfTuMax.value) : data.tu + 20,
          tu_steps: tuStepsEl ? parseInt(tuStepsEl.value) : 24
        };
        
        await loadAndRenderSurface(surfacePayload, lastBaseInput, lastBaseViscosity);

      } catch (e) {
        if (errorDiv) errorDiv.textContent = `Ошибка: ${e.message}`;
        if (e instanceof FlowModelAPI.AuthError) window.location.href = '/login';
      }
    });

    if (surfaceForm) {
      surfaceForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        if (!lastBaseInput) {
          alert('Сначала выполните расчёт!');
          return;
        }

        const vuMin = parseFloat(document.getElementById('surf_vu_min')?.value);
        const vuMax = parseFloat(document.getElementById('surf_vu_max')?.value);
        const tuMin = parseFloat(document.getElementById('surf_tu_min')?.value);
        const tuMax = parseFloat(document.getElementById('surf_tu_max')?.value);
        
        const vuStepsEl = document.getElementById('surf_vu_steps');
        const tuStepsEl = document.getElementById('surf_tu_steps');

        if (isNaN(vuMin) || isNaN(vuMax) || isNaN(tuMin) || isNaN(tuMax)) {
          alert("ОШИБКА: Все поля должны быть заполнены числами");
          return;
        }

        if (vuMax <= vuMin || tuMax <= tuMin) {
          alert("ОШИБКА: Значения MAX должны быть строго больше значений MIN");
          return;
        }

        const surfacePayload = {
          ...lastBaseInput,
          vu_min: vuMin,
          vu_max: vuMax,
          vu_steps: vuStepsEl ? parseInt(vuStepsEl.value) : 24,
          tu_min: tuMin,
          tu_max: tuMax,
          tu_steps: tuStepsEl ? parseInt(tuStepsEl.value) : 24
        };

        await loadAndRenderSurface(surfacePayload, lastBaseInput, lastBaseViscosity);
      });
    }
  }
})();