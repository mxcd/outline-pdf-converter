package converter

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"

	"net/http"

	"github.com/google/uuid"
)

func GetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20) // limit your max input length!
		if err != nil {
			http.Error(w, "Error parsing multipart form", http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error retrieving the file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error reading the file", http.StatusInternalServerError)
			return
		}

		zipReader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
		if err != nil {
			http.Error(w, "Error reading zip content", http.StatusInternalServerError)
			return
		}

		document := &Document{
			Markdown:    "",
			Attachments: []*Attachment{},
		}

		for _, zipFile := range zipReader.File {
			zf, err := zipFile.Open()
			if err != nil {
				http.Error(w, "Error opening zip file content", http.StatusInternalServerError)
				return
			}

			defer zf.Close()

			content, err := io.ReadAll(zf)
			if err != nil {
				http.Error(w, "Error reading zip file content", http.StatusInternalServerError)
				zf.Close()
				return
			}

			if zipFile.FileInfo().IsDir() {
				continue
			}

			if strings.HasSuffix(zipFile.Name, ".md") {
				if document.Markdown != "" {
					http.Error(w, "Multiple markdown files found", http.StatusBadRequest)
					return
				}
				document.Markdown = string(content)
				document.FileName = zipFile.Name
			} else {
				document.Attachments = append(document.Attachments, &Attachment{
					Name:    zipFile.Name,
					Content: content,
				})
			}

		}
		document.Print()

		processId := uuid.New()

		go document.ConvertToPDF(processId)

		if err != nil {
			http.Error(w, "Error converting to PDF", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(processId.String()))
	}
}
