(function () {
  "use strict";

  const DEFAULT_CREDENTIALS = "same-origin";

  class ApiError extends Error {
    constructor(message, response, data) {
      super(message);
      this.name = "ApiError";
      this.status = response.status;
      this.statusText = response.statusText;
      this.data = data;
    }
  }

  class AuthError extends ApiError {
    constructor(message, response, data) {
      super(message, response, data);
      this.name = "AuthError";
    }
  }

  class ForbiddenError extends ApiError {
    constructor(message, response, data) {
      super(message, response, data);
      this.name = "ForbiddenError";
    }
  }

  function trimSlashes(value) {
    return value.replace(/^\/+|\/+$/g, "");
  }

  function joinURL(baseURL, path) {
    const cleanPath = path.startsWith("/") ? path : "/" + path;
    if (!baseURL) return cleanPath;
    return baseURL.replace(/\/+$/g, "") + cleanPath;
  }

  function isPlainObject(value) {
    return Object.prototype.toString.call(value) === "[object Object]";
  }

  function buildHeaders(headers, body) {
    const nextHeaders = new Headers(headers || {});
    if (body !== undefined && !(body instanceof FormData) && !nextHeaders.has("Content-Type")) {
      nextHeaders.set("Content-Type", "application/json");
    }
    if (!nextHeaders.has("Accept")) {
      nextHeaders.set("Accept", "application/json");
    }
    return nextHeaders;
  }

  function encodePathPart(value) {
    return encodeURIComponent(String(value));
  }

  function getFilename(response) {
    const disposition = response.headers.get("Content-Disposition");
    if (!disposition) return "";
    const match = disposition.match(/filename="?([^"]+)"?/i);
    return match ? match[1] : "";
  }

  async function parseResponse(response, responseType) {
    if (response.status === 204) return null;
    if (responseType === "blob") return response.blob();
    if (responseType === "text") return response.text();
    const contentType = response.headers.get("Content-Type") || "";
    if (contentType.includes("application/json")) return response.json();
    const text = await response.text();
    return text || null;
  }

  function getErrorMessage(data, response) {
    if (isPlainObject(data) && typeof data.message === "string" && data.message) return data.message;
    if (typeof data === "string" && data) return data;
    return response.statusText || "API request failed";
  }

  function createError(response, data) {
    const message = getErrorMessage(data, response);
    if (response.status === 401) return new AuthError(message, response, data);
    if (response.status === 403) return new ForbiddenError(message, response, data);
    return new ApiError(message, response, data);
  }

  function dispatchAuthEvent(name, error) {
    window.dispatchEvent(new CustomEvent(name, { detail: { error, status: error.status, message: error.message, data: error.data } }));
  }

  function createClient(options) {
    const settings = Object.assign({ baseURL: "", credentials: DEFAULT_CREDENTIALS, onUnauthorized: undefined, onForbidden: undefined }, options || {});

    async function send(path, options) {
      const requestOptions = Object.assign({ method: "GET", headers: undefined, body: undefined, responseType: "json" }, options || {});
      const fetchOptions = {
        method: requestOptions.method,
        credentials: settings.credentials,
        headers: buildHeaders(requestOptions.headers, requestOptions.body),
      };

      const token = sessionStorage.getItem('flowmodel_token');
      if (token) {
        fetchOptions.headers.set('Authorization', 'Bearer ' + token);
      }

      if (requestOptions.body !== undefined) {
        fetchOptions.body = requestOptions.body instanceof FormData ? requestOptions.body : JSON.stringify(requestOptions.body);
      }

      const response = await fetch(joinURL(settings.baseURL, path), fetchOptions);
      const data = await parseResponse(response, response.ok ? requestOptions.responseType : "json");

      if (!response.ok) {
        const error = createError(response, data);
        if (error instanceof AuthError) {
          dispatchAuthEvent("flowmodel:unauthorized", error);
          if (typeof settings.onUnauthorized === "function") settings.onUnauthorized(error);
        }
        if (error instanceof ForbiddenError) {
          dispatchAuthEvent("flowmodel:forbidden", error);
          if (typeof settings.onForbidden === "function") settings.onForbidden(error);
        }
        throw error;
      }

      return { data, response };
    }

    async function request(path, options) {
      const result = await send(path, options);
      return result.data;
    }

    function get(path, options) { return request(path, Object.assign({}, options, { method: "GET" })); }
    function post(path, body, options) { return request(path, Object.assign({}, options, { method: "POST", body })); }
    function put(path, body, options) { return request(path, Object.assign({}, options, { method: "PUT", body })); }
    function del(path, options) { return request(path, Object.assign({}, options, { method: "DELETE" })); }

    return {
      request,
      auth: {
        register: (payload) => post("/api/auth/register", payload),
        login: (payload) => post("/api/auth/login", payload),
        logout: () => post("/api/auth/logout"),
        me: () => get("/api/auth/me"),
      },
      materials: { list: () => get("/api/materials") },
      calculation: {
        validate: (payload) => post("/api/validate", payload),
        calculate: (payload) => post("/api/calculate", payload),
      },
      results: {
        list: () => get("/api/results"),
        get: (id) => get("/api/results/" + encodePathPart(id)),
        report: (id) => get("/api/results/" + encodePathPart(id) + "/report"),
        download: async (id) => {
          const result = await send("/api/results/" + encodePathPart(id) + "/download", 
          { method: "GET", responseType: "blob" });
        return { blob: result.data, filename: getFilename(result.response) || "report_" + encodePathPart(id)};
        },
      },
      admin: {
        materials: {
          list: () => get("/api/admin/materials"),
          get: (id) => get("/api/admin/materials/" + encodePathPart(id)),
          create: (payload) => post("/api/admin/materials", payload),
          update: (id, payload) => put("/api/admin/materials/" + encodePathPart(id), payload),
          remove: (id) => del("/api/admin/materials/" + encodePathPart(id)),
          parameters: {
            list: (materialId) => get("/api/admin/materials/" + encodePathPart(materialId) + "/parameters"),
            update: (materialId, payload) => put("/api/admin/materials/" + encodePathPart(materialId) + "/parameters", payload),
          },
        },
        parameters: {
          list: () => get("/api/admin/parameters"),
          get: (id) => get("/api/admin/parameters/" + encodePathPart(id)),
          create: (payload) => post("/api/admin/parameters", payload),
          update: (id, payload) => put("/api/admin/parameters/" + encodePathPart(id), payload),
          remove: (id) => del("/api/admin/parameters/" + encodePathPart(id)),
        },
        users: {
          list: () => get("/api/admin/users"),
          get: (id) => get("/api/admin/users/" + encodePathPart(id)),
          create: (payload) => post("/api/admin/users", payload),
          update: (id, payload) => put("/api/admin/users/" + encodePathPart(id), payload),
          remove: (id) => del("/api/admin/users/" + encodePathPart(id)),
        },
      },
    };
  }

  window.FlowModelAPI = {
    ApiError, AuthError, ForbiddenError,
    createClient,
    client: createClient(),
    utils: { trimSlashes, getFilename },
  };
})();