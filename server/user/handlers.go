package user

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/config"
	"github.com/marcopiovanello/yt-dlp-web-ui/v3/server/rest" // Import for RespondWithErrorJSON
)

const TOKEN_COOKIE_NAME = "jwt-yt-dlp-webui"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.RespondWithErrorJSON(w, http.StatusBadRequest, "Invalid login request payload.", err) // Changed from StatusInternalServerError
		return
	}

	var (
		username = config.Instance().Username
		password = config.Instance().Password
	)

	if username != req.Username || password != req.Password {
		rest.RespondWithErrorJSON(w, http.StatusUnauthorized, "Invalid username or password.", nil) // Changed from StatusBadRequest
		return
	}

	expiresAt := time.Now().Add(time.Hour * 24 * 30)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"expiresAt": expiresAt,
		"username":  req.Username,
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		rest.RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to sign JWT token.", err)
		return
	}

	if err := json.NewEncoder(w).Encode(tokenString); err != nil {
		rest.RespondWithErrorJSON(w, http.StatusInternalServerError, "Failed to encode JWT token response.", err)
		return
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     TOKEN_COOKIE_NAME,
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now(),
		Value:    "",
		Path:     "/",
	}

	http.SetCookie(w, cookie)
}
