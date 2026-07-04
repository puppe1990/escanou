package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/barcode"
	"github.com/puppe1990/cais/pkg/cais/httpx"
	"github.com/puppe1990/cais/pkg/cais/meta"
	"github.com/puppe1990/cais/pkg/cais/middleware"
	"github.com/puppe1990/cais/pkg/cais/money"
	"github.com/puppe1990/cais/pkg/cais/session"

	"github.com/puppe1990/mercado/internal/models"
	"github.com/puppe1990/mercado/internal/store"
)

type SupermarketPageData struct {
	meta.Site
	Scans   []PriceScanView
	Markets []MarketView

	Badges              []BadgeView
	Leaders             []LeaderView
	LevelXP             int
	LevelXPMax          int
	LevelProgressPct    int
	UnlockedBadgeCount  int
	TotalBadgeCount     int
	UserDisplayName     string
	UserCity            string
	LookupProduct       *ProductView
	LookupError     string
	LookupBarcode   string
	LookupNeedsName bool
	ReportProductID int64
	SupermarketOpts []SupermarketOption
	SubmissionCount   int
	ShoppingListTotal string
}

type SupermarketOption struct {
	ID   int64
	Name string
}

type PriceScanView struct {
	ID              int
	ProductName     string
	SupermarketName string
	Price           string
	TimeAgo         string
	Contributor     string
	Level           int
	ConfirmedCount  int
	DisputeCount    int
	ConfirmDisabled bool
	DisputeDisabled bool
	OwnReport       bool
	Verified        bool
	Outdated        bool
	Flagged         bool
}

type MarketView struct {
	Name     string
	Address  string
	Distance string
	Offers   int
	BestDeal string
}

type ProductView struct {
	ID       int64
	Name     string
	Barcode  string
	Category string
	Brand    string
	Quantity string
	ImageURL string
	Source   string
	AvgPrice string
}

type BadgeView struct {
	ID          string
	Name        string
	Description string
	Unlocked    bool
	Icon        string
}

type LeaderView struct {
	Name   string
	Points int
	Rank   int
	Level  int
	IsYou  bool
}

type SupermarketHandler struct {
	renderer *cais.Renderer
	store    store.Store
	site     meta.Site
	cfg      cais.Config
	barcode  *barcode.Client
}

func NewSupermarketHandler(renderer *cais.Renderer, st store.Store, site meta.Site, cfg cais.Config) *SupermarketHandler {
	return &SupermarketHandler{
		renderer: renderer,
		store:    st,
		site:     site,
		cfg:      cfg,
		barcode:  &barcode.Client{},
	}
}

