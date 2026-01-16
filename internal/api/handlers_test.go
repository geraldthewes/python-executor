package api

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/geraldthewes/python-executor/internal/storage"
	"github.com/geraldthewes/python-executor/pkg/client"
)

func TestBuildTarFromFiles(t *testing.T) {
	tests := []struct {
		name    string
		files   []client.CodeFile
		wantErr bool
	}{
		{
			name: "single file",
			files: []client.CodeFile{
				{Name: "main.py", Content: "print('hello')"},
			},
			wantErr: false,
		},
		{
			name: "multiple files",
			files: []client.CodeFile{
				{Name: "main.py", Content: "from helper import greet\ngreet()"},
				{Name: "helper.py", Content: "def greet():\n    print('hello')"},
			},
			wantErr: false,
		},
		{
			name:    "empty files",
			files:   []client.CodeFile{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tarData, err := buildTarFromFiles(tt.files)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildTarFromFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify tar contents
			tr := tar.NewReader(bytes.NewReader(tarData))
			fileCount := 0
			for {
				header, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("reading tar: %v", err)
				}

				// Find matching file
				var found bool
				for _, f := range tt.files {
					if f.Name == header.Name {
						found = true
						content, _ := io.ReadAll(tr)
						if string(content) != f.Content {
							t.Errorf("file %s content = %q, want %q", f.Name, string(content), f.Content)
						}
						break
					}
				}
				if !found {
					t.Errorf("unexpected file in tar: %s", header.Name)
				}
				fileCount++
			}

			if fileCount != len(tt.files) {
				t.Errorf("tar file count = %d, want %d", fileCount, len(tt.files))
			}
		})
	}
}

func TestExecuteEval_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Tests that only validate request parsing (don't need executor)
	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantErr    string
	}{
		{
			name:       "missing code and files",
			body:       client.SimpleExecRequest{},
			wantStatus: http.StatusBadRequest,
			wantErr:    "either 'code' or 'files' must be provided",
		},
		{
			name:       "invalid JSON",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid JSON",
		},
		{
			name: "code too large",
			body: client.SimpleExecRequest{
				Code: strings.Repeat("x", maxCodeSize+1),
			},
			wantStatus: http.StatusRequestEntityTooLarge,
			wantErr:    "exceeds limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create server with storage but nil executor
			// These tests only validate request parsing
			memStorage := storage.NewMemoryStorage()
			server := &Server{storage: memStorage}

			router := gin.New()
			router.POST("/eval", server.ExecuteEval)

			var body []byte
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/eval", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantErr != "" && !strings.Contains(w.Body.String(), tt.wantErr) {
				t.Errorf("response body = %q, want to contain %q", w.Body.String(), tt.wantErr)
			}
		})
	}
}
