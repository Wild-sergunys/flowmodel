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
    try {
      const user = await FlowModelAPI.client.auth.me();
      if (user.role !== 'admin') { 
        window.location.href = '/cabinet'; 
        return false; 
      }
      return user;
    } catch (e) { 
      window.location.href = '/login'; 
      return false; 
    }
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
    return {};
  }

  async function ensureParametersLoaded() {
    if (cachedParameters.length) return;
    try {
      const params = await FlowModelAPI.client.admin.parameters.list();
      cachedParameters = Array.isArray(params) ? params : [];
    } catch (e) {
      console.error('Не удалось загрузить параметры:', e);
    }
  }

  function populateMaterialParameterSelect() {
    const select = document.getElementById('material-parameter-material');
    if (!select) return;
    const previous = select.value;
    select.innerHTML = (cachedMaterials || []).map(m => 
      `<option value="${escapeHtml(m.id)}">${escapeHtml(m.name)}</option>`
    ).join('');
    if (previous && cachedMaterials.some(m => String(m.id) === previous)) {
      select.value = previous;
    }
  }

  function renderMaterialParameterFields() {
    const tbody = document.querySelector('#material-parameters-table tbody');
    if (!tbody) return;
    if (!cachedMaterials || !cachedMaterials.length) {
      tbody.innerHTML = '<tr><td colspan="6">Сначала добавьте материал</td></tr>';
      return;
    }
    if (!cachedParameters.length) {
      tbody.innerHTML = '<tr><td colspan="6">Параметры не загружены</td></tr>';
      return;
    }
    tbody.innerHTML = cachedParameters.map(p => {
      const hasValue = Object.prototype.hasOwnProperty.call(materialParameterValues, p.code);
      const value = hasValue ? materialParameterValues[p.code] : '';
      const inputType = p.data_type === 'string' ? 'text' : 'number';
      const step = p.data_type === 'int' ? '1' : 'any';
      
      let min = '';
      let max = '';
      if (p.code === 'density' || p.code === 'heat_capacity' || 
          p.code === 'melting_temp' || p.code === 'mu0' || 
          p.code === 'Ea' || p.code === 'Tr') {
        min = 'min="0.0001"';
      }
      if (p.code === 'n') {
        min = 'min="0.01"';
        max = 'max="0.99"';
      }
      if (p.code === 'alpha_u') {
        min = 'min="0"';
      }
      
      return `
        <tr>
          <td><input type="checkbox" name="bind" value="${escapeHtml(p.id)}" data-code="${escapeHtml(p.code)}" ${hasValue ? 'checked' : ''}></td>
          <td>${escapeHtml(p.code)}</td>
          <td>${escapeHtml(p.name)}</td>
          <td>${escapeHtml(p.data_type)}</td>
          <td>${escapeHtml(p.unit || '-')}</td>
          <td><input type="${inputType}" step="${step}" name="value_${escapeHtml(p.id)}" value="${escapeHtml(value ?? '')}" ${min} ${max}></td>
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
    await ensureParametersLoaded();
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
      cachedMaterials = Array.isArray(materials) ? materials : [];
      tbody.innerHTML = cachedMaterials.map(m => `
        <tr>
          <td>${m.id}</td>
          <td>${escapeHtml(m.name)}</td>
          <td>${escapeHtml(m.description || '-')}</td>
          <td><button class="btn btn-small" onclick="deleteMaterial(${m.id})">🗑️</button></td>
        </tr>
      `).join('');
      populateMaterialParameterSelect();
      await ensureParametersLoaded();
      await loadMaterialParameterValues();
    } catch (e) { console.error(e); }
  }

  window.deleteMaterial = async (id) => {
    if (!confirm('Удалить материал?')) return;
    await FlowModelAPI.client.admin.materials.remove(id);
    loadMaterials();
  };

  const materialForm = document.getElementById('material-form');
  if (materialForm) {
    materialForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      showFormMessage('material-form-message', '');
      const fd = new FormData(materialForm);
      const name = String(fd.get('name') || '').trim();
      if (!name) {
        showFormMessage('material-form-message', 'Введите название материала');
        return;
      }
      try {
        await FlowModelAPI.client.admin.materials.create({
          name,
          description: getOptionalString(fd, 'description'),
        });
        materialForm.reset();
        showFormMessage('material-form-message', 'Материал добавлен', false);
        await loadMaterials();
      } catch (e) {
        showFormMessage('material-form-message', e.message || 'Не удалось добавить материал');
      }
    });
  }

  async function loadParameters() {
    const tbody = document.querySelector('#parameters-table tbody') ||
                  document.querySelector('#parameters-table').appendChild(document.createElement('tbody'));
    try {
      const params = await FlowModelAPI.client.admin.parameters.list();
      cachedParameters = Array.isArray(params) ? params : [];
      tbody.innerHTML = cachedParameters.map(p => `
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

  const parameterForm = document.getElementById('parameter-form');
  if (parameterForm) {
    parameterForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      showFormMessage('parameter-form-message', '');
      const fd = new FormData(parameterForm);
      const code = String(fd.get('code') || '').trim();
      const name = String(fd.get('name') || '').trim();
      if (!code || !/^[a-z0-9_]+$/.test(code)) {
        showFormMessage('parameter-form-message', 'Условное обозначение должно содержать только латинские строчные буквы, цифры и _');
        return;
      }
      if (!name) {
        showFormMessage('parameter-form-message', 'Введите название параметра');
        return;
      }
      try {
        await FlowModelAPI.client.admin.parameters.create({
          code,
          name,
          unit: getOptionalString(fd, 'unit'),
          data_type: String(fd.get('data_type') || 'float'),
          category: String(fd.get('category') || 'material_property'),
          description: getOptionalString(fd, 'description'),
        });
        parameterForm.reset();
        showFormMessage('parameter-form-message', 'Параметр добавлен', false);
        await loadParameters();
      } catch (e) {
        showFormMessage('parameter-form-message', e.message || 'Не удалось добавить параметр');
      }
    });
  }

  const materialParameterSelect = document.getElementById('material-parameter-material');
  if (materialParameterSelect) {
    materialParameterSelect.addEventListener('change', loadMaterialParameterValues);
  }

  const reloadBtn = document.getElementById('reload-material-parameters');
  if (reloadBtn) {
    reloadBtn.addEventListener('click', async () => {
      const materialId = getSelectedMaterialId();
      if (!materialId) {
        showFormMessage('material-parameters-message', 'Выберите материал');
        return;
      }
      showFormMessage('material-parameters-message', 'Обновление...', false);
      await loadMaterialParameterValues();
      showFormMessage('material-parameters-message', 'Обновлено', false);
    });
  }

  const materialParametersForm = document.getElementById('material-parameters-form');
  if (materialParametersForm) {
    materialParametersForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      showFormMessage('material-parameters-message', '');

      const materialId = getSelectedMaterialId();
      if (!materialId) {
        showFormMessage('material-parameters-message', 'Выберите материал');
        return;
      }

      const checked = Array.from(materialParametersForm.querySelectorAll('input[name="bind"]:checked'));
      const parameters = checked.map(checkbox => {
        const parameterId = Number(checkbox.value);
        const code = checkbox.dataset.code;
        const input = materialParametersForm.querySelector(`[name="value_${parameterId}"]`);
        const rawValue = String(input?.value ?? '').trim();
        const parameter = cachedParameters.find(p => Number(p.id) === parameterId);
        const isString = parameter?.data_type === 'string';

        return {
          parameter_id: parameterId,
          code: code,
          value: isString ? rawValue : Number(rawValue),
          value_float: isString ? null : Number(rawValue),
          value_string: isString ? rawValue : null,
        };
      });

      // Проверка значений
      const invalid = parameters.find(p => {
        if (p.value === '' || (typeof p.value === 'number' && !Number.isFinite(p.value))) return true;
        if (p.code === 'n' && (p.value <= 0 || p.value >= 1)) return true;
        if ((p.code === 'density' || p.code === 'heat_capacity' || 
            p.code === 'melting_temp' || p.code === 'mu0' || 
            p.code === 'Ea' || p.code === 'Tr') && p.value <= 0) return true;
        if (p.code === 'alpha_u' && p.value < 0) return true;
        return false;
      });

      if (invalid) {
        showFormMessage('material-parameters-message', 
          'Проверьте значения: плотность/теплоёмкость/температуры/μ0/Ea > 0, n ∈ (0; 1), αu ≥ 0');
        return;
      }

      const values = {};
      parameters.forEach(p => {
        values[p.code] = p.value;
      });

      try {
        await FlowModelAPI.client.admin.materials.parameters.update(materialId, {
          parameters,
          values,
        });
        showFormMessage('material-parameters-message', 'Параметры материала сохранены', false);
        await loadMaterialParameterValues();
      } catch (e) {
        console.error(e);
        showFormMessage('material-parameters-message', e.message || 'Не удалось сохранить параметры');
      }
    });
  }

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
    cancelBtn.addEventListener('click', () => { modal.style.display = 'none'; });
  }
  if (modal) {
    modal.addEventListener('click', (e) => { if (e.target === modal) modal.style.display = 'none'; });
  }
  if (createForm) {
    createForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      errorDiv.textContent = '';
      const formData = new FormData(createForm);
      try {
        await FlowModelAPI.client.admin.users.create({
          login: formData.get('login'),
          password: formData.get('password'),
          role: formData.get('role')
        });
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

  // Инициализация
  const user = await checkAdmin();
  if (user) {
    currentUserId = user.id;
    loadMaterials();
  }
})();