func (h *SupermarketHandler) requireAuth(w http.ResponseWriter, r *http.Request) bool {
	if _, ok := session.UserID(r); ok {
		return true
	}
	if cais.IsHTMX(r) {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	httpx.SeeOther(w, r, "/login")
	return false
}

func (h *SupermarketHandler) base(r *http.Request) SupermarketPageData {
	site := meta.ForRequest(h.site, r)
	if stats, ok := middleware.UserStatsFrom(r); ok {
		site.UserLevel = stats.Level
		site.UserPoints = stats.Points
		site.UserRank = stats.Rank
	}
	if _, ok := session.UserID(r); ok {
		site.LoggedIn = true
	}
	return SupermarketPageData{Site: site}
}

func (h *SupermarketHandler) Scan(w http.ResponseWriter, r *http.Request) {
	if !h.requireAuth(w, r) {
		return
	}
	data := h.base(r)
	data.ActiveNav = "scan"
	data.SupermarketOpts = h.supermarketOptions()
	if uid, ok := session.UserID(r); ok {
		data.Scans, data.ShoppingListTotal = h.userShoppingList(uid)
	}
	httpx.RenderOrError(w, h.renderer, "base", "scan", data, h.cfg)
}

func (h *SupermarketHandler) Map(w http.ResponseWriter, r *http.Request) {
	if !h.requireAuth(w, r) {
		return
	}
	data := h.base(r)
	data.ActiveNav = "map"
	data.Markets = h.marketViews()
	httpx.RenderOrError(w, h.renderer, "base", "map", data, h.cfg)
}

func (h *SupermarketHandler) Feed(w http.ResponseWriter, r *http.Request) {
	if !h.requireAuth(w, r) {
		return
	}
	data := h.base(r)
	data.ActiveNav = "feed"
	uid, _ := session.UserID(r)
	data.Scans = h.feedScans(uid)
	httpx.RenderOrError(w, h.renderer, "base", "feed", data, h.cfg)
}

func (h *SupermarketHandler) Achievements(w http.ResponseWriter, r *http.Request) {
	if !h.requireAuth(w, r) {
		return
	}
	data := h.base(r)
	data.ActiveNav = "achievements"
	uid, _ := session.UserID(r)
	data.Badges = h.badgeViews(uid)
	data.Leaders = h.leaderViews(uid)
	if uid > 0 {
		reports, _ := h.store.ListUserReports(uid, 1000)
		data.SubmissionCount = len(reports)
		if level, points, rank, err := h.store.LoadStats(uid); err == nil {
			data.UserLevel = level
			data.UserPoints = points
			data.UserRank = rank
			data.LevelXP, data.LevelXPMax, data.LevelProgressPct = levelProgress(points)
		}
		if p, err := h.store.GetOrCreateProfile(uid); err == nil {
			data.UserDisplayName = p.DisplayName
			if data.UserDisplayName == "" {
				data.UserDisplayName = "Colaborador"
			}
			data.UserCity = p.City
		}
	}
	for _, b := range data.Badges {
		data.TotalBadgeCount++
		if b.Unlocked {
			data.UnlockedBadgeCount++
		}
	}
	httpx.RenderOrError(w, h.renderer, "base", "achievements", data, h.cfg)
}

const pointsPerLevel = 100

func levelProgress(points int) (xp, max, pct int) {
	max = pointsPerLevel
	xp = points % max
	pct = xp
	return xp, max, pct
}

func (h *SupermarketHandler) NFCe(w http.ResponseWriter, r *http.Request) {
	if !h.requireAuth(w, r) {
		return
	}
	data := h.base(r)
	data.ActiveNav = "nfce"
	httpx.RenderOrError(w, h.renderer, "base", "nfce", data, h.cfg)
}

func (h *SupermarketHandler) LookupPost(w http.ResponseWriter, r *http.Request) {
	if !h.requireAuth(w, r) {
		return
	}
	data := h.base(r)
	data.ActiveNav = "scan"
	ean := normalizeBarcode(r.FormValue("barcode"))
	name := strings.TrimSpace(r.FormValue("name"))
	if ean == "" {
		data.LookupError = "Informe o código de barras"
		h.renderLookup(w, r, data)
		return
	}
	product, found, err := h.store.FindProductByBarcode(ean)
	if err != nil {
		http.Error(w, "lookup failed", http.StatusInternalServerError)
		return
	}
	if !found {
		if !validBarcode(ean) {
			data.LookupError = "Código inválido — confira os dígitos"
			data.LookupBarcode = ean
			h.renderLookupInvalid(w, r, data)
			return
		}
		if name == "" {
			off, ok, err := h.barcode.Lookup(r.Context(), ean)
			if err == nil && ok {
				product, err = h.store.CreateProductFromOFF(off)
				if err != nil {
					http.Error(w, "create product failed", http.StatusInternalServerError)
					return
				}
				found = true
			}
		}
		if !found {
			if name == "" {
				data.LookupError = "Produto não encontrado — informe o nome para cadastrar"
				data.LookupBarcode = ean
				data.LookupNeedsName = true
				h.renderLookup(w, r, data)
				return
			}
			id, err := h.store.CreateProduct(name, ean, "Geral")
			if err != nil {
				http.Error(w, "create product failed", http.StatusInternalServerError)
				return
			}
			product = models.Product{ID: id, Name: name, Barcode: ean, Category: "Geral"}
		}
	}
	pv := h.productView(product)
	data.LookupProduct = &pv
	data.ReportProductID = product.ID
	data.SupermarketOpts = h.supermarketOptions()
	h.renderLookup(w, r, data)
}

func (h *SupermarketHandler) renderLookup(w http.ResponseWriter, r *http.Request, data SupermarketPageData) {
	httpx.RenderPageOrPartial(w, r, h.renderer, httpx.RenderOptions{
		Layout:  "base",
		Page:    "scan",
		Partial: "scan_result",
		Data:    data,
	}, h.cfg)
}

func (h *SupermarketHandler) renderLookupInvalid(w http.ResponseWriter, r *http.Request, data SupermarketPageData) {
	opts := httpx.RenderOptions{
		Layout:  "base",
		Page:    "scan",
		Partial: "scan_result",
		Data:    data,
	}
	if cais.IsHTMX(r) {
		opts.Status = http.StatusUnprocessableEntity
	}
	httpx.RenderPageOrPartial(w, r, h.renderer, opts, h.cfg)
}

func (h *SupermarketHandler) ReportPost(w http.ResponseWriter, r *http.Request) {
	if !h.requireAuth(w, r) {
		return
	}
	userID, _ := session.UserID(r)
	productID, _ := strconv.ParseInt(r.FormValue("product_id"), 10, 64)
	supermarketID, _ := strconv.ParseInt(r.FormValue("supermarket_id"), 10, 64)
	priceCents, err := parsePriceCents(r.FormValue("price"))
	if err != nil || productID == 0 || supermarketID == 0 || priceCents <= 0 || priceCents > 999_900 {
		data := h.base(r)
		data.ActiveNav = "scan"
		data.LookupError = "Preencha supermercado e preço válidos"
		data.SupermarketOpts = h.supermarketOptions()
		if cais.IsHTMX(r) {
			httpx.RenderPageOrPartial(w, r, h.renderer, httpx.RenderOptions{
				Partial: "scan_result",
				Data:    data,
				Status:  http.StatusUnprocessableEntity,
			}, h.cfg)
			return
		}
		httpx.RenderOrError(w, h.renderer, "base", "scan", data, h.cfg)
		return
	}
	if _, err := h.store.CreatePriceReport(userID, productID, supermarketID, priceCents); err != nil {
		http.Error(w, "report failed", http.StatusInternalServerError)
		return
	}
	cais.SetToast(w, "+10 pts — preço registrado!")
	if cais.IsHTMX(r) {
		data := h.base(r)
		data.ActiveNav = "scan"
		data.Scans, data.ShoppingListTotal = h.userShoppingList(userID)
		httpx.RenderPageOrPartial(w, r, h.renderer, httpx.RenderOptions{
			Partial: "scan_report_done",
			Data:    data,
		}, h.cfg)
		return
	}
	httpx.SeeOther(w, r, "/feed")
}

func (h *SupermarketHandler) ConfirmPost(w http.ResponseWriter, r *http.Request, reportID int64) {
	if !h.requireAuth(w, r) {
		return
	}
	userID, _ := session.UserID(r)
	count, err := h.store.ConfirmPriceReport(reportID, userID)
	disputeCount := h.disputeCountFor(reportID)
	if err == store.ErrOwnReport || err == store.ErrAlreadyConfirmed {
		h.renderFeedVotes(w, feedVoteData{
			ID: int(reportID), ConfirmedCount: count, DisputeCount: disputeCount,
			ConfirmDisabled: true, DisputeDisabled: h.viewerDisputed(reportID, userID),
			OwnReport: err == store.ErrOwnReport,
		})
		return
	}
	if err != nil {
		http.Error(w, "confirm failed", http.StatusInternalServerError)
		return
	}
	cais.SetToast(w, "+2 pts — obrigado por confirmar!")
	h.renderFeedVotes(w, feedVoteData{
		ID: int(reportID), ConfirmedCount: count, DisputeCount: disputeCount,
		ConfirmDisabled: true, DisputeDisabled: h.viewerDisputed(reportID, userID),
	})
}

func (h *SupermarketHandler) FlagPost(w http.ResponseWriter, r *http.Request, reportID int64) {
	if !h.requireAuth(w, r) {
		return
	}
	userID, _ := session.UserID(r)
	count, err := h.store.DisputePriceReport(reportID, userID)
	confirmCount := h.confirmCountFor(reportID)
	if err == store.ErrOwnReport || err == store.ErrAlreadyDisputed {
		h.renderFeedVotes(w, feedVoteData{
			ID: int(reportID), ConfirmedCount: confirmCount, DisputeCount: count,
			ConfirmDisabled: h.viewerConfirmed(reportID, userID),
			DisputeDisabled: true,
			OwnReport: err == store.ErrOwnReport,
		})
		return
	}
	if err != nil {
		http.Error(w, "dispute failed", http.StatusInternalServerError)
		return
	}
	if count >= store.DisputeFlagThreshold {
		cais.SetToast(w, "Preço ocultado após várias contestações")
		cais.SetRetarget(w, fmt.Sprintf("#feed-item-%d", reportID))
		w.Header().Set("HX-Reswap", "delete")
		w.WriteHeader(http.StatusOK)
		return
	}
	h.renderFeedVotes(w, feedVoteData{
		ID: int(reportID), ConfirmedCount: confirmCount, DisputeCount: count,
		ConfirmDisabled: h.viewerConfirmed(reportID, userID),
		DisputeDisabled: true,
	})
}

type feedVoteData struct {
	ID              int
	ConfirmedCount  int
	DisputeCount    int
	ConfirmDisabled bool
	DisputeDisabled bool
	OwnReport       bool
}

func (h *SupermarketHandler) renderFeedVotes(w http.ResponseWriter, data feedVoteData) {
	if err := httpx.RenderPartial(w, h.renderer, "feed_vote_btns", data); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
	}
}

