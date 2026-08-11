package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"canvas-cli/internal/config"
)

// newTestClient builds a Client that talks to the given test-server base URL
// and skips the SSO login flow.
func newTestClient(baseURL string) *Client {
	jar, _ := cookiejar.New(nil)
	c := &Client{
		SiteURL: strings.TrimRight(baseURL, "/"),
		BaseURL: strings.TrimRight(baseURL, "/") + "/api/v1",
		Config:  &config.Config{APIURL: baseURL},
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
			Jar:     jar,
		},
		PerPage: 50,
	}
	c.loggedIn = true // bypass SSO for the test
	return c
}

// writeTempFile creates a temp file with known contents and returns its path.
func writeTempFile(t *testing.T, name, contents string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(contents), 0644); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	return p
}

// TestUploadFile_Direct201 exercises the happy path where the storage endpoint
// responds directly with 201 + the file object (inst-fs style).
func TestUploadFile_Direct201(t *testing.T) {
	const fileContents = "hello canvas attachment"
	var gotInit bool
	var gotUpload bool
	var uploadedFileField string

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/courses/1/assignments/2/submissions/self/files", func(w http.ResponseWriter, r *http.Request) {
		gotInit = true
		if r.Method != http.MethodPost {
			t.Errorf("init: expected POST, got %s", r.Method)
		}
		r.ParseForm()
		if r.Form.Get("name") != "report.pdf" {
			t.Errorf("init: expected name=report.pdf, got %q", r.Form.Get("name"))
		}
		if r.Form.Get("size") != fmt.Sprintf("%d", len(fileContents)) {
			t.Errorf("init: expected size=%d, got %q", len(fileContents), r.Form.Get("size"))
		}
		if r.Form.Get("content_type") != "application/pdf" {
			t.Errorf("init: expected content_type=application/pdf, got %q", r.Form.Get("content_type"))
		}
		host := "http://" + r.Host
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"upload_url":"%s/upload","upload_params":{"key":"abc123","filename":"report.pdf"}}`, host)
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		gotUpload = true
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("upload: parsing multipart: %v", err)
		}
		// upload_params must be present
		if r.FormValue("key") != "abc123" {
			t.Errorf("upload: expected key=abc123, got %q", r.FormValue("key"))
		}
		// file field must be present with the right contents
		f, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("upload: missing file field: %v", err)
		}
		defer f.Close()
		b, _ := io.ReadAll(f)
		uploadedFileField = string(b)
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":555,"display_name":"report.pdf","filename":"report.pdf","size":23}`)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(srv.URL)
	path := writeTempFile(t, "report.pdf", fileContents)

	res, err := c.UploadFile("/courses/1/assignments/2/submissions/self/files", path)
	if err != nil {
		t.Fatalf("UploadFile returned error: %v", err)
	}
	if !gotInit {
		t.Error("init endpoint was never called")
	}
	if !gotUpload {
		t.Error("upload endpoint was never called")
	}
	if uploadedFileField != fileContents {
		t.Errorf("uploaded contents mismatch: got %q", uploadedFileField)
	}
	if res.ID != 555 {
		t.Errorf("expected file ID 555, got %d", res.ID)
	}
	if res.DisplayName != "report.pdf" {
		t.Errorf("expected display_name report.pdf, got %q", res.DisplayName)
	}
}

// TestUploadFile_Redirect exercises the S3-style path where the storage
// endpoint returns a 3xx and the file object is fetched from the Location.
func TestUploadFile_Redirect(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/courses/1/assignments/2/submissions/self/files", func(w http.ResponseWriter, r *http.Request) {
		host := "http://" + r.Host
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"upload_url":"%s/s3upload","upload_params":{"key":"k","acl":"private"}}`, host)
	})

	mux.HandleFunc("/s3upload", func(w http.ResponseWriter, r *http.Request) {
		host := "http://" + r.Host
		http.Redirect(w, r, host+"/api/v1/files/777/confirm", http.StatusFound)
	})

	mux.HandleFunc("/api/v1/files/777/confirm", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("confirm: expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":777,"display_name":"notes.txt","filename":"notes.txt","size":5}`)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(srv.URL)
	path := writeTempFile(t, "notes.txt", "hello")

	res, err := c.UploadFile("/courses/1/assignments/2/submissions/self/files", path)
	if err != nil {
		t.Fatalf("UploadFile returned error: %v", err)
	}
	if res.ID != 777 {
		t.Errorf("expected file ID 777, got %d", res.ID)
	}
}

// TestUploadFile_MissingFile ensures a clear error when the path doesn't exist.
func TestUploadFile_MissingFile(t *testing.T) {
	c := newTestClient("http://127.0.0.1:0")
	_, err := c.UploadFile("/courses/1/assignments/2/submissions/self/files", "/no/such/file.pdf")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// TestUploadFile_InitParamOrdering verifies upload_params are written before
// the file part in the multipart body (Canvas/S3 require this ordering).
func TestUploadFile_ParamOrderingBeforeFile(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/courses/1/assignments/2/submissions/self/files", func(w http.ResponseWriter, r *http.Request) {
		host := "http://" + r.Host
		fmt.Fprintf(w, `{"upload_url":"%s/upload","upload_params":{"key":"ORDERKEY"}}`, host)
	})
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		body := string(raw)
		keyIdx := strings.Index(body, "ORDERKEY")
		fileIdx := strings.Index(body, `name="file"`)
		if keyIdx < 0 || fileIdx < 0 {
			t.Fatalf("could not find both fields in body")
		}
		if keyIdx > fileIdx {
			t.Errorf("upload_params must appear before the file part (key at %d, file at %d)", keyIdx, fileIdx)
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"id":1,"display_name":"a.txt"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(srv.URL)
	path := writeTempFile(t, "a.txt", "x")
	if _, err := c.UploadFile("/courses/1/assignments/2/submissions/self/files", path); err != nil {
		t.Fatalf("UploadFile error: %v", err)
	}
}
