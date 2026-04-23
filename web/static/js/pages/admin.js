(async function() {
  async function checkAdmin() {
    const token = sessionStorage.getItem('flowmodel_token');
    if (!token) { window.location.href = '/login'; return false; }
    try {
      const user = await FlowModelAPI.client.auth.me();
      if (user.role !== 'admin') { window.location.href = '/cabinet'; return false; }
      return user; // Возвращаем текущего пользователя
    } catch (e) { window.location.href = '/login'; return false; }
  }

  let currentUserId = null;

  const tabs = {
    materials: document.getElementById('tab-materials'),
    parameters: document.getElementById('tab-parameters'),
    users: document.getElementById('tab-users')
  };

  document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      Object.values(tabs).forEach(tab => tab.style.display = 'none');
      tabs[btn.dataset.tab].style.display = 'block';
      
      if (btn.dataset.tab === 'materials') loadMaterials();
      else if (btn.dataset.tab === 'parameters') loadParameters();
      else if (btn.dataset.tab === 'users') loadUsers();
    });
  });

  async function loadMaterials() {
    const tbody = document.querySelector('#materials-table tbody') || 
                  document.querySelector('#materials-table').appendChild(document.createElement('tbody'));
    try {
      const materials = await FlowModelAPI.client.admin.materials.list();
      tbody.innerHTML = materials.map(m => `
        <tr>
          <td>${m.id}</td>
          <td>${m.name}</td>
          <td>${m.description || '-'}</td>
          <td><button class="btn btn-small" onclick="deleteMaterial(${m.id})">🗑️</button></td>
        </tr>
      `).join('');
    } catch (e) { console.error(e); }
  }

  window.deleteMaterial = async (id) => {
    if (!confirm('Удалить материал?')) return;
    await FlowModelAPI.client.admin.materials.remove(id);
    loadMaterials();
  };

  async function loadParameters() {
    const tbody = document.querySelector('#parameters-table tbody') ||
                  document.querySelector('#parameters-table').appendChild(document.createElement('tbody'));
    try {
      const params = await FlowModelAPI.client.admin.parameters.list();
      tbody.innerHTML = params.map(p => `
        <tr>
          <td>${p.id}</td>
          <td>${p.code}</td>
          <td>${p.name}</td>
          <td>${p.unit || '-'}</td>
          <td>${p.data_type}</td>
          <td>${p.category}</td>
          <td><button class="btn btn-small" onclick="deleteParameter(${p.id})">🗑️</button></td>
        </tr>
      `).join('');
    } catch (e) { console.error(e); }
  }

  window.deleteParameter = async (id) => {
    if (!confirm('Удалить параметр?')) return;
    await FlowModelAPI.client.admin.parameters.remove(id);
    loadParameters();
  };

  async function loadUsers() {
    const tbody = document.querySelector('#users-table tbody') ||
                  document.querySelector('#users-table').appendChild(document.createElement('tbody'));
    try {
      const users = await FlowModelAPI.client.admin.users.list();
      tbody.innerHTML = users.map(u => `
        <tr>
          <td>${u.id}</td>
          <td>${u.login}</td>
          <td>${u.role}</td>
          <td>${new Date(u.created_at).toLocaleDateString()}</td>
          <td>
            ${u.id !== currentUserId ? 
              `<button class="btn btn-small" onclick="deleteUser(${u.id})">🗑️</button>` : 
              '<span style="color: var(--muted);" title="Нельзя удалить самого себя">текущий</span>'}
          </td>
        </tr>
      `).join('');
    } catch (e) { console.error(e); }
  }

  window.deleteUser = async (id) => {
    if (id === currentUserId) {
      alert('Нельзя удалить самого себя!');
      return;
    }
    if (!confirm('Удалить пользователя?')) return;
    try {
      await FlowModelAPI.client.admin.users.remove(id);
      loadUsers();
    } catch (e) {
      alert('Ошибка: ' + e.message);
    }
  };

  const modal = document.getElementById('create-user-modal');
  const addBtn = document.getElementById('add-user-btn');
  const cancelBtn = document.getElementById('cancel-create-user');
  const createForm = document.getElementById('create-user-form');
  const errorDiv = document.getElementById('create-user-error');

  addBtn.addEventListener('click', () => {
    modal.style.display = 'flex';
    errorDiv.textContent = '';
    createForm.reset();
  });

  cancelBtn.addEventListener('click', () => {
    modal.style.display = 'none';
  });

  modal.addEventListener('click', (e) => {
    if (e.target === modal) modal.style.display = 'none';
  });

  createForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    errorDiv.textContent = '';

    const formData = new FormData(createForm);
    const data = {
      login: formData.get('login'),
      password: formData.get('password'),
      role: formData.get('role')
    };

    try {
      await FlowModelAPI.client.admin.users.create(data);
      modal.style.display = 'none';
      loadUsers();
    } catch (e) {
      errorDiv.textContent = e.message;
    }
  });

  const user = await checkAdmin();
  if (user) {
    currentUserId = user.id;
    loadMaterials();
  }
})();