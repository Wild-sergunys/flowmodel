(function () {
  "use strict";

  const DEFAULT_CREDENTIALS = "same-origin";
  const AUTH_STORAGE_KEY = "flowmodel:auth";

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
    if (!baseURL) {
      return cleanPath;
    }

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
    if (!disposition) {
      return "";
    }

    const match = disposition.match(/filename="?([^"]+)"?/i);
    return match ? match[1] : "";
  }

  function readAuthState() {
    try {
      const raw = window.localStorage.getItem(AUTH_STORAGE_KEY);
      return raw ? JSON.parse(raw) : null;
    } catch (error) {
      return null;
    }
  }

  function writeAuthState(authState) {
    window.localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(authState));
  }

  function clearAuthState() {
    window.localStorage.removeItem(AUTH_STORAGE_KEY);
  }

  function getToken() {
    const authState = readAuthState();
    return authState && authState.token ? authState.token : "";
  }

  function saveLoginResponse(data) {
    if (!data || typeof data.token !== "string" || !data.token) {
      return;
    }

    writeAuthState({
      token: data.token,
      role: data.role || "",
    });
  }

  async function parseResponse(response, responseType) {
    if (response.status === 204) {
      return null;
    }
    if (responseType === "blob") {
      return response.blob();
    }
    if (responseType === "text") {
      return response.text();
    }

    const contentType = response.headers.get("Content-Type") || "";
    if (contentType.includes("application/json")) {
      return response.json();
    }

    const text = await response.text();
    return text || null;
  }

  function getErrorMessage(data, response) {
    if (isPlainObject(data) && typeof data.message === "string" && data.message) {
      return data.message;
    }
    if (typeof data === "string" && data) {
      return data;
    }

    return response.statusText || "API request failed";
  }

  function createError(response, data) {
    const message = getErrorMessage(data, response);
    if (response.status === 401) {
      return new AuthError(message, response, data);
    }
    if (response.status === 403) {
      return new ForbiddenError(message, response, data);
    }

    return new ApiError(message, response, data);
  }

  function dispatchAuthEvent(name, error) {
    window.dispatchEvent(
      new CustomEvent(name, {
        detail: {
          error: error,
          status: error.status,
          message: error.message,
          data: error.data,
        },
      }),
    );
  }

  function createClient(options) {
    const settings = Object.assign(
      {
        baseURL: "",
        credentials: DEFAULT_CREDENTIALS,
        onUnauthorized: undefined,
        onForbidden: undefined,
      },
      options || {},
    );

    async function send(path, options) {
      const requestOptions = Object.assign(
        {
          method: "GET",
          headers: undefined,
          body: undefined,
          responseType: "json",
          auth: true,
        },
        options || {},
      );

      const headers = buildHeaders(requestOptions.headers, requestOptions.body);
      const token = getToken();
      if (requestOptions.auth && token && !headers.has("Authorization")) {
        headers.set("Authorization", "Bearer " + token);
      }

      const fetchOptions = {
        method: requestOptions.method,
        credentials: settings.credentials,
        headers: headers,
      };

      if (requestOptions.body !== undefined) {
        fetchOptions.body =
          requestOptions.body instanceof FormData ? requestOptions.body : JSON.stringify(requestOptions.body);
      }

      const response = await fetch(joinURL(settings.baseURL, path), fetchOptions);
      const data = await parseResponse(response, response.ok ? requestOptions.responseType : "json");
      if (!response.ok) {
        const error = createError(response, data);
        if (error instanceof AuthError) {
          clearAuthState();
          dispatchAuthEvent("flowmodel:unauthorized", error);
          if (typeof settings.onUnauthorized === "function") {
            settings.onUnauthorized(error);
          }
        }
        if (error instanceof ForbiddenError) {
          dispatchAuthEvent("flowmodel:forbidden", error);
          if (typeof settings.onForbidden === "function") {
            settings.onForbidden(error);
          }
        }

        throw error;
      }

      return {
        data: data,
        response: response,
      };
    }

    async function request(path, options) {
      const result = await send(path, options);
      return result.data;
    }

    function get(path, options) {
      return request(path, Object.assign({}, options, { method: "GET" }));
    }

    function post(path, body, options) {
      return request(path, Object.assign({}, options, { method: "POST", body }));
    }

    function put(path, body, options) {
      return request(path, Object.assign({}, options, { method: "PUT", body }));
    }

    function del(path, options) {
      return request(path, Object.assign({}, options, { method: "DELETE" }));
    }

    return {
      request,

      auth: {
        register: function (payload) {
          return post("/api/auth/register", payload, { auth: false });
        },
        login: async function (payload) {
          const data = await post("/api/auth/login", payload, { auth: false });
          saveLoginResponse(data);
          return data;
        },
        logout: async function () {
          try {
            return await post("/api/auth/logout");
          } finally {
            clearAuthState();
          }
        },
        me: function () {
          return get("/api/auth/me");
        },
        getState: function () {
          return readAuthState();
        },
        getToken: function () {
          return getToken();
        },
        isAuthenticated: function () {
          return Boolean(getToken());
        },
        clear: function () {
          clearAuthState();
        },
      },

      materials: {
        list: function () {
          return get("/api/materials");
        },
      },

      calculation: {
        validate: function (payload) {
          return post("/api/validate", payload);
        },
        calculate: function (payload) {
          return post("/api/calculate", payload);
        },
      },

      results: {
        list: function () {
          return get("/api/results");
        },
        get: function (id) {
          return get("/api/results/" + encodePathPart(id));
        },
        report: function (id) {
          return get("/api/results/" + encodePathPart(id) + "/report");
        },
        download: async function (id) {
          const result = await send("/api/results/" + encodePathPart(id) + "/download", {
            method: "GET",
            responseType: "blob",
          });

          return {
            blob: result.data,
            filename: getFilename(result.response) || "report_" + encodePathPart(id) + ".json",
          };
        },
      },

      admin: {
        materials: {
          list: function () {
            return get("/api/admin/materials");
          },
          get: function (id) {
            return get("/api/admin/materials/" + encodePathPart(id));
          },
          create: function (payload) {
            return post("/api/admin/materials", payload);
          },
          update: function (id, payload) {
            return put("/api/admin/materials/" + encodePathPart(id), payload);
          },
          remove: function (id) {
            return del("/api/admin/materials/" + encodePathPart(id));
          },
          parameters: {
            list: function (materialId) {
              return get("/api/admin/materials/" + encodePathPart(materialId) + "/parameters");
            },
            update: function (materialId, payload) {
              return put("/api/admin/materials/" + encodePathPart(materialId) + "/parameters", payload);
            },
          },
        },

        parameters: {
          list: function () {
            return get("/api/admin/parameters");
          },
          get: function (id) {
            return get("/api/admin/parameters/" + encodePathPart(id));
          },
          create: function (payload) {
            return post("/api/admin/parameters", payload);
          },
          update: function (id, payload) {
            return put("/api/admin/parameters/" + encodePathPart(id), payload);
          },
          remove: function (id) {
            return del("/api/admin/parameters/" + encodePathPart(id));
          },
        },

        users: {
          list: function () {
            return get("/api/admin/users");
          },
          get: function (id) {
            return get("/api/admin/users/" + encodePathPart(id));
          },
          create: function (payload) {
            return post("/api/admin/users", payload);
          },
          update: function (id, payload) {
            return put("/api/admin/users/" + encodePathPart(id), payload);
          },
          remove: function (id) {
            return del("/api/admin/users/" + encodePathPart(id));
          },
        },
      },
    };
  }

  window.FlowModelAPI = {
    ApiError: ApiError,
    AuthError: AuthError,
    ForbiddenError: ForbiddenError,
    createClient: createClient,
    client: createClient(),
    utils: {
      trimSlashes: trimSlashes,
      getFilename: getFilename,
    },
  };
})();
