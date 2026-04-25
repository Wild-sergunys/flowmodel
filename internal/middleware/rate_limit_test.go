package middleware

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(5, 15*time.Minute, 15*time.Minute)
	ip := "192.168.1.1"

	tests := []struct {
		name   string
		action func() (bool, time.Duration)
		wantOk bool
	}{
		// 5 успешных попыток
		{"попытка 1", func() (bool, time.Duration) { return rl.Allow(ip) }, true},
		{"попытка 2", func() (bool, time.Duration) { return rl.Allow(ip) }, true},
		{"попытка 3", func() (bool, time.Duration) { return rl.Allow(ip) }, true},
		{"попытка 4", func() (bool, time.Duration) { return rl.Allow(ip) }, true},
		{"попытка 5", func() (bool, time.Duration) { return rl.Allow(ip) }, true},
		// 6-я — уже блок
		{"попытка 6 - блок", func() (bool, time.Duration) { return rl.Allow(ip) }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := tt.action()
			if ok != tt.wantOk {
				t.Errorf("Allow() = %v, want %v", ok, tt.wantOk)
			}
		})
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(3, 15*time.Minute, 15*time.Minute)
	ip := "10.0.0.1"

	// Исчерпали попытки
	for i := 0; i < 3; i++ {
		rl.Allow(ip)
	}
	// 4-я - блок
	if ok, _ := rl.Allow(ip); ok {
		t.Fatal("ожидался блок после 3 попыток")
	}

	// Сброс
	rl.Reset(ip)

	// Снова должно пускать
	if ok, _ := rl.Allow(ip); !ok {
		t.Fatal("после Reset должно пускать")
	}
}

func TestRateLimiter_ExpiredBlock(t *testing.T) {
	// Маленькое окно блокировки для теста
	rl := NewRateLimiter(2, 50*time.Millisecond, 50*time.Millisecond)
	ip := "10.0.0.2"

	// 2 попытки - ок
	rl.Allow(ip)
	rl.Allow(ip)
	// 3-я - блок
	if ok, _ := rl.Allow(ip); ok {
		t.Fatal("ожидался блок")
	}

	// Ждём истечения блокировки
	time.Sleep(100 * time.Millisecond)

	// После истечения - снова пускает
	if ok, _ := rl.Allow(ip); !ok {
		t.Fatal("после истечения блокировки должно пускать")
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(2, 15*time.Minute, 15*time.Minute)

	// Исчерпали попытки для IP1
	rl.Allow("192.168.1.1")
	rl.Allow("192.168.1.1")
	if ok, _ := rl.Allow("192.168.1.1"); ok {
		t.Fatal("IP1 должен быть заблокирован")
	}

	// IP2 - всё ещё может входить
	if ok, _ := rl.Allow("192.168.1.2"); !ok {
		t.Fatal("IP2 должен проходить")
	}
}
