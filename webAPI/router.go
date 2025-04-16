package webAPI

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// NotFoundHandler gère les requêtes pour les pages non trouvées
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	t, err := template.ParseFiles("public/HTML/404.html")
	if err != nil {
		http.Error(w, "Page non trouvée", http.StatusNotFound)
		return
	}
	t.Execute(w, nil)
}

// Debug est un indicateur global pour activer/désactiver les logs de débogage
var Debug = true

// DebugPrintf imprime les logs de débogage conditionnellement
func DebugPrintf(format string, a ...interface{}) {
	if Debug {
		fmt.Printf("[DEBUG] "+format+"\n", a...)
	}
}

// RateLimiter gère le nombre de requêtes par IP
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	limit    int           // Nombre max de requêtes
	window   time.Duration // Période de temps
}

// NewRateLimiter crée un nouveau limiteur de requêtes
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	DebugPrintf("Création d'un nouveau RateLimiter : %d requêtes par %v", limit, window)
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow vérifie si une requête est autorisée
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	// Nettoyer les requêtes expirées
	requests := rl.requests[ip]
	var cleaned []time.Time

	for _, req := range requests {
		if now.Sub(req) <= rl.window {
			cleaned = append(cleaned, req)
		}
	}

	// Vérifier si le nombre de requêtes dépasse la limite
	if len(cleaned) >= rl.limit {
		DebugPrintf("🚫 RATE LIMIT: IP %s BLOQUÉE - %d requêtes dans la fenêtre", ip, len(cleaned))
		return false
	}

	// Ajouter la nouvelle requête
	rl.requests[ip] = append(cleaned, now)
	DebugPrintf("✅ RATE LIMIT: IP %s autorisée - %d requêtes actuelles", ip, len(rl.requests[ip]))
	return true
}

// CustomRouter est un routeur personnalisé qui gère les erreurs 404 et le rate limiting
type CustomRouter struct {
	routes      map[string]http.HandlerFunc
	static      http.Handler
	rateLimiter *RateLimiter
}

// NewCustomRouter crée un nouveau routeur personnalisé avec rate limiting
func NewCustomRouter() *CustomRouter {
	DebugPrintf("Création d'un nouveau CustomRouter avec RateLimiter")
	return &CustomRouter{
		routes:      make(map[string]http.HandlerFunc),
		static:      http.FileServer(http.Dir("public")),
		rateLimiter: NewRateLimiter(150, 1*time.Minute), // 150 requêtes par minute
	}
}

// getClientIP extrait l'adresse IP du client de manière robuste
func getClientIP(r *http.Request) string {
	// Vérifier d'abord les en-têtes de proxy
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		DebugPrintf("IP via X-Forwarded-For: %s", ip)
		return ip
	}

	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		DebugPrintf("IP via X-Real-IP: %s", ip)
		return ip
	}

	// Récupérer l'IP à partir de RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		DebugPrintf("Erreur lors de la récupération de l'IP : %v", err)
		ip = r.RemoteAddr
	}

	DebugPrintf("IP via RemoteAddr: %s", ip)
	return ip
}

// HandleFunc ajoute une route au routeur
func (r *CustomRouter) HandleFunc(path string, handler http.HandlerFunc) {
	DebugPrintf("Ajout d'une route : %s", path)
	r.routes[path] = func(w http.ResponseWriter, req *http.Request) {
		// Log de débogage pour chaque requête
		DebugPrintf("Requête reçue : Path=%s, Method=%s", path, req.Method)

		// Récupérer l'IP du client
		ip := getClientIP(req)
		
		// Vérifier la limite de requêtes
		if !r.rateLimiter.Allow(ip) {
			DebugPrintf("🚨 BLOQUÉ: Requête de %s rejetée", ip)
			http.Error(w, "Trop de requêtes. Veuillez réessayer plus tard.", http.StatusTooManyRequests)
			return
		}
		
		// Appeler le gestionnaire original
		handler(w, req)
	}
}

// Handle ajoute un gestionnaire au routeur
func (r *CustomRouter) Handle(path string, handler http.Handler) {
	DebugPrintf("Ajout d'un gestionnaire : %s", path)
	r.routes[path] = func(w http.ResponseWriter, req *http.Request) {
		// Log de débogage pour chaque requête
		DebugPrintf("Requête reçue : Path=%s, Method=%s", path, req.Method)

		// Récupérer l'IP du client
		ip := getClientIP(req)
		
		// Vérifier la limite de requêtes
		if !r.rateLimiter.Allow(ip) {
			DebugPrintf("🚨 BLOQUÉ: Requête de %s rejetée", ip)
			http.Error(w, "Trop de requêtes. Veuillez réessayer plus tard.", http.StatusTooManyRequests)
			return
		}
		
		// Appeler le gestionnaire original
		handler.ServeHTTP(w, req)
	}
}

// ServeHTTP implémente l'interface http.Handler
func (r *CustomRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	DebugPrintf("ServeHTTP appelé pour : %s", req.URL.Path)

	// Vérifier si la route est pour des fichiers statiques
	if strings.HasPrefix(req.URL.Path, "/public/") {
		DebugPrintf("Route statique détectée : %s", req.URL.Path)
		handler, ok := r.routes["/public/"]
		if ok {
			handler(w, req)
			return
		}
	}

	// Vérifier si la route existe
	handler, ok := r.routes[req.URL.Path]
	if ok {
		DebugPrintf("Route trouvée : %s", req.URL.Path)
		handler(w, req)
		return
	}

	// Sinon, retourner une erreur 404
	DebugPrintf("Route non trouvée : %s", req.URL.Path)
	NotFoundHandler(w, req)
}