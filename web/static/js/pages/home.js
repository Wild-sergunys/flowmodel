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

  function render3DChart(profile) {
    const container = document.getElementById('chart3d');
    container.innerHTML = '';
    
    const width = container.clientWidth || 800;
    const height = 500;

    const scene = new THREE.Scene();
    scene.background = new THREE.Color('#f2ede8');

    const camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 1000);
    camera.position.set(2, 2.5, 3.5);
    camera.lookAt(0.5, 0.5, 0.4);

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
    controls.target.set(0.5, 0.5, 0.4);

    scene.add(new THREE.AmbientLight(0xffffff, 0.6));
    const dirLight1 = new THREE.DirectionalLight(0xffffff, 0.8);
    dirLight1.position.set(1, 2, 1);
    scene.add(dirLight1);
    const dirLight2 = new THREE.DirectionalLight(0xffffff, 0.4);
    dirLight2.position.set(-1, 1, -1);
    scene.add(dirLight2);

    // Нормализация
    const maxTemp = Math.max(...profile.map(p => p.temperature));
    const minTemp = Math.min(...profile.map(p => p.temperature));
    const maxVisc = Math.max(...profile.map(p => p.viscosity));
    const minVisc = Math.min(...profile.map(p => p.viscosity));
    
    const points = profile.map(p => ({
      temp: (p.temperature - minTemp) / (maxTemp - minTemp || 1),
      length: p.x / 10,
      visc: 1 - (p.viscosity - minVisc) / (maxVisc - minVisc || 1)  // Инвертировано!
    }));

    // Три линии, начинающиеся из нуля по длине
    const line1 = points.map(p => new THREE.Vector3(p.temp, p.length, 0));      // Длина от температуры
    const line2 = points.map(p => new THREE.Vector3(p.temp, 0, p.visc));       // Вязкость от температуры
    const line3 = points.map(p => new THREE.Vector3(0, p.length, p.visc));     // Вязкость от длины

    // Добавляем линии
    addLine(line1, 0xff3c8e);
    addLine(line2, 0xc8ff00);
    addLine(line3, 0x3366cc);

    // Одно полотно-треугольник
    const vertices = [];
    const indices = [];
    const n = points.length;
    
    for (let i = 0; i < n; i++) {
      vertices.push(line1[i].x, line1[i].y, line1[i].z);
      vertices.push(line2[i].x, line2[i].y, line2[i].z);
      vertices.push(line3[i].x, line3[i].y, line3[i].z);
    }
    
    for (let i = 0; i < n - 1; i++) {
      const base = i * 3;
      indices.push(base, base + 1, base + 2);
      indices.push(base, base + 3, base + 1);
      indices.push(base + 1, base + 3, base + 4);
      indices.push(base + 1, base + 4, base + 2);
      indices.push(base + 2, base + 4, base + 5);
      indices.push(base + 2, base + 5, base);
      indices.push(base, base + 5, base + 3);
    }
    
    const geom = new THREE.BufferGeometry();
    geom.setAttribute('position', new THREE.Float32BufferAttribute(vertices, 3));
    geom.setIndex(indices);
    geom.computeVertexNormals();
    
    const mat = new THREE.MeshPhongMaterial({
      color: 0xff6600,
      transparent: true,
      opacity: 0.45,
      side: THREE.DoubleSide,
      emissive: 0x331100
    });
    
    const mesh = new THREE.Mesh(geom, mat);
    scene.add(mesh);

    function addLine(points, color) {
      const geom = new THREE.BufferGeometry().setFromPoints(points);
      const mat = new THREE.LineBasicMaterial({ color: color });
      const line = new THREE.Line(geom, mat);
      scene.add(line);
    }

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

    createLabel('Температура (°C)', '#ff3c8e', new THREE.Vector3(1.3, -0.1, -0.1), 18);
    createLabel('Длина (м)', '#1a1a1a', new THREE.Vector3(-0.1, 1.3, -0.1), 18);
    createLabel('Вязкость (Па·с)', '#c8ff00', new THREE.Vector3(-0.1, -0.1, 1.3), 18);

    [0, 0.25, 0.5, 0.75, 1.0].forEach(t => {
      createLabel((minTemp + t * (maxTemp - minTemp)).toFixed(0) + '°C', '#ff3c8e', new THREE.Vector3(t, -0.12, -0.12), 12);
    });

    [0, 0.25, 0.5, 0.75, 1.0].forEach(t => {
      createLabel((t * 10).toFixed(1) + 'м', '#1a1a1a', new THREE.Vector3(-0.15, t, -0.15), 12);
    });

    [0, 0.25, 0.5, 0.75, 1.0].forEach(t => {
      const val = (maxVisc - t * (maxVisc - minVisc)).toFixed(0);
      createLabel(val, '#c8ff00', new THREE.Vector3(-0.15, -0.15, t), 12);
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
        document.getElementById('productivity').textContent = result.productivity.toFixed(6);
        document.getElementById('temperature').textContent = result.temperature.toFixed(1);
        document.getElementById('viscosity').textContent = result.viscosity.toFixed(1);
        resultsDiv.style.display = 'block';
        render2DChart(result.profile);
        render3DChart(result.profile);
      } catch (e) {
        errorDiv.textContent = `Ошибка: ${e.message}`;
        if (e instanceof FlowModelAPI.AuthError) window.location.href = '/login';
      }
    });
  }
})();