func (h *SupermarketHandler) confirmCountFor(reportID int64) int {
	reports, _ := h.store.ListFeedReports(1000, 0)
	for _, r := range reports {
		if r.ID == reportID {
			return r.Confirmations
		}
	}
	return 0
}

func (h *SupermarketHandler) disputeCountFor(reportID int64) int {
	reports, _ := h.store.ListFeedReports(1000, 0)
	for _, r := range reports {
		if r.ID == reportID {
			return r.Disputes
		}
	}
	return 0
}

func (h *SupermarketHandler) viewerConfirmed(reportID, userID int64) bool {
	reports, _ := h.store.ListFeedReports(1000, userID)
	for _, r := range reports {
		if r.ID == reportID {
			return r.ViewerConfirmed
		}
	}
	return false
}

func (h *SupermarketHandler) viewerDisputed(reportID, userID int64) bool {
	reports, _ := h.store.ListFeedReports(1000, userID)
	for _, r := range reports {
		if r.ID == reportID {
			return r.ViewerDisputed
		}
	}
	return false
}

func (h *SupermarketHandler) feedScans(viewerID int64) []PriceScanView {
	reports, err := h.store.ListFeedReports(50, viewerID)
	if err != nil {
		return nil
	}
	return priceReportsToViews(reports, viewerID)
}

