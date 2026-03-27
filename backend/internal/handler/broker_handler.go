package handler

import (
	"encoding/json"
	"net/http"
	"trackly-backend/internal/service"
	"trackly-backend/pkg/filter"

	"github.com/go-playground/validator/v10"
)

type BrokerHandler struct {
	Service       service.BrokerService
	AllowedFields []string
	Validator     *validator.Validate
}

func NewBrokerHandler(s service.BrokerService) *BrokerHandler {
	return &BrokerHandler{
		Service:       s,
		AllowedFields: []string{"symbol", "name", "time"},
		Validator:     validator.New(),
	}
}

// Register mendaftarkan rute dari handler ini ke mux
func (h *BrokerHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/brokers", h.brokers)
}

func (h *BrokerHandler) brokers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 1. Parsing query menggunakan filter package
		opt := filter.ParseQuery(r)

		// 2. Mengambil data dari service
		brokers, total, err := h.Service.GetAllBrokers(opt, h.AllowedFields)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 3. Menangani nilai default paginasi
		page := opt.Page
		if page < 1 {
			page = 1
		}
		limit := opt.Limit
		if limit < 1 {
			limit = 10
		}

		// 4. Membungkus hasil ke format standar paginasi
		resp := filter.WrapPaginated(brokers, total, page, limit)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

	case http.MethodPost:
		var input struct {
			Symbol string `json:"symbol" validate:"required,min=1,max=20"`
			Name   string `json:"name" validate:"required,min=2,max=100"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Validasi input
		if err := h.Validator.Struct(input); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := h.Service.AddBroker(input.Symbol, input.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Broker created successfully"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
