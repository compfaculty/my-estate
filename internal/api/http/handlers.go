package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/alexg/my-estate/internal/collectors/mock"
	"github.com/alexg/my-estate/internal/domain"
	"github.com/alexg/my-estate/internal/ingest"
	"github.com/go-chi/chi/v5"
)

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok", Time: time.Now().UTC().Format(time.RFC3339)})
}

func (s *Server) listSources(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.Sources().List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]sourceJSON, 0, len(items))
	for _, src := range items {
		out = append(out, sourceJSON{
			ID:                 src.ID.String(),
			Name:               src.Name,
			Enabled:            src.Enabled,
			FetchMode:          src.FetchMode,
			LastSuccessAt:      fmtTimePtr(src.LastSuccessAt),
			LastError:          src.LastError,
			LastCrawlAt:        fmtTimePtr(src.LastCrawlAt),
			LastRunStatus:      src.LastRunStatus,
			ListingsDiscovered: src.ListingsDiscovered,
			ListingsUpdated:    src.ListingsUpdated,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) searchProperties(w http.ResponseWriter, r *http.Request) {
	filter, err := parseSearch(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, total, err := s.store.Properties().Search(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	cards := make([]propertyCard, 0, len(items))
	for _, it := range items {
		cards = append(cards, cardFrom(it))
	}
	writeJSON(w, http.StatusOK, pageResponse{
		Items:  cards,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
}

func (s *Server) getProperty(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := s.store.Properties().Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if p == nil {
		writeError(w, http.StatusNotFound, "property not found")
		return
	}
	listings, err := s.store.Listings().ListByProperty(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	outListings := make([]listingJSON, 0, len(listings))
	for _, l := range listings {
		outListings = append(outListings, listingFrom(l))
	}
	writeJSON(w, http.StatusOK, propertyDetail{
		Property: propertyFrom(*p),
		Listings: outListings,
	})
}

func (s *Server) getPropertyHistory(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := s.store.Properties().Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if p == nil {
		writeError(w, http.StatusNotFound, "property not found")
		return
	}
	hist, err := s.store.History().ListByProperty(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]historyJSON, 0, len(hist))
	for _, h := range hist {
		out = append(out, historyFrom(h))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) devCollect(w http.ResponseWriter, r *http.Request) {
	if !s.enableDev {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	col := mock.New()
	src, err := s.store.Sources().GetByName(r.Context(), col.Name())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	wave := 1
	if src != nil && src.LastCrawlAt != nil {
		wave = 2
	}
	col.SetWave(wave)
	svc := ingest.New(s.store, s.log)
	result, err := svc.Run(r.Context(), col)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func parseSearch(r *http.Request) (domain.SearchFilter, error) {
	q := r.URL.Query()
	f := domain.SearchFilter{
		City:         q.Get("city"),
		Neighborhood: q.Get("neighborhood"),
		Street:       q.Get("street"),
		Query:        q.Get("q"),
		Source:       q.Get("source"),
	}
	if v := q.Get("property_type"); v != "" {
		f.PropertyType = domain.PropertyType(v)
	}
	if v, err := parseFloat(q.Get("min_price")); err != nil {
		return f, err
	} else {
		f.MinPrice = v
	}
	if v, err := parseFloat(q.Get("max_price")); err != nil {
		return f, err
	} else {
		f.MaxPrice = v
	}
	if v, err := parseFloat(q.Get("min_area")); err != nil {
		return f, err
	} else {
		f.MinArea = v
	}
	if v, err := parseFloat(q.Get("max_area")); err != nil {
		return f, err
	} else {
		f.MaxArea = v
	}
	if v, err := parseFloat(q.Get("min_rooms")); err != nil {
		return f, err
	} else {
		f.MinRooms = v
	}
	if v, err := parseFloat(q.Get("max_rooms")); err != nil {
		return f, err
	} else {
		f.MaxRooms = v
	}
	if v, err := parseInt(q.Get("min_floor")); err != nil {
		return f, err
	} else {
		f.MinFloor = v
	}
	if v, err := parseInt(q.Get("max_floor")); err != nil {
		return f, err
	} else {
		f.MaxFloor = v
	}
	if v, err := parseFloat(q.Get("min_lat")); err != nil {
		return f, err
	} else {
		f.MinLat = v
	}
	if v, err := parseFloat(q.Get("min_lng")); err != nil {
		return f, err
	} else {
		f.MinLng = v
	}
	if v, err := parseFloat(q.Get("max_lat")); err != nil {
		return f, err
	} else {
		f.MaxLat = v
	}
	if v, err := parseFloat(q.Get("max_lng")); err != nil {
		return f, err
	} else {
		f.MaxLng = v
	}
	if v, err := parseFloat(q.Get("lat")); err != nil {
		return f, err
	} else {
		f.Lat = v
	}
	if v, err := parseFloat(q.Get("lng")); err != nil {
		return f, err
	} else {
		f.Lng = v
	}
	if v, err := parseFloat(q.Get("radius_m")); err != nil {
		return f, err
	} else {
		f.RadiusM = v
	}
	if q.Get("elevator") == "true" {
		t := true
		f.Elevator = &t
	}
	if q.Get("parking") == "true" {
		t := true
		f.Parking = &t
	}
	if q.Get("balcony") == "true" {
		t := true
		f.Balcony = &t
	}
	if q.Get("safe_room") == "true" {
		t := true
		f.SafeRoom = &t
	}
	if days := q.Get("listing_age_days"); days != "" {
		n, err := strconv.Atoi(days)
		if err != nil {
			return f, err
		}
		d := time.Duration(n) * 24 * time.Hour
		f.ListingAge = &d
	}
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return f, err
		}
		f.Limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return f, err
		}
		f.Offset = n
	}
	return f.WithDefaults(), nil
}

func parseFloat(s string) (*float64, error) {
	if s == "" {
		return nil, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func parseInt(s string) (*int, error) {
	if s == "" {
		return nil, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
