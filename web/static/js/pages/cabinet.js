(function() {
  async function loadHistory() {
    const token = sessionStorage.getItem('flowmodel_token');
    if (!token) { window.location.href = '/login'; return; }
    
    let tbody = document.querySelector('#history-table tbody');
    if (!tbody) {
      tbody = document.createElement('tbody');
      document.querySelector('#history-table').appendChild(tbody);
    }
    
    try {
      const results = await FlowModelAPI.client.results.list();
      const materials = await FlowModelAPI.client.materials.list();
      const materialMap = Object.fromEntries(materials.map(m => [m.id, m.name]));
      
      tbody.innerHTML = results.map(r => {
        const input = JSON.parse(r.input_json);
        const result = JSON.parse(r.result_json);
        return `
          <tr>
            <td>${r.id}</td>
            <td>${new Date(r.created_at).toLocaleString()}</td>
            <td>${materialMap[r.material_id] || '-'}</td>
            <td>${result.productivity.toFixed(6)}</td>
            <td>${result.temperature.toFixed(1)}</td>
            <td>${result.viscosity.toFixed(1)}</td>
            <td>
              <button class="btn btn-small view-btn" data-id="${r.id}" title="Просмотреть отчёт">👁️</button>
              <button class="btn btn-small download-btn" data-id="${r.id}" title="Скачать JSON">📥</button>
            </td>
          </tr>
        `;
      }).join('');
      
      document.querySelectorAll('.view-btn').forEach(btn => {
        btn.addEventListener('click', async () => {
          const report = await FlowModelAPI.client.results.report(btn.dataset.id);
          alert(JSON.stringify(report, null, 2));
        });
      });
      
      document.querySelectorAll('.download-btn').forEach(btn => {
        btn.addEventListener('click', async () => {
          const { blob, filename } = await FlowModelAPI.client.results.download(btn.dataset.id);
          const url = URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = url;
          a.download = filename;
          a.click();
          URL.revokeObjectURL(url);
        });
      });
      
    } catch (e) {
      console.error('Ошибка загрузки истории:', e);
      if (e instanceof FlowModelAPI.AuthError) window.location.href = '/login';
    }
  }
  
  loadHistory();
})();