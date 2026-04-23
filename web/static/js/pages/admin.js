(async function() {
  function escapeHtml(value) {
    return String(value ?? '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  function getOptionalString(formData, name) {
    const value = String(formData.get(name) || '').trim();
    return value || null;
  }

  function showFormMessage(id, message, isError = true) {
    const el = document.getElementById(id);
    if (!el) return;
    el.textContent = message || '';
    el.style.color = isError ? 'var(--pink)' : 'var(--text)';
  }

  async function checkAdmin() {
    const token = sessionStorage.getItem('flowmodel_token');
    if (!token) { window.location.href = '/login'; return false; }
    try {
      const user = await FlowModelAPI.client.auth.me();
      if (user.role !== 'admin') { window.location.href = '/cabinet'; return false; }
      return user;
    } catch (e) { window.location.href = '/login'; return false; }
  }

  let currentUserId = null;

  let cachedMaterials = [];
  let cachedParameters = [];
  let materialParameterValues = {};

  const tabs = {
    materials: document.getElementById('tab-materials'),
    parameters: document.getElementById('tab-parameters'),
    users: document.getElementById('tab-users')
  };

  function getSelectedMaterialId() {
    const select = document.getElementById('material-parameter-material');
    return select?.value ? Number(select.value) : null;
  }

  function normalizeMaterialParameterValues(data) {
    if (!data) return {};
    if (Array.isArray(data)) {
      return data.reduce((acc, item) => {
        const code = item.code || item.parameter_code || item.parameter?.code;
        if (!code) return acc;
        acc[code] = item.value_float ?? item.valueFloat ?? item.value_string ?? item.valueString ?? item.value;
        return acc;
      }, {});
    }
    if (Array.isArray(data.values)) return normalizeMaterialParameterValues(data.values);
    if (Array.isArray(data.parameters)) return normalizeMaterialParameterValues(data.parameters);
    return Object.keys(data).reduce((acc, key) => {
      if (typeof data[key] !== 'object' || data[key] === null) {
        acc[key] = data[key];
      }
      return acc;
    }, {});
  }

  function populateMaterialParameterSelect() {
    const select = document.getElementById('material-parameter-material');
    if (!select) return;
    const previous = select.value;
    select.innerHTML = cachedMaterials.map(m => 
      `<option value="${escapeHtml(m.id)}">${escapeHtml(m.name)}</option>`
    ).join('');
    if (previous && cachedMaterials.some(m => String(m.id) === previous)) {
      select.value = previous;
    }
  }

  function renderMaterialParameterFields() {
    const tbody = document.querySelector('#material-parameters-table tbody');
    if (!tbody) return;
    if (!cachedMaterials.length) {
      tbody.innerHTML = '<tr><td colspan="6">Сначала добавьте материал</td></tr>';
      return;
    }
    if (!cachedParameters.length) {
      tbody.innerHTML = '<tr><td colspan="6">Сначала добавьте параметры</td></tr>';
      return;
    }
    tbody.innerHTML = cachedParameters.map(p => {
      const hasValue = Object.prototype.hasOwnProperty.call(materialParameterValues, p.code);
      const value = hasValue ? materialParameterValues[p.code] : '';
      const inputType = p.data_type === 'string' ? 'text' : 'number';
      const step = p.data_type === 'int' ? '1' : 'any';
      return `
        <tr>
          <td><input type="checkbox" name="bind" value="${escapeHtml(p.id)}" data-code="${escapeHtml(p.code)}" ${hasValue ? 'checked' : ''}></td>
          <td>${escapeHtml(p.code)}</td>
          <td>${escapeHtml(p.name)}</td>
          <td>${escapeHtml(p.data_type)}</td>
          <td>${escapeHtml(p.unit || '-')}</td>
          <td><input type="${inputType}" step="${step}" name="value_${escapeHtml(p.id)}" value="${escapeHtml(value ?? '')}"></td>
        </tr>
      `;
    }).join('');
  }

  async function loadMaterialParameterValues() {
    const materialId = getSelectedMaterialId();
    materialParameterValues = {};
    showFormMessage('material-parameters-message', '');
    if (!materialId) {
      renderMaterialParameterFields();
      return;
    }
    try {
      const data = await FlowModelAPI.client.admin.materials.parameters.list(materialId);
      materialParameterValues = normalizeMaterialParameterValues(data);
    } catch (e) {
      console.error(e);
      showFormMessage('material-parameters-message', 'Не удалось загрузить параметры материала');
    }
    renderMaterialParameterFields();
  }

  async function loadMaterials() {
    const tbody = document.querySelector('#materials-table tbody') || 
                  document.querySelector('#materials-table').appendChild(document.createElement('tbody'));
    try {
      const materials = await FlowModelAPI.client.admin.materials.list();
      cachedMaterials = materials;
      tbody.innerHTML = materials.map(m => `
        <tr>
          <td>${m.id}</td>
          <td>${escapeHtml(m.name)}</td>
          <td>${escapeHtml(m.description || '-')}</td>
          <td><button class="btn btn-small" onclick="deleteMaterial(${m.id})">🗑️</button></td>
        </tr>
      `).join('');
      populateMaterialParameterSelect();
      await loadMaterialParameterValues();
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
      cachedParameters = params;
      tbody.innerHTML = params.map(p => `
        <tr>
          <td>${p.id}</td>
          <td>${escapeHtml(p.code)}</td>
          <td>${escapeHtml(p.name)}</td>
          <td>${escapeHtml(p.unit || '-')}</td>
          <td>${escapeHtml(p.data_type)}</td>
          <td>${escapeHtml(p.category)}</td>
          <td><button class="btn btn-small" onclick="deleteParameter(${p.id})">🗑️</button></td>
        </tr>
      `).join('');
      renderMaterialParameterFields();
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
          <td>${escapeHtml(u.login)}</td>
          <td>${escapeHtml(u.role)}</td>
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

  if (addBtn) {
    addBtn.addEventListener('click', () => {
      modal.style.display = 'flex';
      errorDiv.textContent = '';
      createForm.reset();
    });
  }

  if (cancelBtn) {
    cancelBtn.addEventListener('click', () => {
      modal.style.display = 'none';
    });
  }

  if (modal) {
    modal.addEventListener('click', (e) => {
      if (e.target === modal) modal.style.display = 'none';
    });
  }

  if (createForm) {
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
  }

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

  const user = await checkAdmin();
  if (user) {
    currentUserId = user.id;
    loadMaterials();
  }
})();