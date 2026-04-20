document.documentElement.classList.add("js");

window.addEventListener("flowmodel:unauthorized", function () {
  if (window.location.pathname !== "/login") {
    window.sessionStorage.setItem("flowmodel:returnTo", window.location.pathname + window.location.search);
    window.location.assign("/login");
  }
});

(function () {
  "use strict";

  const page = document.querySelector("[data-login-page]");
  if (!page || !window.FlowModelAPI) {
    return;
  }

  const api = window.FlowModelAPI.client;
  const form = page.querySelector("[data-login-form]");
  const message = page.querySelector("[data-login-message]");
  const submitButton = page.querySelector("[data-login-submit]");

  function setMessage(text, type) {
    message.textContent = text || "";
    message.dataset.type = type || "";
  }

  function getReturnPath() {
    const savedPath = window.sessionStorage.getItem("flowmodel:returnTo");
    if (!savedPath || savedPath === "/login") {
      return "/";
    }

    return savedPath;
  }

  form.addEventListener("submit", async function (event) {
    event.preventDefault();
    setMessage("");

    if (!form.reportValidity()) {
      return;
    }

    submitButton.disabled = true;
    submitButton.textContent = "Вход...";

    try {
      await api.auth.login({
        login: form.elements.login.value.trim(),
        password: form.elements.password.value,
      });

      const returnPath = getReturnPath();
      window.sessionStorage.removeItem("flowmodel:returnTo");
      window.location.assign(returnPath);
    } catch (error) {
      setMessage(error.message || "Не удалось выполнить вход.", "error");
    } finally {
      submitButton.disabled = false;
      submitButton.textContent = "Войти";
    }
  });
})();