func shoppingListTotalCents(prices []int) int {
	total := 0
	for _, p := range prices {
		total += p
	}
	return total
}

func (h *SupermarketHandler) userShoppingList(userID int64) ([]PriceScanView, string) {
	reports, err := h.store.ListUserReports(userID, 20)
	if err != nil {
		return nil, money.FormatBRL(0)
	}
	total := 0
	for _, r := range reports {
		total += r.PriceCents
	}
	return priceReportsToViews(reports, userID), money.FormatBRL(total)
}

func priceReportsToViews(reports []models.PriceReport, viewerID int64) []PriceScanView {
	out := make([]PriceScanView, 0, len(reports))
	for _, r := range reports {
		out = append(out, PriceScanView{
			ID:              int(r.ID),
			ProductName:     r.ProductName,
			SupermarketName: r.SupermarketName,
			Price:           money.FormatBRL(r.PriceCents),
			TimeAgo:         timeAgo(r.CreatedAt),
			Contributor:     r.Contributor,
			Level:           r.ContributorLvl,
			ConfirmedCount:  r.Confirmations,
			DisputeCount:    r.Disputes,
			ConfirmDisabled: r.ViewerConfirmed || r.UserID == viewerID,
			DisputeDisabled: r.ViewerDisputed || r.UserID == viewerID,
			OwnReport:       r.UserID == viewerID,
			Verified:        store.ReportVerified(r.Confirmations),
			Outdated:        store.ReportOutdated(r.CreatedAt),
			Flagged:         r.Flagged,
		})
	}
	return out
}

