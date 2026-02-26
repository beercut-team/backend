package service

import (
	"context"
	"strings"

	"github.com/beercut-team/backend-boilerplate/internal/domain"
	"github.com/beercut-team/backend-boilerplate/internal/repository"
)

type MedicalStandardsService interface {
	UpdateMedicalMetadata(ctx context.Context, patientID uint, req domain.UpdateMedicalMetadataRequest) error
	SearchICD10Codes(ctx context.Context, query string) []domain.MedicalCodeSearchResult
	SearchSNOMEDCodes(ctx context.Context, query string) []domain.MedicalCodeSearchResult
	SearchLOINCCodes(ctx context.Context, query string) []domain.MedicalCodeSearchResult
}

type medicalStandardsService struct {
	patientRepo repository.PatientRepository
}

func NewMedicalStandardsService(patientRepo repository.PatientRepository) MedicalStandardsService {
	return &medicalStandardsService{patientRepo: patientRepo}
}

func (s *medicalStandardsService) UpdateMedicalMetadata(ctx context.Context, patientID uint, req domain.UpdateMedicalMetadataRequest) error {
	patient, err := s.patientRepo.FindByID(ctx, patientID)
	if err != nil {
		return err
	}

	if patient.MedicalMetadata == nil {
		patient.MedicalMetadata = &domain.MedicalStandardsMetadata{}
	}

	if len(req.DiagnosisCodes) > 0 {
		patient.MedicalMetadata.DiagnosisCodes = req.DiagnosisCodes
	}
	if len(req.ProcedureCodes) > 0 {
		patient.MedicalMetadata.ProcedureCodes = req.ProcedureCodes
	}
	if len(req.Observations) > 0 {
		patient.MedicalMetadata.Observations = req.Observations
	}
	if len(req.Extensions) > 0 {
		patient.MedicalMetadata.Extensions = req.Extensions
	}

	return s.patientRepo.Update(ctx, patient)
}

func (s *medicalStandardsService) SearchICD10Codes(ctx context.Context, query string) []domain.MedicalCodeSearchResult {
	query = strings.ToLower(query)
	results := []domain.MedicalCodeSearchResult{}

	// Predefined cataract codes
	cataractCodes := map[string]string{
		"H25.0": "Старческая начальная катаракта",
		"H25.1": "Старческая ядерная катаракта",
		"H25.2": "Старческая морганиева катаракта",
		"H25.8": "Другие старческие катаракты",
		"H25.9": "Старческая катаракта неуточненная",
		"H26.0": "Детская, юношеская и пресенильная катаракта",
		"H26.1": "Травматическая катаракта",
		"H26.2": "Осложненная катаракта",
		"H26.3": "Катаракта, вызванная лекарственными средствами",
		"H26.4": "Вторичная катаракта",
		"H26.8": "Другая уточненная катаракта",
		"H26.9": "Катаракта неуточненная",
		"H40.0": "Подозрение на глаукому",
		"H40.1": "Первичная открытоугольная глаукома",
		"H40.2": "Первичная закрытоугольная глаукома",
		"H43.1": "Кровоизлияние в стекловидное тело",
	}

	for code, display := range cataractCodes {
		if strings.Contains(strings.ToLower(display), query) || strings.Contains(strings.ToLower(code), query) {
			results = append(results, domain.MedicalCodeSearchResult{
				Code:    code,
				Display: display,
				System:  "ICD-10",
			})
		}
	}

	return results
}

func (s *medicalStandardsService) SearchSNOMEDCodes(ctx context.Context, query string) []domain.MedicalCodeSearchResult {
	query = strings.ToLower(query)
	results := []domain.MedicalCodeSearchResult{}

	// Predefined ophthalmic procedure codes
	procedureCodes := map[string]string{
		"172522003": "Факоэмульсификация с имплантацией ИОЛ",
		"231744001": "Факоэмульсификация катаракты",
		"308694008": "Имплантация интраокулярной линзы",
		"397544007": "Экстракапсулярная экстракция катаракты",
		"415089008": "Интракапсулярная экстракция катаракты",
		"46309007":  "Трабекулэктомия",
		"397193006": "Витрэктомия",
		"231760002": "Лазерная капсулотомия",
		"231761003": "YAG лазерная капсулотомия",
		"397544008": "Факоаспирация",
		"252957005": "Биомикроскопия глаза",
		"252958000": "Офтальмоскопия",
		"252959008": "Тонометрия",
		"252960003": "Периметрия",
		"252961004": "Оптическая когерентная томография",
	}

	for code, display := range procedureCodes {
		if strings.Contains(strings.ToLower(display), query) || strings.Contains(code, query) {
			results = append(results, domain.MedicalCodeSearchResult{
				Code:    code,
				Display: display,
				System:  "SNOMED-CT",
			})
		}
	}

	return results
}

func (s *medicalStandardsService) SearchLOINCCodes(ctx context.Context, query string) []domain.MedicalCodeSearchResult {
	query = strings.ToLower(query)
	results := []domain.MedicalCodeSearchResult{}

	// Predefined ocular biometry codes
	biometryCodes := map[string]string{
		"79894-2": "Длина оси правого глаза",
		"79895-9": "Длина оси левого глаза",
		"79897-5": "Кератометрия K1 правого глаза",
		"79898-3": "Кератометрия K2 правого глаза",
		"79899-1": "Кератометрия K1 левого глаза",
		"79900-7": "Кератометрия K2 левого глаза",
		"79901-5": "Глубина передней камеры правого глаза",
		"79902-3": "Глубина передней камеры левого глаза",
		"79903-1": "Толщина хрусталика правого глаза",
		"79904-9": "Толщина хрусталика левого глаза",
		"79905-6": "Острота зрения правого глаза",
		"79906-4": "Острота зрения левого глаза",
		"79907-2": "Рефракция правого глаза",
		"79908-0": "Рефракция левого глаза",
		"79909-8": "Внутриглазное давление правого глаза",
		"79910-6": "Внутриглазное давление левого глаза",
	}

	for code, display := range biometryCodes {
		if strings.Contains(strings.ToLower(display), query) || strings.Contains(code, query) {
			results = append(results, domain.MedicalCodeSearchResult{
				Code:    code,
				Display: display,
				System:  "LOINC",
			})
		}
	}

	return results
}
