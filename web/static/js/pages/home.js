(function() {
  const form = document.getElementById('calculate-form');
  const surfaceForm = document.getElementById('surface-form');
  const resultsDiv = document.getElementById('results');
  const errorDiv = document.getElementById('error-message');
  
  let chart2d = null;
  
  let lastBaseInput = null; 
  let lastBaseViscosity = null;
  let materialParametersRequestId = 0;

  function escapeHtml(value) {
    return String(value ?? '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

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
        select.innerHTML = materials
          .map(m => `<option value="${escapeHtml(m.id)}">${escapeHtml(m.name)}</option>`)
          .join('');
        await loadMaterialParameters(select.value);
      }
    } catch (e) {
      console.error('Ошибка загрузки материалов:', e);
    }
  }

  function formatParameterValue(value) {
    const num = Number(value);
    if (!Number.isFinite(num)) return '-';
    return Number.isInteger(num) ? String(num) : String(Number(num.toFixed(6)));
  }

  function renderMaterialParameters(params) {
    const container = document.getElementById('material-parameters');
    if (!container) return;

    container.innerHTML = '';
    if (!params || !params.length) {
      const empty = document.createElement('div');
      empty.className = 'material-parameters__empty';
      empty.textContent = 'Параметры для материала не заданы';
      container.appendChild(empty);
      return;
    }

    const title = document.createElement('div');
    title.className = 'material-parameters__title';
    title.textContent = 'Параметры материала';
    container.appendChild(title);

    const table = document.createElement('table');
    table.className = 'material-parameters__table';
    const tbody = document.createElement('tbody');

    params.forEach((param) => {
      const row = document.createElement('tr');

      const nameCell = document.createElement('td');
      nameCell.textContent = param.name || param.code || '-';

      const valueCell = document.createElement('td');
      const unit = param.unit ? ` ${param.unit}` : '';
      valueCell.textContent = `${formatParameterValue(param.value_float)}${unit}`;

      row.appendChild(nameCell);
      row.appendChild(valueCell);
      tbody.appendChild(row);
    });

    table.appendChild(tbody);
    container.appendChild(table);
  }

  async function loadMaterialParameters(materialId) {
    const container = document.getElementById('material-parameters');
    if (!container || !materialId) return;

    const requestId = ++materialParametersRequestId;
    container.innerHTML = '<div class="material-parameters__empty">Загрузка параметров...</div>';

    try {
      const params = await FlowModelAPI.client.materials.parameters(materialId);
      if (requestId !== materialParametersRequestId) return;
      renderMaterialParameters(params);
    } catch (e) {
      if (requestId !== materialParametersRequestId) return;
      console.error('Ошибка загрузки параметров материала:', e);
      container.innerHTML = '<div class="material-parameters__empty">Не удалось загрузить параметры</div>';
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
    const container = document.querySelector('#surface-table');
    if (!container) return;
    
    if (!points || points.length === 0) {
      container.innerHTML = '<div class="card"><p style="text-align:center; color:var(--muted);">Нет данных для отображения</p></div>';
      return;
    }
    
    const vuSet = [...new Set(points.map(p => p.vu))].sort((a, b) => a - b);
    const tuSet = [...new Set(points.map(p => p.tu))].sort((a, b) => a - b);
    
    const viscosityMap = {};
    points.forEach(p => {
      const key = `${p.vu}|${p.tu}`;
      viscosityMap[key] = p.viscosity;
    });
    
    let html = `
      <div class="card">
        <h3>Таблица для 3D графика (зависимость вязкости от температуры и скорости крышки)</h3>
        <div class="table-wrapper" style="max-height: 400px; overflow: auto;">
          <table style="min-width: 500px;">
            <thead>
              <tr>
                <th style="position: sticky; left: 0; background: var(--text); z-index: 2;">Vu \ d\Tu</th>
    `;
    
    tuSet.forEach(tu => {
      html += `<th style="white-space: nowrap;">${tu.toFixed(0)}°C</th>`;
    });
    
    html += `
              </tr>
            </thead>
            <tbody>
    `;
    
    vuSet.forEach(vu => {
      html += `<tr><td style="position: sticky; left: 0; background: var(--card); font-weight: bold; white-space: nowrap;">${vu.toFixed(2)} м/с</td>`;
      
      tuSet.forEach(tu => {
        const key = `${vu}|${tu}`;
        const viscosity = viscosityMap[key];
        const value = viscosity !== undefined ? viscosity.toFixed(1) : '—';
        html += `<td style="text-align: center; white-space: nowrap;">${value}</td>`;
      });
      
      html += `</tr>`;
    });
    
    html += `
            </tbody>
          </table>
        </div>
      </div>
    `;
    
    container.innerHTML = html;
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
          x: { title: { display: true, text: 'Координата по длине канала, м' } },
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
      document.getElementById('chart3d').innerHTML = '<p style="color:red;">Ошибка загрузки данных 3D графика: ' + (e.message || 'неизвестная ошибка') + '</p>';
    }
  }

  if (form) {
    checkAuth().then(() => {
      loadMaterials();
    });

    const materialSelect = document.querySelector('select[name="material_id"]');
    if (materialSelect) {
      materialSelect.addEventListener('change', () => loadMaterialParameters(materialSelect.value));
    }
    
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

        document.getElementById('productivity').textContent = Math.round(result.productivity) + ' кг/ч';
        document.getElementById('temperature').textContent = result.temperature.toFixed(1);
        document.getElementById('viscosity').textContent = result.viscosity.toFixed(1);
        
        resultsDiv.style.display = 'block';
        renderProfileTable(result.profile);
        render2DChart(result.profile);

        document.getElementById('surf_vu_min').value = Math.max(data.vu * 0.5, 0.01).toFixed(2);
        document.getElementById('surf_vu_max').value = (data.vu * 1.5).toFixed(2);
        document.getElementById('surf_tu_min').value = Math.max(data.tu - 20, 0).toFixed(0);
        document.getElementById('surf_tu_max').value = (data.tu + 20).toFixed(0);
        document.getElementById('surf_steps').value = Math.min(Math.max(data.steps, 3), 50);

        const surfacePayload = {
          ...data,
          vu_min: parseFloat(document.getElementById('surf_vu_min').value),
          vu_max: parseFloat(document.getElementById('surf_vu_max').value),
          vu_steps: parseInt(document.getElementById('surf_steps').value) || 24,
          tu_min: parseFloat(document.getElementById('surf_tu_min').value),
          tu_max: parseFloat(document.getElementById('surf_tu_max').value),
          tu_steps: parseInt(document.getElementById('surf_steps').value) || 24,
          steps: parseInt(document.getElementById('surf_steps').value) || 24
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
        let steps = parseInt(document.getElementById('surf_steps').value);

        if (isNaN(steps)) steps = 24;
        steps = Math.min(Math.max(steps, 3), 50);
        document.getElementById('surf_steps').value = steps;

        if (vu_max <= vu_min || tu_max <= tu_min) {
          alert("ОШИБКА: Значения MAX должны быть строго больше значений MIN");
          return;
        }

        const surfacePayload = {
          w: lastBaseInput.w,
          h: lastBaseInput.h,
          l: lastBaseInput.l,
          material_id: lastBaseInput.material_id,
          vu_min: vu_min,
          vu_max: vu_max,
          vu_steps: steps,
          tu_min: tu_min,
          tu_max: tu_max,
          tu_steps: steps,
          steps: steps
        };

        await loadAndRenderSurface(surfacePayload, lastBaseInput, lastBaseViscosity);
      });
    }
  }
})();