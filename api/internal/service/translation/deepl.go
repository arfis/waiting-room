package translation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/arfis/waiting-room/internal/config"
)

// DeepLTranslationService handles translation using DeepL API
type DeepLTranslationService struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	cache      *TranslationCache
}

// TranslationRequest represents a request to DeepL API
type TranslationRequest struct {
	Text       []string `json:"text"`
	SourceLang string   `json:"source_lang"`
	TargetLang string   `json:"target_lang"`
}

// TranslationResponse represents the response from DeepL API
type TranslationResponse struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
}

// NewDeepLTranslationService creates a new DeepL translation service
func NewDeepLTranslationService(config config.DeepLConfig) *DeepLTranslationService {
	apiKey := config.APIKey
	if apiKey == "" {
		// Return nil if no API key is configured
		return nil
	}

	// Create cache with configurable settings
	// Max 10000 entries, 7 days expiration
	cache := NewTranslationCache(10000, 7*24*time.Hour)

	return &DeepLTranslationService{
		apiKey:  apiKey,
		baseURL: "https://api-free.deepl.com/v2/translate",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: cache,
	}
}

// Translate translates text from source language to target language
func (s *DeepLTranslationService) Translate(text string, sourceLang, targetLang string) (string, error) {
	if s == nil {
		return text, fmt.Errorf("DeepL service not configured - no API key provided")
	}

	// Convert language codes to DeepL format
	sourceLangCode := s.convertLanguageCode(sourceLang)
	targetLangCode := s.convertLanguageCode(targetLang)

	// Check cache first
	if cachedTranslation, found := s.cache.Get(text, sourceLangCode, targetLangCode); found {
		return cachedTranslation, nil
	}

	// Cache miss - make API call
	request := TranslationRequest{
		Text:       []string{text},
		SourceLang: sourceLangCode,
		TargetLang: targetLangCode,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return text, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return text, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "DeepL-Auth-Key "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return text, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return text, fmt.Errorf("DeepL API returned status %d", resp.StatusCode)
	}

	var response TranslationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return text, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Translations) == 0 {
		return text, fmt.Errorf("no translations returned")
	}

	translatedText := response.Translations[0].Text

	// Store in cache for future use
	s.cache.Set(text, translatedText, sourceLangCode, targetLangCode)

	return translatedText, nil
}

// TranslateService translates a service object
func (s *DeepLTranslationService) TranslateService(service map[string]interface{}, sourceLang, targetLang string) (map[string]interface{}, error) {
	if s == nil {
		return service, fmt.Errorf("DeepL service not configured")
	}

	translatedService := make(map[string]interface{})

	// Copy all fields first
	for key, value := range service {
		translatedService[key] = value
	}

	// Translate name field
	if name, ok := service["name"].(string); ok && name != "" {
		translatedName, err := s.Translate(name, sourceLang, targetLang)
		if err != nil {
			// If translation fails, keep original name
			translatedService["name"] = name
		} else {
			translatedService["name"] = translatedName
		}
	}

	// Translate description field if it exists
	if description, ok := service["description"].(string); ok && description != "" {
		translatedDescription, err := s.Translate(description, sourceLang, targetLang)
		if err != nil {
			// If translation fails, keep original description
			translatedService["description"] = description
		} else {
			translatedService["description"] = translatedDescription
		}
	}

	return translatedService, nil
}

// TranslateServices translates an array of services
func (s *DeepLTranslationService) TranslateServices(services []map[string]interface{}, sourceLang, targetLang string) ([]map[string]interface{}, error) {
	if s == nil {
		return services, fmt.Errorf("DeepL service not configured")
	}

	translatedServices := make([]map[string]interface{}, len(services))

	for i, service := range services {
		translatedService, err := s.TranslateService(service, sourceLang, targetLang)
		if err != nil {
			// If translation fails for one service, keep original
			translatedServices[i] = service
		} else {
			translatedServices[i] = translatedService
		}
	}

	return translatedServices, nil
}

// convertLanguageCode converts our language codes to DeepL format
func (s *DeepLTranslationService) convertLanguageCode(lang string) string {
	switch lang {
	case "en":
		return "EN"
	case "sk":
		return "SK"
	case "es":
		return "ES"
	case "fr":
		return "FR"
	case "de":
		return "DE"
	case "it":
		return "IT"
	case "pt":
		return "PT"
	case "ar":
		return "AR"
	case "zh":
		return "ZH"
	case "ja":
		return "JA"
	case "ko":
		return "KO"
	default:
		return "EN" // Default to English
	}
}

// IsConfigured returns true if the DeepL service is properly configured
func (s *DeepLTranslationService) IsConfigured() bool {
	return s != nil && s.apiKey != ""
}

// GetCacheStats returns translation cache statistics
func (s *DeepLTranslationService) GetCacheStats() map[string]interface{} {
	if s == nil || s.cache == nil {
		return map[string]interface{}{
			"error": "Translation service or cache not initialized",
		}
	}
	return s.cache.GetStats()
}

// LogCacheStats logs the current cache statistics
func (s *DeepLTranslationService) LogCacheStats() {
	if s != nil && s.cache != nil {
		s.cache.LogStats()
	}
}

// ClearCache clears all cached translations
func (s *DeepLTranslationService) ClearCache() {
	if s != nil && s.cache != nil {
		s.cache.Clear()
	}
}
