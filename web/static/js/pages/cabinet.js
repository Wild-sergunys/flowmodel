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
            <td>${(result.productivity || 0).toFixed(2)} кг/ч</td>
            <td>${(result.temperature || 0).toFixed(1)} °C</td>
            <td>${(result.viscosity || 0).toFixed(1)} Па·с</td>
            <td>
              <button class="btn btn-small view-btn" data-id="${r.id}" title="Посмотреть отчёт">👁️</button>
              <button class="btn btn-small download-btn" data-id="${r.id}" title="Скачать Excel">📥</button>
            </td>
          </tr>
        `;
      }).join('');
      
      document.querySelectorAll('.view-btn').forEach(btn => {
        btn.addEventListener('click', async () => {
          try {
            const report = await FlowModelAPI.client.results.report(btn.dataset.id);
            
            // Форматируем для красивого показа
            let text = '';
            text += '=== ВХОДНЫЕ ПАРАМЕТРЫ ===\n';
            if (report.input) {
              const input = typeof report.input === 'string' ? JSON.parse(report.input) : report.input;
              text += `Ширина: ${input.w} м\n`;
              text += `Глубина: ${input.h} м\n`;
              text += `Длина: ${input.l} м\n`;
              text += `Скорость крышки: ${input.vu} м/с\n`;
              text += `Температура крышки: ${input.tu} °C\n`;
              text += `Шагов: ${input.steps}\n`;
            }
            text += '\n=== РЕЗУЛЬТАТЫ ===\n';
            if (report.result) {
              const result = typeof report.result === 'string' ? JSON.parse(report.result) : report.result;
              text += `Производительность: ${(result.productivity || 0).toFixed(2)} кг/ч\n`;
              text += `Температура продукта: ${(result.temperature || 0).toFixed(1)} °C\n`;
              text += `Вязкость продукта: ${(result.viscosity || 0).toFixed(1)} Па·с\n`;
            }
            alert(text);
          } catch (e) {
            alert('Ошибка загрузки отчёта');
          }
        });
      });
      
      document.querySelectorAll('.download-btn').forEach(btn => {
        btn.addEventListener('click', async () => {
          try {
            const { blob, filename } = await FlowModelAPI.client.results.download(btn.dataset.id);
            const excelFilename = filename.includes('.xlsx') ? filename : filename + '.xlsx';
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = excelFilename;
            a.click();
            URL.revokeObjectURL(url);
          } catch (e) {
            alert('Ошибка скачивания: ' + e.message);
          }
        });
      });
      
    } catch (e) {
      console.error('Ошибка загрузки истории:', e);
      if (e instanceof FlowModelAPI.AuthError) window.location.href = '/login';
    }
  }
  
  loadHistory();
})();