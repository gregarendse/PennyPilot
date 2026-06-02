package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pennypilot/pennypilot/backend/internal/config"
	"github.com/pennypilot/pennypilot/backend/internal/domain"
	"github.com/pennypilot/pennypilot/backend/internal/store"
	"github.com/pennypilot/pennypilot/backend/internal/sync"
	"github.com/pennypilot/pennypilot/backend/internal/sync/csv"
	"github.com/pennypilot/pennypilot/backend/internal/sync/gocardless"
	"github.com/pennypilot/pennypilot/backend/internal/sync/monzo"
	"github.com/pennypilot/pennypilot/backend/internal/sync/truelayer"
)

// Dependencies groups collaborators needed by the HTTP layer.
type Dependencies struct {
	Config config.Config
	Logger *slog.Logger
	Store  *store.Store
}

// Handler exposes the REST API and OAuth callback endpoints.
type Handler struct {
	cfg      config.Config
	logger   *slog.Logger
	store    *store.Store
	registry *sync.Registry
}

func NewHandler(deps Dependencies) Handler {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	registry := sync.NewRegistry()
	if deps.Config.MonzoClientID != "" && deps.Config.MonzoClientSecret != "" {
		registry.Register(monzo.New(deps.Config.MonzoClientID, deps.Config.MonzoClientSecret, deps.Config.MonzoRedirectURL))
	}
	if deps.Config.TrueLayerClientID != "" && deps.Config.TrueLayerClientSecret != "" {
		registry.Register(truelayer.New(deps.Config.TrueLayerClientID, deps.Config.TrueLayerClientSecret, deps.Config.TrueLayerRedirectURL))
	}
	if deps.Config.GoCardlessSecretID != "" && deps.Config.GoCardlessSecretKey != "" {
		registry.Register(gocardless.New(deps.Config.GoCardlessSecretID, deps.Config.GoCardlessSecretKey))
	}
	registry.Register(csv.NewImporter())

	return Handler{
		cfg:      deps.Config,
		logger:   logger,
		store:    deps.Store,
		registry: registry,
	}
}

func (h Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", h.health)
	r.Get("/api/providers/ping", h.pingProviders)
	r.Get("/auth/{provider}", h.startAuth)
	r.Get("/auth/{provider}/callback", h.handleCallback)
	r.Get("/api/accounts", h.listAccounts)
	r.Post("/api/accounts/{accountID}/sync", h.syncAccount)
	r.Get("/api/transactions", h.listTransactions)
	r.Get("/api/categories", h.listCategories)
	r.Get("/api/budgets", h.listBudgets)

	if h.cfg.StaticPath != "" {
		r.Get("/*", spaHandler(h.cfg.StaticPath).ServeHTTP)
	}

	return r
}

func (h Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h Handler) pingProviders(w http.ResponseWriter, r *http.Request) {
	results := make(map[string]string)
	for name, connector := range h.registry.All() {
		if err := connector.Ping(r.Context()); err != nil {
			results[name] = "error: " + err.Error()
		} else {
			results[name] = "ok"
		}
	}

	writeJSON(w, http.StatusOK, results)
}

func (h Handler) startAuth(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	connector, err := h.registry.Get(provider)
	if err != nil {
		h.logger.Error("failed to find connector", "provider", provider, "error", err)
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, connector.AuthURL("replace-with-csrf-state"), http.StatusFound)
}

func (h Handler) handleCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	_, err := h.registry.Get(provider)
	if err != nil {
		h.logger.Error("failed to find connector for callback", "provider", provider, "error", err)
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"provider": provider,
		"status":   "oauth callback received; credential persistence is not implemented yet",
	})
}

func (h Handler) listAccounts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []domain.Account{})
}

func (h Handler) syncAccount(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusAccepted, map[string]string{
		"accountId": chi.URLParam(r, "accountID"),
		"status":    "sync job placeholder accepted",
	})
}

func (h Handler) listTransactions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []domain.Transaction{})
}

func (h Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, defaultCategories())
}

func (h Handler) listBudgets(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []domain.Budget{})
}

func defaultCategories() []domain.Category {
	return []domain.Category{
		{ID: "groceries", Name: "Groceries", Color: "#22c55e", Icon: "shopping-basket"},
		{ID: "bills", Name: "Bills", Color: "#3b82f6", Icon: "receipt"},
		{ID: "transport", Name: "Transport", Color: "#f97316", Icon: "train"},
		{ID: "income", Name: "Income", Color: "#14b8a6", Icon: "wallet"},
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to write json response", "error", err, "time", time.Now())
	}
}
