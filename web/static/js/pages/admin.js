(function() {
  async function checkAdmin() {
    const token = sessionStorage.getItem('flowmodel_token');
    if (!token) { window.location.href = '/login'; return false; }
    try {
      const user = await FlowModelAPI.client.auth.me();
      if (user.role !== 'admin') { window.location.href = '/cabinet'; return false; }
      return true;
    } catch (e) { window.location.href = '/login'; return false; }
  }

  function getTbody(tableId) {
    let tbody = document.querySelector(`#${tableId} tbody`);
    if (!tbody) {
      tbody = document.createElement('tbody');
      document.querySelector(`#${tableId}`).appendChild(tbody);
    }
    return tbody;
  }

  async function loadMaterials() {
    const tbody = getTbody('materials-table');
    try {
      const materials = await FlowModelAPI.client.admin.materials.list();
      tbody.innerHTML = materials.map(m => `
        <tr><td>${m.id}</td><td>${m.name}</td><td>${m.description || '-'}</td>
        <td><button class="btn btn-small" onclick="deleteMaterial(${m.id})" title="Удалить">🗑️</button></td></tr>
      `).join('');
    } catch (e) { console.error(e); }
  }

  async function loadParameters() {
    const tbody = getTbody('parameters-table');
    try {
      const params = await FlowModelAPI.client.admin.parameters.list();
      tbody.innerHTML = params.map(p => `
        <tr><td>${p.id}</td><td>${p.code}</td><td>${p.name}</td><td>${p.unit || '-'}</td><td>${p.data_type}</td><td>${p.category}</td>
        <td><button class="btn btn-small" onclick="deleteParameter(${p.id})" title="Удалить">🗑️</button></td></tr>
      `).join('');
    } catch (e) { console.error(e); }
  }

  async function loadUsers() {
    const tbody = getTbody('users-table');
    try {
      const users = await FlowModelAPI.client.admin.users.list();
      tbody.innerHTML = users.map(u => `
        <tr><td>${u.id}</td><td>${u.login}</td><td>${u.role}</td><td>${new Date(u.created_at).toLocaleDateString()}</td>
        <td><button class="btn btn-small" onclick="deleteUser(${u.id})" title="Удалить">🗑️</button></td></tr>
      `).join('');
    } catch (e) { console.error(e); }
  }

  window.deleteMaterial = async (id) => { if (confirm('Удалить материал?')) { await FlowModelAPI.client.admin.materials.remove(id); loadMaterials(); } };
  window.deleteParameter = async (id) => { if (confirm('Удалить параметр?')) { await FlowModelAPI.client.admin.parameters.remove(id); loadParameters(); } };
  window.deleteUser = async (id) => { if (confirm('Удалить пользователя?')) { await FlowModelAPI.client.admin.users.remove(id); loadUsers(); } };

  document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      document.querySelectorAll('.tab-content').forEach(tab => tab.style.display = 'none');
      document.getElementById(`tab-${btn.dataset.tab}`).style.display = 'block';
      if (btn.dataset.tab === 'materials') loadMaterials();
      else if (btn.dataset.tab === 'parameters') loadParameters();
      else if (btn.dataset.tab === 'users') loadUsers();
    });
  });

  (async () => { if (await checkAdmin()) loadMaterials(); })();
})();