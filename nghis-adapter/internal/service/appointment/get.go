package appointment

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	errs "errors"

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

func (s *Service) GetAppointmentsForPatient(ctx context.Context, nationalID string) ([]dto.AppointmentResp, error) {
	currTime := time.Now()
	resp := make([]dto.AppointmentResp, 0)

	patient, httpResp, err := s.personClient.PatientAPI.FindPatientByNationalID(ctx, nationalID).Execute()
	if err != nil {
		s.logger.Error("unable to find patient by national ID", "nationalID", nationalID, "httpResp", httpResp, "err", err)
		if httpResp.StatusCode == 404 {
			return nil, errors.NotFound(errs.New("patient not found"), nil)
		}
		return nil, errors.ServiceCall(err, nil)
	}

	events, httpResp, err := s.clinicalClient.CareItemAPI.FindAllCareItems(ctx).
		PatientId(strconv.FormatInt(patient.Id, 10)).
		AfterRequestedTime(currTime).
		BeforeRequestedTime(currTime.Add(12 * time.Hour)). // 12-hour window atm
		CareItemStatuses([]nghisclinicalclient.CareItemStatusEnum{nghisclinicalclient.CAREITEMSTATUSENUM_PLANNED}).
		Size(100).
		Page(0).
		Execute()
	if err != nil {
		s.logger.Error("unable to load events for patient", "patientID", patient.GetId(), "httpResp", httpResp, "err", err)
		return nil, errors.ServiceCall(err, nil)
	}

	for _, event := range events.GetContent() {
		resp = append(resp, dto.AppointmentResp{
			Id:            strconv.FormatInt(event.GetId(), 10),
			Duration:      event.GetDuration(),
			RequestedTime: event.GetRequestedTime(),
			ServiceName:   event.RequestedService.GetName(),
		})
	}

	return resp, nil
}
