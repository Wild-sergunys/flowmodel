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
          y1: { type: 'linear', position: 'right', title: { display: true, text: 'Вязкость (Па·с)' }, grid: { drawOnChartArea: false } }
        }
      }
    });
  }

  function render3DChart(profile) {
    const container = document.getElementById('chart3d');
    if (!container) return;
    container.innerHTML = '';
    const width = container.clientWidth || 800, height = 400;
    const scene = new THREE.Scene();
    scene.background = new THREE.Color('#f2ede8');
    const camera = new THREE.PerspectiveCamera(45, width/height, 0.1, 1000);
    camera.position.set(4, 3, 5); camera.lookAt(0.5, 0.5, 0);
    const renderer = new THREE.WebGLRenderer({ antialias: true });
    renderer.setSize(width, height); renderer.setPixelRatio(1);
    container.appendChild(renderer.domElement);
    
    const controls = new THREE.OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true; controls.autoRotate = true; controls.autoRotateSpeed = 1.0;
    controls.target.set(0.5, 0.5, 0.5);
    
    scene.add(new THREE.AmbientLight(0xffffff, 0.6));
    const dirLight = new THREE.DirectionalLight(0xffffff, 0.8); dirLight.position.set(1, 2, 1); scene.add(dirLight);
    
    const points = profile.map(p => ({ x: p.x/10, temp: (p.temperature-140)/20, visc: (p.viscosity-5000)/2000 }));
    
    const tempCurve = new THREE.CatmullRomCurve3(points.map(p => new THREE.Vector3(p.x, p.temp, 0)));
    const tempTube = new THREE.Mesh(new THREE.TubeGeometry(tempCurve, 64, 0.02, 8, false), new THREE.MeshPhongMaterial({ color: 0xff3c8e, transparent: true, opacity: 0.7, side: THREE.DoubleSide }));
    scene.add(tempTube);
    
    const viscCurve = new THREE.CatmullRomCurve3(points.map(p => new THREE.Vector3(p.x, p.visc, 1)));
    const viscTube = new THREE.Mesh(new THREE.TubeGeometry(viscCurve, 64, 0.02, 8, false), new THREE.MeshPhongMaterial({ color: 0xc8ff00, transparent: true, opacity: 0.7, side: THREE.DoubleSide }));
    scene.add(viscTube);
    
    const gridHelper = new THREE.GridHelper(2, 20, 0x1a1a1a, 0xcccccc); gridHelper.position.y = -0.5; scene.add(gridHelper);
    scene.add(new THREE.AxesHelper(2));
    
    function animate() { requestAnimationFrame(animate); controls.update(); renderer.render(scene, camera); }
    animate();
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