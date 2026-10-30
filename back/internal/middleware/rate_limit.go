package middleware

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/djsilvajr/simple-project-template/back/internal/config"
	"github.com/djsilvajr/simple-project-template/back/internal/response"
)

const (
	cleanupInterval = time.Minute
	visitorTTL      = 3 * time.Minute
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter limita requisições por IP usando token bucket em memória.
// Cada instância tem seus próprios contadores, então é possível criar
// limites diferentes por grupo de rotas (ex.: um mais restrito no login).
//
// Obs.: por ser em memória, o limite vale por instância da API. Para várias
// réplicas, troque o mapa por um contador no Redis.
type RateLimiter struct {
	mu         sync.Mutex
	visitors   map[string]*visitor
	limit      rate.Limit
	burst      int
	trustProxy bool
}

func NewRateLimiter(requestsPerSecond, burst int, trustProxy bool) *RateLimiter {
	requestsPerSecond = max(requestsPerSecond, 1)
	burst = max(burst, 1)

	rl := &RateLimiter{
		visitors:   make(map[string]*visitor),
		limit:      rate.Limit(requestsPerSecond),
		burst:      burst,
		trustProxy: trustProxy,
	}
	go rl.cleanup()
	return rl
}

// RateLimit cria o middleware a partir do config. Quando desabilitado,
// retorna um middleware que apenas repassa a requisição.
func RateLimit(cfg *config.Config) Middleware {
	if !cfg.RateLimit.Enabled {
		return func(next http.Handler) http.Handler { return next }
	}
	return NewRateLimiter(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst, cfg.App.TrustProxy).Middleware
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := rl.get(ClientIP(r, rl.trustProxy))

		if !limiter.Allow() {
			// Tempo (em segundos, arredondado para cima) até liberar 1 token.
			retryAfter := int(math.Ceil(1 / float64(rl.limit)))
			w.Header().Set("Retry-After", strconv.Itoa(max(retryAfter, 1)))
			response.Error(w, http.StatusTooManyRequests, "Muitas requisições. Tente novamente em instantes.")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) get(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(rl.limit, rl.burst)}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > visitorTTL {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// ClientIP devolve o IP do cliente. Com trustProxy, usa X-Real-IP /
// X-Forwarded-For (definidos pelo nginx); sem ele, usa o RemoteAddr.
func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
			return ip
		}
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			first, _, _ := strings.Cut(fwd, ",")
			return strings.TrimSpace(first)
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
