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

func TestParseErrorFromStderr(t *testing.T) {
	tests := []struct {
		name          string
		stderr        string
		wantErrorType string
		wantErrorLine int
	}{
		{
			name: "NameError",
			stderr: `Traceback (most recent call last):
  File "main.py", line 1, in <module>
    print(undefined_var)
NameError: name 'undefined_var' is not defined`,
			wantErrorType: "NameError",
			wantErrorLine: 1,
		},
		{
			name: "SyntaxError",
			stderr: `  File "main.py", line 3
    if True
          ^
SyntaxError: expected ':'`,
			wantErrorType: "SyntaxError",
			wantErrorLine: 3,
		},
		{
			name: "TypeError",
			stderr: `Traceback (most recent call last):
  File "main.py", line 5, in <module>
    result = add(1, "two")
TypeError: unsupported operand type(s) for +: 'int' and 'str'`,
			wantErrorType: "TypeError",
			wantErrorLine: 5,
		},
		{
			name: "IndexError with nested calls",
			stderr: `Traceback (most recent call last):
  File "main.py", line 10, in <module>
    main()
  File "main.py", line 7, in main
    print(items[5])
IndexError: list index out of range`,
			wantErrorType: "IndexError",
			wantErrorLine: 10, // First line number found
		},
		{
			name:          "empty stderr",
			stderr:        "",
			wantErrorType: "",
			wantErrorLine: 0,
		},
		{
			name:          "no error pattern",
			stderr:        "Some random output without error patterns",
			wantErrorType: "",
			wantErrorLine: 0,
		},
		{
			name: "ValueError",
			stderr: `Traceback (most recent call last):
  File "main.py", line 2, in <module>
    x = int("not a number")
ValueError: invalid literal for int() with base 10: 'not a number'`,
			wantErrorType: "ValueError",
			wantErrorLine: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorType, errorLine := parseErrorFromStderr(tt.stderr)
			if errorType != tt.wantErrorType {
				t.Errorf("errorType = %q, want %q", errorType, tt.wantErrorType)
			}
			if errorLine != tt.wantErrorLine {
				t.Errorf("errorLine = %d, want %d", errorLine, tt.wantErrorLine)
			}
		})
	}
}

func TestParseResultFromStdout(t *testing.T) {
	tests := []struct {
		name           string
		stdout         string
		wantStdout     string
		wantResult     *string
	}{
		{
			name:       "simple expression result",
			stdout:     "___PYEXEC_RESULT___\"4\"\n",
			wantStdout: "",
			wantResult: strPtr("4"),
		},
		{
			name:       "expression result with prior output",
			stdout:     "hello world\n___PYEXEC_RESULT___\"15\"\n",
			wantStdout: "hello world",
			wantResult: strPtr("15"),
		},
		{
			name:       "no result marker",
			stdout:     "hello world\n",
			wantStdout: "hello world\n",
			wantResult: nil,
		},
		{
			name:       "list result",
			stdout:     "___PYEXEC_RESULT___\"[1, 2, 3]\"\n",
			wantStdout: "",
			wantResult: strPtr("[1, 2, 3]"),
		},
		{
			name:       "string result with quotes",
			stdout:     "___PYEXEC_RESULT___\"'hello'\"\n",
			wantStdout: "",
			wantResult: strPtr("'hello'"),
		},
		{
			name:       "empty stdout",
			stdout:     "",
			wantStdout: "",
			wantResult: nil,
		},
		{
			name:       "result without trailing newline",
			stdout:     "___PYEXEC_RESULT___\"42\"",
			wantStdout: "",
			wantResult: strPtr("42"),
		},
		{
			name:       "multiple lines before result",
			stdout:     "line1\nline2\nline3\n___PYEXEC_RESULT___\"result\"\n",
			wantStdout: "line1\nline2\nline3",
			wantResult: strPtr("result"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStdout, gotResult := parseResultFromStdout(tt.stdout)
			if gotStdout != tt.wantStdout {
				t.Errorf("stdout = %q, want %q", gotStdout, tt.wantStdout)
			}
			if (gotResult == nil) != (tt.wantResult == nil) {
				t.Errorf("result nil = %v, want nil = %v", gotResult == nil, tt.wantResult == nil)
			}
			if gotResult != nil && tt.wantResult != nil && *gotResult != *tt.wantResult {
				t.Errorf("result = %q, want %q", *gotResult, *tt.wantResult)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
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
		{
			name: "invalid python version",
			body: client.SimpleExecRequest{
				Code:          "print('hi')",
				PythonVersion: "2.7",
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "unsupported python_version",
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
