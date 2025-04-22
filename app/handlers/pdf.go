package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

// PDFRequest represents the request structure for the PDF.co API
type PDFRequest struct {
	URL       string `json:"url"`
	Name      string `json:"name"`
	PaperSize string `json:"paperSize"`
	Async     bool   `json:"async"`
}

// PDFResponse represents the response from PDF.co API
type PDFResponse struct {
	URL string `json:"url"`
}

// GetPDF handles requests to convert a URL to PDF using PDF.co service
func GetPDF(c *fiber.Ctx) error {
	id := c.Params("id")

	// TODO: Move this to a service

	// Create the HTTP client with 5-minute timeout
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	// Create request body
	reqBody := PDFRequest{
		URL:       fmt.Sprintf("https://whale-app-nm9uv.ondigitalocean.app/%s", id),
		Name:      fmt.Sprintf("%s.pdf", id),
		PaperSize: "A4",
		Async:     false,
	}

	// Marshal the request body to JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create request body",
		})
	}

	// Create new request
	req, err := http.NewRequest("POST", "https://api.pdf.co/v1/pdf/convert/from/url", bytes.NewBuffer(jsonData))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create request",
		})
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", viper.GetString("PDFCO_API_KEY"))

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Request failed: %v", err),
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return c.Status(resp.StatusCode).JSON(fiber.Map{
			"error":  "PDF.co API request failed",
			"status": resp.Status,
			"body":   string(bodyBytes),
		})
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read response",
		})
	}

	var pdfResponse PDFResponse
	if err := json.Unmarshal(responseBody, &pdfResponse); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse response",
		})
	}

	// TODO: Service ends here

	// TODO: Download the PDF file and store it in S3

	return c.JSON(fiber.Map{
		"url": pdfResponse.URL,
	})
}
