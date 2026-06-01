package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/pennypilot/pennypilot/backend/internal/config"
	"github.com/pennypilot/pennypilot/backend/internal/domain"
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
}

// Handler exposes the REST API and OAuth callback endpoints.
type Handler struct {
	cfg      config.Config
	logger   *slog.Logger
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

	return Handler{cfg: deps.Config, logger: logger, registry: registry}
}

func (h Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /api/providers/ping", h.pingProviders)
	mux.HandleFunc("GET /auth/{provider}", h.startAuth)
	mux.HandleFunc("GET /auth/{provider}/callback", h.handleCallback)
	mux.HandleFunc("GET /api/accounts", h.listAccounts)
	mux.HandleFunc("POST /api/accounts/{accountID}/sync", h.syncAccount)
	mux.HandleFunc("GET /api/transactions", h.listTransactions)
	mux.HandleFunc("GET /api/categories", h.listCategories)
	mux.HandleFunc("GET /api/budgets", h.listBudgets)

	return h.requestLogger(mux)
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
	provider := r.PathValue("provider")
	connector, err := h.registry.Get(provider)
	if err != nil {
		h.logger.Error("failed to find connector", "provider", provider, "error", err)
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, connector.AuthURL("replace-with-csrf-state"), http.StatusFound)
}

func (h Handler) handleCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
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
		"accountId": r.PathValue("accountID"),
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

func (h Handler) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		h.logger.Info("http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(startedAt).String())
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to write json response", "error", err, "time", time.Now())
	}
}