func (h *SupermarketHandler) productView(p models.Product) ProductView {
	avg, _ := h.store.ProductAvgPriceCents(p.ID)
	avgStr := ""
	if avg > 0 {
		avgStr = money.FormatBRL(avg)
	}
	return ProductView{
		ID: p.ID, Name: p.Name, Barcode: p.Barcode, Category: p.Category,
		Brand: p.Brand, Quantity: p.Quantity, ImageURL: p.ImageURL, Source: p.Source,
		AvgPrice: avgStr,
	}
}

func (h *SupermarketHandler) supermarketOptions() []SupermarketOption {
	markets, err := h.store.ListSupermarkets()
	if err != nil {
		return nil
	}
	out := make([]SupermarketOption, 0, len(markets))
	for _, m := range markets {
		out = append(out, SupermarketOption{ID: m.ID, Name: m.Name})
	}
	return out
}

func (h *SupermarketHandler) marketViews() []MarketView {
	markets, err := h.store.ListSupermarkets()
	if err != nil {
		return nil
	}
	out := make([]MarketView, 0, len(markets))
	for i, m := range markets {
		offers, _ := h.store.SupermarketOfferCount(m.ID)
		best, _ := h.store.SupermarketBestDeal(m.ID)
		out = append(out, MarketView{
			Name: m.Name, Address: m.Address,
			Distance: fmt.Sprintf("%.1f km", float64(i+1)*1.2),
			Offers:   offers, BestDeal: best,
		})
	}
	return out
}

func (h *SupermarketHandler) badgeViews(userID int64) []BadgeView {
	badges, err := h.store.ListBadges(userID)
	if err != nil {
		return nil
	}
	out := make([]BadgeView, 0, len(badges))
	for _, b := range badges {
		out = append(out, BadgeView{
			ID: b.Slug, Name: b.Name, Description: b.Description, Unlocked: b.Unlocked, Icon: b.Icon,
		})
	}
	return out
}

func (h *SupermarketHandler) leaderViews(userID int64) []LeaderView {
	leaders, err := h.store.Leaderboard(10, userID)
	if err != nil {
		return nil
	}
	out := make([]LeaderView, 0, len(leaders))
	for _, l := range leaders {
		name := l.Name
		if l.IsYou {
			name = "Você"
		}
		out = append(out, LeaderView{
			Name: name, Points: l.Points, Rank: l.Rank, Level: l.Level, IsYou: l.IsYou,
		})
	}
	return out
}

func parsePriceCents(raw string) (int, error) {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, "R$", ""))
	raw = strings.ReplaceAll(raw, " ", "")
	if raw == "" {
		return 0, fmt.Errorf("price required")
	}
	raw = strings.ReplaceAll(raw, ",", ".")
	parts := strings.SplitN(raw, ".", 2)
	reais, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	frac := 0
	if len(parts) == 2 {
		p := parts[1]
		if len(p) > 2 {
			p = p[:2]
		}
		for len(p) < 2 {
			p += "0"
		}
		frac, _ = strconv.Atoi(p)
	}
	cents := reais*100 + frac
	if cents <= 0 {
		return 0, fmt.Errorf("price must be positive")
	}
	return cents, nil
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return "Há poucos minutos"
	case d < 24*time.Hour:
		return fmt.Sprintf("Há %d horas", int(d.Hours()))
	case d < 48*time.Hour:
		return "Ontem"
	default:
		return fmt.Sprintf("Há %d dias", int(d.Hours()/24))
	}
}