(function () {
  "use strict";

  const app = document.querySelector("[data-researcher-app]");
  if (!app || !window.FlowModelAPI) {
    return;
  }

  const api = window.FlowModelAPI.client;
  const form = app.querySelector("[data-calculation-form]");
  const panels = Array.from(app.querySelectorAll("[data-step-panel]"));
  const indicators = Array.from(app.querySelectorAll("[data-step-indicator]"));
  const prevButton = app.querySelector("[data-prev-step]");
  const nextButton = app.querySelector("[data-next-step]");
  const submitButton = app.querySelector("[data-submit-calculation]");
  const message = app.querySelector("[data-form-message]");
  const materialSelect = app.querySelector("[data-material-select]");
  const materialStatus = app.querySelector("[data-material-status]");
  const resultSummary = app.querySelector("[data-result-summary]");
  const chartArea = app.querySelector("[data-chart-area]");
  const reportPanel = app.querySelector("[data-report-panel]");
  const reportStatus = app.querySelector("[data-report-status]");
  const generateReportButton = app.querySelector("[data-generate-report]");
  const downloadReportButton = app.querySelector("[data-download-report]");

  let activeStep = 0;
  let calculationResult = null;
  let reportData = null;
  let reportReady = false;

  function setMessage(text, type) {
    message.textContent = text || "";
    message.dataset.type = type || "";
  }

  function setBusy(isBusy) {
    prevButton.disabled = isBusy || activeStep === 0;
    nextButton.disabled = isBusy;
    submitButton.disabled = isBusy;
    generateReportButton.disabled = isBusy || !calculationResult;
    downloadReportButton.disabled = isBusy || !reportReady;
  }

  function showStep(index) {
    activeStep = Math.max(0, Math.min(index, panels.length - 1));

    panels.forEach(function (panel, panelIndex) {
      panel.classList.toggle("is-active", panelIndex === activeStep);
    });
    indicators.forEach(function (indicator, indicatorIndex) {
      indicator.classList.toggle("is-active", indicatorIndex === activeStep);
      indicator.classList.toggle("is-complete", indicatorIndex < activeStep);
    });

    prevButton.disabled = activeStep === 0;
    nextButton.hidden = activeStep >= panels.length - 2;
    submitButton.hidden = activeStep !== panels.length - 2;
  }

  function getCurrentStepFields() {
    return Array.from(panels[activeStep].querySelectorAll("input, select"));
  }

  function validateCurrentStep() {
    const fields = getCurrentStepFields();
    for (const field of fields) {
      if (!field.reportValidity()) {
        return false;
      }
    }

    return true;
  }

  function parseNumber(name) {
    const field = form.elements[name];
    return Number(String(field.value).replace(",", "."));
  }

  function getPayload() {
    return {
      w: parseNumber("w"),
      h: parseNumber("h"),
      l: parseNumber("l"),
      material_id: Number(form.elements.material_id.value),
      vu: parseNumber("vu"),
      tu: parseNumber("tu"),
      t0: parseNumber("t0"),
      steps: Number(form.elements.steps.value),
    };
  }

  function formatNumber(value, digits) {
    if (!Number.isFinite(Number(value))) {
      return "--";
    }

    return Number(value).toLocaleString("ru-RU", {
      maximumFractionDigits: digits,
      minimumFractionDigits: 0,
    });
  }

  function formatScientific(value) {
    const number = Number(value);
    if (!Number.isFinite(number)) {
      return "--";
    }

    return number.toExponential(4).replace(".", ",");
  }

  function chartScale(values, size) {
    const min = Math.min.apply(null, values);
    const max = Math.max.apply(null, values);
    const span = max - min || 1;

    return function (value) {
      return size - ((value - min) / span) * size;
    };
  }

  function renderChart(svg, points, valueKey, stroke, unit) {
    const width = 640;
    const height = 320;
    const padding = { top: 24, right: 24, bottom: 44, left: 70 };
    const plotWidth = width - padding.left - padding.right;
    const plotHeight = height - padding.top - padding.bottom;
    const xValues = points.map(function (point) {
      return Number(point.x);
    });
    const yValues = points.map(function (point) {
      return Number(point[valueKey]);
    });
    const xMax = Math.max.apply(null, xValues) || 1;
    const yMin = Math.min.apply(null, yValues);
    const yMax = Math.max.apply(null, yValues);
    const scaleY = chartScale(yValues, plotHeight);

    function xCoord(value) {
      return padding.left + (Number(value) / xMax) * plotWidth;
    }

    function yCoord(value) {
      return padding.top + scaleY(Number(value));
    }

    const line = points
      .map(function (point) {
        return xCoord(point.x).toFixed(2) + "," + yCoord(point[valueKey]).toFixed(2);
      })
      .join(" ");
    const yTicks = [yMin, yMin + (yMax - yMin) / 2, yMax];

    svg.innerHTML = [
      '<rect class="chart__bg" x="0" y="0" width="' + width + '" height="' + height + '"></rect>',
      '<line class="chart__axis" x1="' +
        padding.left +
        '" y1="' +
        (height - padding.bottom) +
        '" x2="' +
        (width - padding.right) +
        '" y2="' +
        (height - padding.bottom) +
        '"></line>',
      '<line class="chart__axis" x1="' +
        padding.left +
        '" y1="' +
        padding.top +
        '" x2="' +
        padding.left +
        '" y2="' +
        (height - padding.bottom) +
        '"></line>',
      yTicks
        .map(function (tick) {
          const y = yCoord(tick).toFixed(2);
          return (
            '<line class="chart__grid" x1="' +
            padding.left +
            '" y1="' +
            y +
            '" x2="' +
            (width - padding.right) +
            '" y2="' +
            y +
            '"></line><text class="chart__tick" x="' +
            (padding.left - 12) +
            '" y="' +
            (Number(y) + 4) +
            '" text-anchor="end">' +
            formatNumber(tick, 2) +
            "</text>"
          );
        })
        .join(""),
      '<polyline class="chart__line" points="' + line + '" stroke="' + stroke + '"></polyline>',
      '<text class="chart__label" x="' + width / 2 + '" y="' + (height - 10) + '" text-anchor="middle">x, м</text>',
      '<text class="chart__label" x="14" y="' + height / 2 + '" transform="rotate(-90 14 ' + height / 2 + ')">' + unit + "</text>",
    ].join("");
  }

  function showResult(result) {
    calculationResult = result;
    reportData = null;
    reportReady = false;

    app.querySelector("[data-result-productivity]").textContent = formatScientific(result.productivity);
    app.querySelector("[data-result-temperature]").textContent = formatNumber(result.temperature, 2);
    app.querySelector("[data-result-viscosity]").textContent = formatNumber(result.viscosity, 2);

    if (Array.isArray(result.profile) && result.profile.length > 1) {
      renderChart(app.querySelector("[data-temperature-chart]"), result.profile, "temperature", "#0f766e", "T, °C");
      renderChart(app.querySelector("[data-viscosity-chart]"), result.profile, "viscosity", "#334155", "η, Па·с");
      chartArea.hidden = false;
    }

    resultSummary.hidden = false;
    reportPanel.hidden = false;
    reportStatus.textContent = "Расчет выполнен. Отчет может быть сформирован.";
    downloadReportButton.disabled = true;
    showStep(4);
  }

  async function loadMaterials() {
    try {
      const materials = await api.materials.list();
      materialSelect.innerHTML = '<option value="">Выберите материал</option>';

      materials.forEach(function (material) {
        const option = document.createElement("option");
        option.value = material.id;
        option.textContent = material.name;
        materialSelect.appendChild(option);
      });

      materialStatus.textContent = materials.length ? "" : "Справочник материалов пуст.";
    } catch (error) {
      materialSelect.innerHTML = '<option value="">Материалы недоступны</option>';
      materialStatus.textContent = error.message || "Не удалось загрузить материалы.";
    }
  }

  prevButton.addEventListener("click", function () {
    setMessage("");
    showStep(activeStep - 1);
  });

  nextButton.addEventListener("click", function () {
    setMessage("");
    if (validateCurrentStep()) {
      showStep(activeStep + 1);
    }
  });

  form.addEventListener("submit", async function (event) {
    event.preventDefault();
    setMessage("");

    if (!validateCurrentStep()) {
      return;
    }

    const payload = getPayload();
    setBusy(true);
    submitButton.textContent = "Выполняется расчет...";

    try {
      await api.calculation.validate(payload);
      const result = await api.calculation.calculate(payload);
      showResult(result);
      setMessage("Расчет успешно выполнен.", "success");
    } catch (error) {
      setMessage(error.message || "Расчет не выполнен.", "error");
    } finally {
      submitButton.textContent = "Выполнить расчет";
      setBusy(false);
    }
  });

  generateReportButton.addEventListener("click", async function () {
    if (!calculationResult || !calculationResult.id) {
      setMessage("Идентификатор расчета отсутствует.", "error");
      return;
    }

    setBusy(true);
    reportStatus.textContent = "Формирование отчета...";

    try {
      reportData = await api.results.report(calculationResult.id);
      reportReady = Boolean(reportData);
      reportStatus.textContent = "Отчет сформирован. Доступно скачивание JSON-файла.";
      downloadReportButton.disabled = false;
      setMessage("");
    } catch (error) {
      reportStatus.textContent = "Отчет не сформирован.";
      setMessage(error.message || "Не удалось сформировать отчет.", "error");
    } finally {
      setBusy(false);
    }
  });

  downloadReportButton.addEventListener("click", async function () {
    if (!reportReady || !calculationResult || !calculationResult.id) {
      return;
    }

    setBusy(true);

    try {
      const reportDownload = await api.results.download(calculationResult.id);
      const url = URL.createObjectURL(reportDownload.blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = reportDownload.filename;
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
      setMessage("");
    } catch (error) {
      setMessage(error.message || "Не удалось скачать отчет.", "error");
    } finally {
      setBusy(false);
    }
  });

  showStep(0);
  loadMaterials();
})();
