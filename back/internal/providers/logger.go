package providers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/djsilvajr/simple-project-template/back/internal/config"
	"github.com/djsilvajr/simple-project-template/back/internal/middleware"
)

const maxLoggedBody = 4 << 10 // 4 KB

// Campos com esses nomes têm o valor mascarado no log do body.
var sensitiveKeys = map[string]bool{
	"password": true,
	"senha":    true,
	"token":    true,
	"secret":   true,
}

// Logger registra o início (método, url, path, ip e, opcionalmente, o body)
// e o fim (status e tempo de execução) de toda requisição.
type Logger struct {
	logBody    bool
	trustProxy bool
}

func (l *Logger) Register(cfg *config.Config) error {
	l.logBody = cfg.App.LogRequestBody
	l.trustProxy = cfg.App.TrustProxy
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	return nil
}

func (l *Logger) Boot(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := newRequestID()
		w.Header().Set("X-Request-ID", requestID)

		// --- início do fluxo ---
		log.Printf("[%s] --> %s url=%s path=%s ip=%s%s",
			requestID,
			r.Method,
			r.URL.String(),
			r.URL.Path,
			middleware.ClientIP(r, l.trustProxy),
			l.body(r),
		)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		// --- fim do fluxo ---
		log.Printf("[%s] <-- %s %s status=%d duration=%s",
			requestID,
			r.Method,
			r.URL.Path,
			rec.status,
			time.Since(start),
		)
	})
}

// body lê o início do corpo para o log e devolve o corpo intacto ao request.
func (l *Logger) body(r *http.Request) string {
	if !l.logBody || r.Body == nil || r.Body == http.NoBody {
		return ""
	}

	buf, err := io.ReadAll(io.LimitReader(r.Body, maxLoggedBody))
	r.Body = struct {
		io.Reader
		io.Closer
	}{io.MultiReader(bytes.NewReader(buf), r.Body), r.Body}

	if err != nil || len(buf) == 0 {
		return ""
	}

	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		var data any
		if json.Unmarshal(buf, &data) == nil {
			masked, _ := json.Marshal(maskSensitive(data))
			return " body=" + string(masked)
		}
	}

	return " body=[" + r.Header.Get("Content-Type") + ", " + strconv.Itoa(len(buf)) + " bytes]"
}

func maskSensitive(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k, item := range val {
			if sensitiveKeys[strings.ToLower(k)] {
				val[k] = "***"
			} else {
				val[k] = maskSensitive(item)
			}
		}
	case []any:
		for i, item := range val {
			val[i] = maskSensitive(item)
		}
	}
	return v
}

func newRequestID() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "000000000000"
	}
	return hex.EncodeToString(b)
}

// statusRecorder captura o status HTTP escrito pelo handler.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wroteHeader = true
	return r.ResponseWriter.Write(b)
}

// Unwrap permite que http.ResponseController acesse o writer original.
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
