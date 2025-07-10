package router

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

var (
	localInternetAddresses      []net.IP
	localInternetAddressesMutex sync.RWMutex
)

func beginLocalIPMonitoring() {
	updateLocalIPAddresses()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			updateLocalIPAddresses()
		}
	}()
}

func updateLocalIPAddresses() {
	ips, err := net.LookupIP("doorcontrol.doubleunion.org")
	if err != nil {
		log.Printf("Error in updateLocalIPAddresses: %v", err)
	}

	localInternetAddressesMutex.Lock()
	defer localInternetAddressesMutex.Unlock()
	localInternetAddresses = ips

	log.Printf("Updated local IP addresses to: %v", ips)
}

func requireLocalNetworkMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the remote address from the request
		remoteAddr := c.Request().RemoteAddr

		// Parse the IP address
		ip, _, err := net.SplitHostPort(remoteAddr)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to parse IP address")
		}

		localInternetAddressesMutex.RLock()
		defer localInternetAddressesMutex.RUnlock()
		for _, allowedIP := range localInternetAddresses {
			if ip == allowedIP.String() {
				// Continue to the next middleware or route handler
				return next(c)
			}
		}

		// Fail the request, there was no matching IP
		return jsonResponse(c, http.StatusForbidden, "requests not allowed from remote hosts")
	}
}
