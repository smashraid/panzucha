package handlers

import (
	"encoding/json"
	"net/http"
	"panzucha/internal/handlers/dto"
	"panzucha/internal/handlers/mapper"
	"panzucha/internal/services"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type ProductHandler struct {
	services services.ProductService
}

func NewProductHandler(s services.ProductService) *ProductHandler {
	return &ProductHandler{services: s}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Bind request DTO
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 2. Validate request DTO (using validator package)
	if err := validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Map to domain model
	product := mapper.FromCreateProductRequest(&req)

	// 4. Call service (which uses repository – still domain model)
	if err := h.services.Create(r.Context(), product); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Map domain model to response DTO
	resp := mapper.FromDomainToResponse(product)

	// 6. Send JSON response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
