package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func signTestJWT(t *testing.T, claims jwt.MapClaims, secret string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign JWT: %v", err)
	}
	return signed
}

func TestAuthMiddleware_EmitsAdditiveAuthUnauthorizedCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret-m1")

	r := gin.New()
	r.GET("/protected", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id":  c.GetString("user_id"),
			"role":     c.GetString("role"),
			"nickname": c.GetString("nickname"),
		})
	})

	expiredToken := signTestJWT(t, jwt.MapClaims{
		"user_id": "u-1",
		"role":    "learner",
		"exp":     time.Now().Add(-time.Hour).Unix(),
	}, "test-secret-m1")

	missingUserIDToken := signTestJWT(t, jwt.MapClaims{
		"role": "learner",
		"exp":  time.Now().Add(time.Hour).Unix(),
	}, "test-secret-m1")

	cases := []struct {
		name        string
		authHeader  string
		wantError   string
	}{
		{
			name:       "missing authorization header",
			authHeader: "",
			wantError:  "Authorization header is required",
		},
		{
			name:       "non-bearer scheme",
			authHeader: "Basic abc123",
			wantError:  "Authorization header must be Bearer token",
		},
		{
			name:       "expired token",
			authHeader: "Bearer " + expiredToken,
			wantError:  "Invalid or expired token",
		},
		{
			name:       "missing user_id claim",
			authHeader: "Bearer " + missingUserIDToken,
			wantError:  "Invalid user_id in token",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d (%s)", w.Code, w.Body.String())
			}
			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid json: %v", err)
			}
			if body["error"] != tc.wantError {
				t.Fatalf("expected error %q, got %v", tc.wantError, body["error"])
			}
			if body["code"] != "AUTH_UNAUTHORIZED" {
				t.Fatalf("expected code AUTH_UNAUTHORIZED, got %v", body["code"])
			}
		})
	}
}

func TestRequireRoles_EmitsAdditiveAuthForbiddenAndUnauthorizedCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret-m1")

	t.Run("missing role in context returns 401 AUTH_UNAUTHORIZED", func(t *testing.T) {
		r := gin.New()
		r.GET("/admin", RequireRoles("admin"), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
		var body map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		if body["error"] != "Unauthorized" || body["code"] != "AUTH_UNAUTHORIZED" {
			t.Fatalf("unexpected body: %+v", body)
		}
	})

	t.Run("non-string role returns 403 AUTH_FORBIDDEN", func(t *testing.T) {
		r := gin.New()
		r.GET("/admin", func(c *gin.Context) {
			c.Set("role", 123)
			c.Next()
		}, RequireRoles("admin"), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
		var body map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		if body["error"] != "Invalid role" || body["code"] != "AUTH_FORBIDDEN" {
			t.Fatalf("unexpected body: %+v", body)
		}
	})

	t.Run("insufficient role returns 403 AUTH_FORBIDDEN and admin passes", func(t *testing.T) {
		r := gin.New()
		r.GET("/admin", AuthMiddleware(), RequireRoles("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		learnerTok := signTestJWT(t, jwt.MapClaims{
			"user_id":  "u-1",
			"role":     "learner",
			"nickname": "rookie",
			"exp":      time.Now().Add(time.Hour).Unix(),
		}, "test-secret-m1")

		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+learnerTok)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
		var body map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		if body["error"] != "Forbidden: insufficient permissions" || body["code"] != "AUTH_FORBIDDEN" {
			t.Fatalf("unexpected body: %+v", body)
		}

		adminTok := signTestJWT(t, jwt.MapClaims{
			"user_id":  "u-2",
			"role":     "admin",
			"nickname": "boss",
			"exp":      time.Now().Add(time.Hour).Unix(),
		}, "test-secret-m1")

		req2 := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req2.Header.Set("Authorization", "Bearer "+adminTok)
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200 for admin, got %d (%s)", w2.Code, w2.Body.String())
		}
	})
}
