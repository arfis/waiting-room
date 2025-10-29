package service

import (
	"context"
	"fmt"
	"log/slog"

	"git.prosoftke.sk/nghis/openapi/clients/go/nghisclinicalclient/v2"
	"git.prosoftke.sk/nghis/openapi/clients/go/nghispersonserviceclient"
	"github.com/arfis/waiting-room/nghis-adapter/internal/data/dto"
	"github.com/arfis/waiting-room/nghis-adapter/internal/errors"
)

type Service struct {
	logger         *slog.Logger
	clinicalClient *nghisclinicalclient.APIClient
	personClient   *nghispersonserviceclient.APIClient
}

func NewService(
	logger *slog.Logger,
	clinicalClient *nghisclinicalclient.APIClient,
	personClient *nghispersonserviceclient.APIClient,
) *Service {
	return &Service{
		logger:         logger,
		clinicalClient: clinicalClient,
		personClient:   personClient,
	}
}

func (s *Service) FindServices(ctx context.Context, req *dto.FindServicesReq) ([]dto.ServicesResp, error) {
	services, httpResp, err := s.clinicalClient.ServiceByProviderAPI.GetServicesListByAttributes(ctx).ListServicesByAttributesReq(nghisclinicalclient.ListServicesByAttributesReq{
		Limit: req.Limit,
		Attributes: []nghisclinicalclient.Attribute{{
			Key:   nghisclinicalclient.ATTRIBUTEKEY_KIOSK_BOOKABLE,
			Value: "ENABLED",
		}},
	}).Execute()
	if err != nil {
		s.logger.Error("FindAllBookableServices", "httpRep", httpResp, "err", err)
		return nil, errors.ServiceCall(fmt.Errorf("unable to retrieve bookable services"), nil)
	}

	// TODO: Need to filter services also by orgUnitID -> but they don't have any yet
	resp := make([]dto.ServicesResp, 0, len(services))
	for _, sbp := range services {
		resp = append(resp, dto.ServicesResp{
			Code:     sbp.GetCode(),
			Duration: sbp.GetDuration(),
			Id:       sbp.GetId(),
			Name:     sbp.GetName(),
		})
	}

	return resp, nil
}
