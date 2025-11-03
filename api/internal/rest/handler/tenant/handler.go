package tenant

import (
	"encoding/json"
	"net/http"

	"github.com/arfis/waiting-room/internal/data/dto"
	ngErrors "github.com/arfis/waiting-room/internal/errors"
	"github.com/arfis/waiting-room/internal/rest/handler"
	"github.com/arfis/waiting-room/internal/service/tenant"
)

type Handler struct {
	svc                  *tenant.Service
	responseErrorHandler *ngErrors.ResponseErrorHandler
}

func New(
	svc *tenant.Service,
	responseErrorHandler *ngErrors.ResponseErrorHandler,
) *Handler {
	return &Handler{
		svc:                  svc,
		responseErrorHandler: responseErrorHandler,
	}
}

func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var applicationErr error
	req := dto.Tenant{}
	applicationErr = json.NewDecoder(r.Body).Decode(&req)
	if applicationErr != nil {
		h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.New(ngErrors.InternalServerErrorCode, "problem decoding request body", http.StatusInternalServerError, nil))
		return
	}
	applicationErr = handler.GetValidator().Struct(req)
	if applicationErr != nil {
		h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.RequestValidation(applicationErr))
		return
	}
	var resp *dto.Tenant
	resp, applicationErr = h.svc.CreateTenant(
		r.Context(), &req,
	)
	if applicationErr != nil {
		h.responseErrorHandler.HandleAndWriteError(w, r, applicationErr)
		return
	}
	handler.WriteJson(r.Context(), w, 201, resp)
}

func (h *Handler) GetTenant(w http.ResponseWriter, r *http.Request) {
	var applicationErr error
	tenantID := handler.PathParamToString(r, "id")
	var resp *dto.Tenant
	resp, applicationErr = h.svc.GetTenant(
		r.Context(),
		tenantID,
	)
	if applicationErr != nil {
		h.responseErrorHandler.HandleAndWriteError(w, r, applicationErr)
		return
	}
	handler.WriteJson(r.Context(), w, 200, resp)
}

func (h *Handler) GetAllTenants(w http.ResponseWriter, r *http.Request) {
	var applicationErr error
	var resp []dto.Tenant
	resp, applicationErr = h.svc.GetAllTenants(
		r.Context(),
	)
	if applicationErr != nil {
		h.responseErrorHandler.HandleAndWriteError(w, r, applicationErr)
		return
	}
	handler.WriteJson(r.Context(), w, 200, resp)
}

func (h *Handler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	var applicationErr error
	req := dto.Tenant{}
	applicationErr = json.NewDecoder(r.Body).Decode(&req)
	if applicationErr != nil {
		h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.New(ngErrors.InternalServerErrorCode, "problem decoding request body", http.StatusInternalServerError, nil))
		return
	}
	applicationErr = handler.GetValidator().Struct(req)
	if applicationErr != nil {
		h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.RequestValidation(applicationErr))
		return
	}
	var resp *dto.Tenant
	resp, applicationErr = h.svc.UpdateTenant(
		r.Context(), &req,
	)
	if applicationErr != nil {
		h.responseErrorHandler.HandleAndWriteError(w, r, applicationErr)
		return
	}
	handler.WriteJson(r.Context(), w, 200, resp)
}

func (h *Handler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	var applicationErr error
	tenantID := handler.PathParamToString(r, "id")
	applicationErr = h.svc.DeleteTenant(
		r.Context(),
		tenantID,
	)
	if applicationErr != nil {
		h.responseErrorHandler.HandleAndWriteError(w, r, applicationErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
