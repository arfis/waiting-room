package services

import (
	"context"
	"log/slog"
	"time"

	"git.prosoftke.sk/nghis/openapi/clients/go/nghisclinicalclient/v2"
	"github.com/arfis/waiting-room/nghis-adapter/internal/data/dto"
)

type Service struct {
	logger         *slog.Logger
	clinicalClient *nghisclinicalclient.APIClient
}

func NewService(
	logger *slog.Logger,
	clinicalClient *nghisclinicalclient.APIClient,
) *Service {
	return &Service{
		logger:         logger,
		clinicalClient: clinicalClient,
	}
}

func (s *Service) FindServices(ctx context.Context, req *dto.FindServicesReq) ([]dto.ServicesResp, error) {
	servicesFiltered, httpResp, err := s.clinicalClient.ServiceByProviderAPI.FilterServiceByProvider(ctx).ServiceByProviderFilterReq(nghisclinicalclient.ServiceByProviderFilterReq{
		ValidAt:               time.Now().UTC(),
		OrgUnitCodes:          req.OrgUnitCodes,
		ServiceTypeCategories: []nghisclinicalclient.ServiceTypeCategoryEnum{"EXAMINATION"},
	}).Execute()
	if err != nil {
		s.logger.Error("FilterServiceByProvider", "httpRep", httpResp, "err", err)
		return nil, err
	}

	resp := make([]dto.ServicesResp, 0, len(servicesFiltered))
	for _, sbp := range servicesFiltered {
		serviceFound, httpResp, err := s.clinicalClient.ServiceByProviderAPI.ShowServiceByProviderDetail(ctx, sbp.GetId()).Execute()
		if err != nil {
			s.logger.Error("ShowServiceByProviderDetail", "httpRep", httpResp, "err", err)
			return nil, err
		}

		for _, attribute := range serviceFound.Attributes {
			if attribute.Key == nghisclinicalclient.ATTRIBUTEKEY_KIOSK_BOOKABLE && attribute.Value == "ENABLED" {
				resp = append(resp, dto.ServicesResp{
					Code:     sbp.GetCode(),
					Duration: sbp.GetDuration(),
					Id:       sbp.GetId(),
					Name:     sbp.GetName(),
				})
			}
		}
	}

	return resp, nil
}
