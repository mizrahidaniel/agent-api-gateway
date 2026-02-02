package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port     int                 `yaml:"port"`
	Services map[string]*Service `yaml:"services"`
}

type Service struct {
	Target   string            `yaml:"target"`
	Auth     *AuthConfig       `yaml:"auth,omitempty"`
	RateLimit *RateLimitConfig `yaml:"rate_limit,omitempty"`
	proxy    *httputil.ReverseProxy
}

type AuthConfig struct {
	Type   string   `yaml:"type"` // bearer, apikey
	Tokens []string `yaml:"tokens"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
}

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
	}
}

func (rl *rateLimiter) allow(key string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Minute)

	// Clean old requests
	reqs := rl.requests[key]
	filtered := make([]time.Time, 0)
	for _, t := range reqs {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= limit {
		return false
	}

	filtered = append(filtered, now)
	rl.requests[key] = filtered
	return true
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Initialize reverse proxies
	for name, svc := range cfg.Services {
		target, err := url.Parse(svc.Target)
		if err != nil {
			return nil, fmt.Errorf("invalid target URL for %s: %w", name, err)
		}
		svc.proxy = httputil.NewSingleHostReverseProxy(target)
	}

	return &cfg, nil
}

func (c *Config) authenticate(svc *Service, r *http.Request) bool {
	if svc.Auth == nil {
		return true
	}

	switch svc.Auth.Type {
	case "bearer":
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return false
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		for _, validToken := range svc.Auth.Tokens {
			if token == validToken {
				return true
			}
		}
		return false

	case "apikey":
		key := r.Header.Get("X-API-Key")
		for _, validKey := range svc.Auth.Tokens {
			if key == validKey {
				return true
			}
		}
		return false

	default:
		return true
	}
}

func (c *Config) handler(limiter *rateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract service name from path: /service-name/path
		parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 2)
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "Service not specified", http.StatusBadRequest)
			return
		}

		serviceName := parts[0]
		svc, ok := c.Services[serviceName]
		if !ok {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}

		// Authentication
		if !c.authenticate(svc, r) {
			w.Header().Set("WWW-Authenticate", `Bearer realm="gateway"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Rate limiting
		if svc.RateLimit != nil {
			clientIP := r.RemoteAddr
			key := fmt.Sprintf("%s:%s", serviceName, clientIP)
			if !limiter.allow(key, svc.RateLimit.RequestsPerMinute) {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", svc.RateLimit.RequestsPerMinute))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		}

		// Rewrite path to remove service prefix
		if len(parts) > 1 {
			r.URL.Path = "/" + parts[1]
		} else {
			r.URL.Path = "/"
		}

		// Proxy request
		log.Printf("[%s] %s %s -> %s%s", serviceName, r.Method, r.RemoteAddr, svc.Target, r.URL.Path)
		svc.proxy.ServeHTTP(w, r)
	}
}

func main() {
	configPath := "gateway.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Port == 0 {
		cfg.Port = 8080
	}

	limiter := newRateLimiter()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: cfg.handler(limiter),
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Agent API Gateway listening on :%d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}

	log.Println("Gateway stopped")
}
