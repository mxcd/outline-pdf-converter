package converter

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/mxcd/go-config/config"
	"github.com/rs/zerolog/log"
)

type Document struct {
	Markdown    string        `json:"markdown"`
	FileName    string        `json:"fileName"`
	Attachments []*Attachment `json:"attachments"`
}

type Attachment struct {
	Name    string `json:"name"`
	Content []byte `json:"content"`
}

func (d *Document) Print() {
	log.Info().Msgf("Markdown file name: %s", d.FileName)
	for _, attachment := range d.Attachments {
		log.Info().Msgf("Attachment: %s", attachment.Name)
	}
}

func (d *Document) ConvertToPDF(processId uuid.UUID) error {
	// Define the HTML wrapper for the Markdown content
	htmlWrapper := `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>` + d.FileName + `</title>
    <style>
      img {
        max-width: 95%;
      }
      body {
        font-family: Arial, sans-serif;
      }
    </style>
  </head>
  <body>
    {{ toHTML "` + d.FileName + `" }}
  </body>
</html>`

	// Prepare the multipart form data
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add the HTML file to the form
	fw, err := w.CreateFormFile("files", "index.html")
	if err != nil {
		return fmt.Errorf("error creating form file for HTML: %w", err)
	}
	if _, err := fw.Write([]byte(htmlWrapper)); err != nil {
		return fmt.Errorf("error writing HTML content to form: %w", err)
	}

	fw, err = w.CreateFormFile("files", d.FileName)
	if err != nil {
		return fmt.Errorf("error creating form file for Markdown: %w", err)
	}
	rewrittenMarkdownContent := rewriteAttachmentURLs(d.Markdown)
	if _, err := fw.Write([]byte(rewrittenMarkdownContent)); err != nil {
		return fmt.Errorf("error writing Markdown content to form: %w", err)
	}

	// Add the attachments to the form
	for _, attachment := range d.Attachments {
		fw, err = w.CreateFormFile("files", attachment.Name)
		if err != nil {
			return fmt.Errorf("error creating form file for attachment: %w", err)
		}
		if _, err := fw.Write(attachment.Content); err != nil {
			return fmt.Errorf("error writing attachment content to form: %w", err)
		}
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("error closing multipart writer: %w", err)
	}

	gotenbergURL := config.Get().String("GOTENBERG_URL") + "/forms/chromium/convert/markdown"

	// Create and send the request
	req, err := http.NewRequest("POST", gotenbergURL, &b)
	if err != nil {
		return fmt.Errorf("error creating POST request to Gotenberg: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending POST request to Gotenberg: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gotenberg responded with status code: %d", resp.StatusCode)
	}

	// Read the response body (PDF file)
	pdf, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	pdfDocument := &PdfDocument{
		FileName: d.FileName + ".pdf",
		Content:  pdf,
	}

	documentCache.Add(processId, pdfDocument)

	return nil
}

func rewriteAttachmentURLs(markdown string) string {
	matcher := regexp.MustCompile(`attachments/([a-f0-9-]+)\.([a-z]+)`)
	return matcher.ReplaceAllString(markdown, "$1.$2")
}
