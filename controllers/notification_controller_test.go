package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cybersim/models"

	"github.com/gin-gonic/gin"
)

func TestTemplatedNotificationsAndLegacyFallbackDTO(t *testing.T) {
	created := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)

	approved := newTemplatedNotification(
		"user-1",
		"Specialist Application Approved",
		"Congratulations! Your application has been approved. You now have full specialist permissions.",
		"specialist_approved",
		nil,
		"/specialist/reviews",
	)
	approved.NotificationID = "notif-1"
	approved.CreatedAt = created

	approvedDTO := notificationToDTO(approved)
	if approvedDTO.TemplateKey != "specialist_approved" {
		t.Fatalf("TemplateKey = %q, want specialist_approved", approvedDTO.TemplateKey)
	}
	if approvedDTO.Title != "Specialist Application Approved" {
		t.Fatalf("canonical English Title altered: %q", approvedDTO.Title)
	}

	reply := newTemplatedNotification(
		"user-2",
		"New reply to your comment",
		"alice replied to your comment on 'DNS Tunneling Notes'",
		"comment_reply",
		map[string]any{"actor": "alice", "thread_title": "DNS Tunneling Notes"},
		"/discussions/t1#comment-c2",
	)
	reply.NotificationID = "notif-2"
	reply.CreatedAt = created

	raw, err := json.Marshal(notificationToDTO(reply))
	if err != nil {
		t.Fatalf("json.Marshal(replyDTO) error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}
	if decoded["template_key"] != "comment_reply" {
		t.Fatalf("decoded template_key = %v, want comment_reply", decoded["template_key"])
	}
	params, ok := decoded["template_params"].(map[string]any)
	if !ok || params["actor"] != "alice" || params["thread_title"] != "DNS Tunneling Notes" {
		t.Fatalf("unexpected template_params = %#v", decoded["template_params"])
	}

	legacy := models.UserNotification{
		NotificationID: "notif-legacy",
		UserID:         "user-3",
		Title:          "Legacy Notification",
		Message:        "Pre-Spec-11 notification body",
		CreatedAt:      created,
	}
	legacyDTO := notificationToDTO(legacy)
	if legacyDTO.TemplateKey != "" || legacyDTO.Title != "Legacy Notification" {
		t.Fatalf("legacy DTO mismatch: %+v", legacyDTO)
	}
}

func TestDecodeAnyJSONMap(t *testing.T) {
	rawBytes := []byte(`{"th":{"title":"การทดสอบ"}}`)
	got, ok := decodeAnyJSONMap(rawBytes).(map[string]any)
	if !ok || got["th"] == nil {
		t.Fatalf("expected decoded map from []byte, got %#v", decodeAnyJSONMap(rawBytes))
	}

	rawJSON := json.RawMessage(`{"th":{"title":"raw"}}`)
	gotRaw, ok := decodeAnyJSONMap(rawJSON).(map[string]any)
	if !ok || gotRaw["th"] == nil {
		t.Fatalf("expected decoded map from json.RawMessage, got %#v", decodeAnyJSONMap(rawJSON))
	}

	if decodeAnyJSONMap([]byte("not-json")) != nil {
		t.Fatal("expected nil on invalid JSON bytes")
	}
	if decodeAnyJSONMap("") != nil {
		t.Fatal("expected nil on empty string")
	}
}

func TestCommunityAndNotificationControllerErrorCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	nc := NewNotificationController(nil)
	tc := NewThreadController(nil)
	cc := NewCommentController(nil)
	uc := NewUpvoteController(nil)
	sc := NewSpecialistController(nil)

	t.Run("notifications require auth with canonical code", func(t *testing.T) {
		for _, handler := range []gin.HandlerFunc{nc.ListNotifications, nc.MarkAsRead, nc.MarkAllAsRead} {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
			handler(c)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", w.Code)
			}
			var body map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["code"] != "AUTH_UNAUTHORIZED" {
				t.Fatalf("expected code AUTH_UNAUTHORIZED, got %v", body["code"])
			}
		}
	})

	t.Run("thread creation validates short title and content with canonical codes", func(t *testing.T) {
		cases := []struct {
			payload  string
			wantCode string
		}{
			{`{"title":"abc","content":"This content is certainly longer than twenty characters."}`, "THREAD_INVALID_PAYLOAD"},
			{`{"title":"   a   ","content":"This content is certainly longer than twenty characters."}`, "THREAD_TITLE_TOO_SHORT"},
			{`{"title":"Valid Title","content":"     short body     "}`, "THREAD_CONTENT_TOO_SHORT"},
		}
		for _, tcCase := range cases {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("user_id", "user-1")
			c.Request = httptest.NewRequest(http.MethodPost, "/api/threads", strings.NewReader(tcCase.payload))
			c.Request.Header.Set("Content-Type", "application/json")
			tc.CreateThread(c)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for %s, got %d (%s)", tcCase.payload, w.Code, w.Body.String())
			}
			var body map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["code"] != tcCase.wantCode || body["error"] == "" {
				t.Fatalf("expected code %q + error, got %v", tcCase.wantCode, body)
			}
		}
	})

	t.Run("comments and upvotes return canonical codes without auth", func(t *testing.T) {
		for _, handler := range []gin.HandlerFunc{
			cc.CreateComment, cc.UpdateComment, cc.DeleteComment, cc.TogglePinComment, cc.ToggleHideComment,
			uc.ToggleThreadUpvote, uc.ToggleCommentUpvote, sc.Apply, sc.GetApplicationStatus,
		} {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			handler(c)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", w.Code)
			}
			var body map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["code"] != "AUTH_UNAUTHORIZED" {
				t.Fatalf("expected AUTH_UNAUTHORIZED, got %v", body["code"])
			}
		}
	})
}
