package handlers

import (
	"net/http"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/mercado/internal/store"
)

type DashboardData struct {
	meta.Site
	TotalContacts int64
	Env           string
}

type DashboardHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
}

func NewDashboardHandler(renderer *cais.Renderer, s store.Store, site meta.Site, cfg cais.Config) *DashboardHandler {
	return &DashboardHandler{renderer: renderer, store: s, site: site, cfg: cfg}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	count, err := h.store.CountContacts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	site := meta.ForRequest(h.site, r)
	site.ActiveNav = "dashboard"
	httpx.RenderOrError(w, h.renderer, "base", "dashboard", DashboardData{
		Site:          site,
		TotalContacts: count,
		Env:           h.cfg.Env,
	}, h.cfg)
}
