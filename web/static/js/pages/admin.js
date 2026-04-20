(function() {
  let cachedMaterials = [];
  let cachedParameters = [];
  let materialParameterValues = {};

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
    select.innerHTML = cachedMaterials.map(m => (
      `<option value="${escapeHtml(m.id)}">${escapeHtml(m.name)}</option>`
    )).join('');

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
      showFormMessage(
        'material-parameters-message',
        e.status === 404
          ? 'Эндпоинт параметров материала недоступен или материал не найден'
          : (e.message || 'Не удалось загрузить параметры материала')
      );
    }

    renderMaterialParameterFields();
  }

  async function loadMaterials() {
    const tbody = getTbody('materials-table');
    try {
      const materials = await FlowModelAPI.client.admin.materials.list();
      cachedMaterials = materials;
      tbody.innerHTML = materials.map(m => `
        <tr><td>${escapeHtml(m.id)}</td><td>${escapeHtml(m.name)}</td><td>${escapeHtml(m.description || '-')}</td>
        <td><button class="btn btn-small" onclick="deleteMaterial(${m.id})" title="Удалить">🗑️</button></td></tr>
      `).join('');
      populateMaterialParameterSelect();
      await loadMaterialParameterValues();
    } catch (e) { console.error(e); }
  }

  async function loadParameters() {
    const tbody = getTbody('parameters-table');
    try {
      const params = await FlowModelAPI.client.admin.parameters.list();
      cachedParameters = params;
      tbody.innerHTML = params.map(p => `
        <tr><td>${escapeHtml(p.id)}</td><td>${escapeHtml(p.code)}</td><td>${escapeHtml(p.name)}</td><td>${escapeHtml(p.unit || '-')}</td><td>${escapeHtml(p.data_type)}</td><td>${escapeHtml(p.category)}</td>
        <td><button class="btn btn-small" onclick="deleteParameter(${p.id})" title="Удалить">🗑️</button></td></tr>
      `).join('');
      renderMaterialParameterFields();
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

      const payload = {
        name,
        description: getOptionalString(fd, 'description'),
      };

      try {
        await FlowModelAPI.client.admin.materials.create(payload);
        materialForm.reset();
        showFormMessage('material-form-message', 'Материал добавлен', false);
        await loadMaterials();
      } catch (e) {
        console.error(e);
        showFormMessage('material-form-message', e.message || 'Не удалось добавить материал');
      }
    });
  }

  const parameterForm = document.getElementById('parameter-form');
  if (parameterForm) {
    parameterForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      showFormMessage('parameter-form-message', '');

      const fd = new FormData(parameterForm);
      const code = String(fd.get('code') || '').trim();
      const name = String(fd.get('name') || '').trim();
      if (!code || !/^[a-z0-9_]+$/.test(code)) {
        showFormMessage('parameter-form-message', 'Код должен содержать только латинские строчные буквы, цифры и _');
        return;
      }
      if (!name) {
        showFormMessage('parameter-form-message', 'Введите название параметра');
        return;
      }

      const payload = {
        code,
        name,
        unit: getOptionalString(fd, 'unit'),
        data_type: String(fd.get('data_type') || 'float'),
        category: String(fd.get('category') || 'material_property'),
        description: getOptionalString(fd, 'description'),
      };

      try {
        await FlowModelAPI.client.admin.parameters.create(payload);
        parameterForm.reset();
        showFormMessage('parameter-form-message', 'Параметр добавлен', false);
        await loadParameters();
      } catch (e) {
        console.error(e);
        showFormMessage('parameter-form-message', e.message || 'Не удалось добавить параметр');
      }
    });
  }

  const materialParameterSelect = document.getElementById('material-parameter-material');
  if (materialParameterSelect) {
    materialParameterSelect.addEventListener('change', loadMaterialParameterValues);
  }

  const reloadMaterialParametersBtn = document.getElementById('reload-material-parameters');
  if (reloadMaterialParametersBtn) {
    reloadMaterialParametersBtn.addEventListener('click', loadMaterialParameterValues);
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
        const parameter = cachedParameters.find(p => Number(p.id) === parameterId);
        const input = materialParametersForm.querySelector(`[name="value_${parameterId}"]`);
        const rawValue = String(input?.value ?? '').trim();
        const isString = parameter?.data_type === 'string';

        return {
          parameter_id: parameterId,
          code: parameter?.code || checkbox.dataset.code,
          value: isString ? rawValue : Number(rawValue),
          value_float: isString ? null : Number(rawValue),
          value_string: isString ? rawValue : null,
        };
      });

      const invalid = parameters.find(p => p.value === '' || (typeof p.value === 'number' && !Number.isFinite(p.value)));
      if (invalid) {
        showFormMessage('material-parameters-message', 'Заполните значения для всех привязанных параметров');
        return;
      }

      const values = parameters.reduce((acc, p) => {
        acc[p.code] = p.value;
        return acc;
      }, {});

      try {
        await FlowModelAPI.client.admin.materials.parameters.update(materialId, {
          parameters,
          values,
        });
        showFormMessage('material-parameters-message', 'Параметры материала сохранены', false);
        await loadMaterialParameterValues();
      } catch (e) {
        console.error(e);
        showFormMessage(
          'material-parameters-message',
          e.status === 404
            ? 'Эндпоинт сохранения параметров материала недоступен'
            : (e.message || 'Не удалось сохранить параметры материала')
        );
      }
    });
  }

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

  (async () => {
    if (await checkAdmin()) {
      await loadParameters();
      await loadMaterials();
    }
  })();
})();
