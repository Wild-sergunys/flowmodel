package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewHandler(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	if handler == nil {
		t.Fatal("NewHandler() вернул nil")
	}
}

func TestPages(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	tests := []struct {
		name       string
		handler    func(w http.ResponseWriter, r *http.Request)
		wantStatus int
		wantInBody string
	}{
		{
			name:       "главная страница",
			handler:    handler.Home,
			wantStatus: http.StatusOK,
			wantInBody: "FlowModel",
		},
		{
			name:       "страница входа",
			handler:    handler.Login,
			wantStatus: http.StatusOK,
			wantInBody: "Вход - FlowModel",
		},
		{
			name:       "админ-панель",
			handler:    handler.Admin,
			wantStatus: http.StatusOK,
			wantInBody: "Админ-панель - FlowModel",
		},
		{
			name:       "кабинет",
			handler:    handler.Cabinet,
			wantStatus: http.StatusOK,
			wantInBody: "Кабинет - FlowModel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			response := httptest.NewRecorder()

			tt.handler(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("статус = %d, want %d", response.Code, tt.wantStatus)
			}

			body := response.Body.String()
			if !strings.Contains(body, tt.wantInBody) {
				t.Errorf("тело ответа не содержит %q", tt.wantInBody)
			}
		})
	}
}

func TestHomePath(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	tests := []struct {
		name       string
		path       string
		wantStatus int
		want404    bool
	}{
		{
			name:       "корень",
			path:       "/",
			wantStatus: http.StatusOK,
			want404:    false,
		},
		{
			name:       "несуществующий путь",
			path:       "/notexistent",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()

			handler.Home(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("статус = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}

func TestStaticFiles(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantInBody string
	}{
		{
			name:       "CSS файл",
			path:       "/static/css/app.css",
			wantStatus: http.StatusOK,
		},
		{
			name:       "JS файл",
			path:       "/static/js/api.js",
			wantStatus: http.StatusOK,
			wantInBody: "window.FlowModelAPI",
		},
		{
			name:       "несуществующий статический файл",
			path:       "/static/notexistent.txt",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()

			handler.Static().ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("статус = %d, want %d", response.Code, tt.wantStatus)
			}

			if tt.wantInBody != "" {
				body := response.Body.String()
				if !strings.Contains(body, tt.wantInBody) {
					t.Errorf("тело ответа не содержит %q", tt.wantInBody)
				}
			}
		})
	}
}

func TestContentType(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	tests := []struct {
		name            string
		path            string
		wantContentType string
	}{
		{
			name:            "HTML страница",
			path:            "/",
			wantContentType: "text/html; charset=utf-8",
		},
		{
			name:            "CSS файл",
			path:            "/static/css/app.css",
			wantContentType: "text/css",
		},
		{
			name:            "JS файл",
			path:            "/static/js/api.js",
			wantContentType: "application/javascript",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()

			if tt.path == "/" {
				handler.Home(response, request)
			} else {
				handler.Static().ServeHTTP(response, request)
			}

			contentType := response.Header().Get("Content-Type")
			if !strings.Contains(contentType, tt.wantContentType) {
				t.Errorf("Content-Type = %q, want содержит %q", contentType, tt.wantContentType)
			}
		})
	}
}
