package converter

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

type PdfDocument struct {
	FileName string
	Content  []byte
}

var documentCache *expirable.LRU[uuid.UUID, *PdfDocument]

func GetDownloadHandler() http.HandlerFunc {
	documentCache = expirable.NewLRU[uuid.UUID, *PdfDocument](25, nil, time.Minute*10)
	return func(w http.ResponseWriter, r *http.Request) {
		documentID, err := uuid.Parse(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, "Invalid document ID", http.StatusBadRequest)
			return
		}

		document, ok := documentCache.Get(documentID)
		if !ok {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename="+document.FileName)
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(document.Content)
	}
}

func GetStatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		documentID, err := uuid.Parse(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, "Invalid document ID", http.StatusBadRequest)
			return
		}

		_, ok := documentCache.Get(documentID)
		if !ok {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
