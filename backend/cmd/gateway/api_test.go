package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/xiaoran/doutok/internal/config"
	"github.com/xiaoran/doutok/internal/pkg/auth"
	"github.com/xiaoran/doutok/internal/pkg/cache"
	"github.com/xiaoran/doutok/internal/pkg/logger"
	"github.com/xiaoran/doutok/internal/pkg/snowflake"
	"github.com/xiaoran/doutok/internal/repository"
	"github.com/xiaoran/doutok/internal/storage"
)

var testRouter http.Handler
var testToken string

func TestMain(m *testing.M) {
	// Setup
	cfg := config.Load()
	logger.Init("info", "console")
	snowflake.Init(1, 1)
	auth.InitJWT(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.ExpireHour)

	// Connect to real DB (need Docker running)
	if err := repository.InitDB(cfg.Database); err != nil {
		fmt.Println("SKIP: Database not available -", err)
		os.Exit(0)
	}
	defer repository.Close()

	// Redis (optional)
	cache.InitRedis(cfg.Redis)

	// MinIO (optional)
	storage.Init(cfg.MinIO)

	testRouter = setupRouter(cfg)

	os.Exit(m.Run())
}

// ==================== Helper ====================

func doRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

type apiResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func parseResp(w *httptest.ResponseRecorder) apiResponse {
	var resp apiResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp
}

// ==================== Tests ====================

func TestHealthCheck(t *testing.T) {
	w := doRequest("GET", "/api/v1/health", nil, "")
	if w.Code != 200 {
		t.Fatalf("health check failed: %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "ok") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestRegister(t *testing.T) {
	username := fmt.Sprintf("test_%d", snowflake.GenID()%100000)
	body := map[string]string{
		"username": username,
		"password": "test123456",
		"nickname": "Test User",
	}

	w := doRequest("POST", "/api/v1/auth/register", body, "")
	resp := parseResp(w)

	if resp.Code != 0 {
		t.Fatalf("register failed: code=%d msg=%s", resp.Code, resp.Msg)
	}

	// Extract token
	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	token, ok := data["token"].(string)
	if !ok || token == "" {
		t.Fatal("no token in register response")
	}
	testToken = token
	t.Logf("registered: %s, token: %s...", username, token[:20])
}

func TestLogin(t *testing.T) {
	// First register a user
	username := fmt.Sprintf("login_%d", snowflake.GenID()%100000)
	doRequest("POST", "/api/v1/auth/register", map[string]string{
		"username": username,
		"password": "pass123456",
		"nickname": "Login Test",
	}, "")

	// Then login
	w := doRequest("POST", "/api/v1/auth/login", map[string]string{
		"username": username,
		"password": "pass123456",
	}, "")
	resp := parseResp(w)

	if resp.Code != 0 {
		t.Fatalf("login failed: code=%d msg=%s", resp.Code, resp.Msg)
	}

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	if data["token"] == nil || data["token"] == "" {
		t.Fatal("no token in login response")
	}
	testToken = data["token"].(string)
}

func TestLoginWrongPassword(t *testing.T) {
	w := doRequest("POST", "/api/v1/auth/login", map[string]string{
		"username": "nonexistent_user_xyz",
		"password": "wrong",
	}, "")
	resp := parseResp(w)

	if resp.Code == 0 {
		t.Fatal("login should fail with wrong credentials")
	}
}

func TestGetMe(t *testing.T) {
	if testToken == "" {
		t.Skip("no token, run TestLogin first")
	}

	w := doRequest("GET", "/api/v1/me", nil, testToken)
	resp := parseResp(w)

	if resp.Code != 0 {
		t.Fatalf("get me failed: code=%d msg=%s", resp.Code, resp.Msg)
	}

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	if data["username"] == nil {
		t.Fatal("no username in /me response")
	}
}

func TestFeed(t *testing.T) {
	w := doRequest("GET", "/api/v1/feed", nil, "")
	resp := parseResp(w)

	if resp.Code != 0 {
		t.Fatalf("feed failed: code=%d msg=%s", resp.Code, resp.Msg)
	}
}

func TestUploadVideo(t *testing.T) {
	if testToken == "" {
		t.Skip("no token")
	}

	// Create multipart form with a small fake video file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add title field
	writer.WriteField("title", "Test Video Upload")
	writer.WriteField("description", "Integration test video")
	writer.WriteField("tags", "#test")

	// Add a fake video file
	part, err := writer.CreateFormFile("video", "test.mp4")
	if err != nil {
		t.Fatal(err)
	}
	// Write some fake bytes (not a real video, but tests the upload path)
	part.Write([]byte("fake video content for testing"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/videos", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+testToken)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	resp := parseResp(w)
	if resp.Code != 0 {
		t.Fatalf("upload failed: code=%d msg=%s body=%s", resp.Code, resp.Msg, w.Body.String())
	}

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	if data["video_id"] == nil {
		t.Fatal("no video_id in upload response")
	}
	t.Logf("uploaded video_id: %v, play_url: %v", data["video_id"], data["play_url"])
}

func TestFeedAfterUpload(t *testing.T) {
	w := doRequest("GET", "/api/v1/feed", nil, testToken)
	resp := parseResp(w)

	if resp.Code != 0 {
		t.Fatalf("feed failed: code=%d msg=%s", resp.Code, resp.Msg)
	}

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	videos, _ := data["videos"].([]interface{})
	t.Logf("feed has %d videos", len(videos))
}

func TestLikeVideo(t *testing.T) {
	if testToken == "" {
		t.Skip("no token")
	}

	// First get a video from feed
	w := doRequest("GET", "/api/v1/feed", nil, testToken)
	resp := parseResp(w)
	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	videos, _ := data["videos"].([]interface{})
	if len(videos) == 0 {
		t.Skip("no videos to like")
	}

	video := videos[0].(map[string]interface{})
	videoID := int64(video["id"].(float64))

	// Like it
	w = doRequest("POST", fmt.Sprintf("/api/v1/videos/%d/like", videoID), nil, testToken)
	resp = parseResp(w)
	if resp.Code != 0 {
		t.Fatalf("like failed: code=%d msg=%s", resp.Code, resp.Msg)
	}

	// Unlike it
	w = doRequest("DELETE", fmt.Sprintf("/api/v1/videos/%d/like", videoID), nil, testToken)
	resp = parseResp(w)
	if resp.Code != 0 {
		t.Fatalf("unlike failed: code=%d msg=%s", resp.Code, resp.Msg)
	}
}

func TestListLiveRooms(t *testing.T) {
	w := doRequest("GET", "/api/v1/lives", nil, "")
	resp := parseResp(w)
	if resp.Code != 0 {
		t.Fatalf("list lives failed: code=%d msg=%s", resp.Code, resp.Msg)
	}
}

func TestSearch(t *testing.T) {
	w := doRequest("GET", "/api/v1/search?q=test&type=video", nil, "")
	resp := parseResp(w)
	if resp.Code != 0 {
		t.Fatalf("search failed: code=%d msg=%s", resp.Code, resp.Msg)
	}
}
