package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Хранит инфу о попытках входа
type LoginAttempt struct {
	Count        int
	FirstTry     time.Time
	BlockedUntil time.Time
}

// Отслеживает попытки входа
type RateLimiter struct {
	mu          sync.RWMutex
	attempts    map[string]*LoginAttempt
	maxAttempts int           // максимальное количество попыток
	window      time.Duration // окно для подсчёта попыток
	blockTime   time.Duration // время блокировки после превышения
}

func NewRateLimiter(maxAttempts int, window, blockTime time.Duration) *RateLimiter {
	rl := &RateLimiter{
		attempts:    make(map[string]*LoginAttempt),
		maxAttempts: maxAttempts,
		window:      window,
		blockTime:   blockTime,
	}

	// Фоновая очистка старых записей (экономим память)
	go rl.cleanup()

	return rl
}

// Периодически удаляет устаревшие записи
func (rl *RateLimiter) cleanup() {
	// Интервал очистки привязываем к window, но с ограничениями
	// Если window маленький - чистим раз в минуту
	// Если window большой - чистим раз в час
	interval := rl.window
	if interval < time.Minute {
		interval = time.Minute
	}
	if interval > time.Hour {
		interval = time.Hour
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, attempt := range rl.attempts {
			if !attempt.BlockedUntil.IsZero() && now.After(attempt.BlockedUntil) {
				if now.Sub(attempt.BlockedUntil) > rl.blockTime {
					delete(rl.attempts, ip)
					continue
				}
			}
			if now.Sub(attempt.FirstTry) > rl.window+rl.blockTime {
				delete(rl.attempts, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Проверяет, можно ли попытаться войти с этого IP
func (rl *RateLimiter) Allow(ip string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	attempt, exists := rl.attempts[ip]
	if !exists {
		// Первая попытка - создаём запись
		rl.attempts[ip] = &LoginAttempt{
			Count:    1,
			FirstTry: now,
		}
		return true, 0
	}

	// Проверяем, не в блоке IP
	if !attempt.BlockedUntil.IsZero() {
		if now.Before(attempt.BlockedUntil) {
			remaining := attempt.BlockedUntil.Sub(now)
			return false, remaining
		}
		// Блок истёк - сбрасываем счётчик
		attempt.BlockedUntil = time.Time{}
		attempt.Count = 1
		attempt.FirstTry = now
		return true, 0
	}

	// Если с первой попытки прошло больше window - сбрасываем
	if now.Sub(attempt.FirstTry) > rl.window {
		attempt.Count = 1
		attempt.FirstTry = now
		return true, 0
	}

	attempt.Count++

	// Превысили лимит - блок
	if attempt.Count > rl.maxAttempts {
		attempt.BlockedUntil = now.Add(rl.blockTime)
		return false, rl.blockTime
	}

	return true, 0
}

func (rl *RateLimiter) Reset(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, ip)
}

// Получает IP клиента из запроса
// Прям до конца не разобрался с этими проксями, но должно работать :)
func getIP(r *http.Request) string {
	// X-Forwarded-For: "клиент, прокси1, прокси2" - берём первый
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// X-Real-IP - если nginx настроен
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Напрямую - RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// Если SplitHostPort не сработал - возвращаем как есть
		return r.RemoteAddr
	}
	return ip
}

func GetIP(r *http.Request) string {
	return getIP(r)
}

// Обёртка для логина, чтобы брутфорсить не могли
func LoginRateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Только на POST /api/auth/login
			if r.Method == http.MethodPost && r.URL.Path == "/api/auth/login" {
				ip := getIP(r)

				allowed, remaining := rl.Allow(ip)
				if !allowed {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusTooManyRequests)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error":   "rate_limit",
						"message": "Слишком много попыток входа. Попробуйте позже.",
						"details": map[string]interface{}{
							"blocked_for_seconds": int(remaining.Seconds()),
						},
					})
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
