package router

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/doubleunion/accesscontrol/door"
	"github.com/doubleunion/accesscontrol/requests"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RunRouter() {
	signingKey := os.Getenv("ACCESS_CONTROL_SIGNING_KEY")
	if signingKey == "" {
		log.Fatal("signing key is missing")
	}

	door := door.New()

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/api/v1/status", func(c echo.Context) error {
		return jsonResponse(c, http.StatusOK, "door control available")
	})

	e.POST("/api/v1/unlock", func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return jsonResponse(c, http.StatusBadRequest, "error reading request body")
		}

		var request requests.UnlockRequest
		err = json.Unmarshal(body, &request)
		if err != nil {
			return jsonResponse(c, http.StatusBadRequest, "error unmarshaling request body")
		}

		token, ok := c.Get("user").(*jwt.Token)
		if !ok {
			return jsonResponse(c, http.StatusBadRequest, "error getting request token")
		}

		subject, err := token.Claims.GetSubject()
		if err != nil {
			return jsonResponse(c, http.StatusBadRequest, "error reading subject from token")
		}

		err = door.UnlockForDuration(time.Duration(request.Seconds)*time.Second, subject)
		if err != nil {
			return jsonResponse(c, http.StatusBadRequest, fmt.Sprintf("%+v", err))
		}

		return jsonResponse(c, http.StatusOK, "access granted")
	}, echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(signingKey),
	}))

	e.Logger.Fatal(e.Start(":8080"))
}

func jsonResponse(c echo.Context, code int, message string) error {
	return c.JSON(code, map[string]string{"message": message})
}
