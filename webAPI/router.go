package webAPI

import (
    "html/template"
    "net/http"
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

// CustomRouter est un routeur personnalisé qui gère les erreurs 404
type CustomRouter struct {
    routes map[string]http.HandlerFunc
    static http.Handler
}

// NewCustomRouter crée un nouveau routeur personnalisé
func NewCustomRouter() *CustomRouter {
    return &CustomRouter{
        routes: make(map[string]http.HandlerFunc),
        static: http.FileServer(http.Dir("public")),
    }
}

// HandleFunc ajoute une route au routeur
func (r *CustomRouter) HandleFunc(path string, handler http.HandlerFunc) {
    r.routes[path] = handler
}

// Handle ajoute un gestionnaire au routeur
func (r *CustomRouter) Handle(path string, handler http.Handler) {
    r.routes[path] = handler.ServeHTTP
}

// ServeHTTP implémente l'interface http.Handler
func (r *CustomRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // Vérifier si la route est pour des fichiers statiques
    if handler, ok := r.routes["/public/"]; ok && len(req.URL.Path) >= 8 && req.URL.Path[:8] == "/public/" {
        handler(w, req)
        return
    }

    // Vérifier si la route existe
    if handler, ok := r.routes[req.URL.Path]; ok {
        handler(w, req)
        return
    }

    // Sinon, retourner une erreur 404
    NotFoundHandler(w, req)
}