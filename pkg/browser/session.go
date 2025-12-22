package browser

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Session struct {
	Cookies   []*proto.NetworkCookie `json:"cookies"`
	UserAgent string                 `json:"user_agent"`
	Timestamp time.Time              `json:"timestamp"`
}

func SaveSession(page *rod.Page, filename string) error {
	cookies, err := page.Cookies([]string{})
	if err != nil {
		return fmt.Errorf("failed to get cookies: %w", err)
	}
	userAgent, err := page.Eval(`() => navigator.userAgent`)
	if err != nil {
		return fmt.Errorf("failed to get user agent: %w", err)
	}
	ua := ""
	if userAgent.Value.Str() != "" {
		ua = userAgent.Value.Str()
	}
	session := Session{Cookies: cookies, UserAgent: ua, Timestamp: time.Now()}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}
	fmt.Printf("Session saved to: %s\n", filename)
	return nil
}

func LoadSession(page *rod.Page, filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("session file not found: %s", filename)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read session file: %w", err)
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("failed to unmarshal session: %w", err)
	}
	if time.Since(session.Timestamp) > 7*24*time.Hour {
		return fmt.Errorf("session expired (older than 7 days)")
	}
	cookieParams := make([]*proto.NetworkCookieParam, len(session.Cookies))
	for i, c := range session.Cookies {
		sourcePort := c.SourcePort
		cookieParams[i] = &proto.NetworkCookieParam{
			Name: c.Name, Value: c.Value, URL: "", Domain: c.Domain, Path: c.Path,
			Secure: c.Secure, HTTPOnly: c.HTTPOnly, SameSite: c.SameSite,
			Expires: c.Expires, Priority: c.Priority, SameParty: c.SameParty,
			SourceScheme: c.SourceScheme, SourcePort: &sourcePort,
		}
	}
	if err := page.SetCookies(cookieParams); err != nil {
		return fmt.Errorf("failed to set cookies: %w", err)
	}
	fmt.Printf("Session loaded from: %s\n", filename)
	return nil
}

func IsSessionValid(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return false
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return false
	}
	if time.Since(session.Timestamp) > 7*24*time.Hour {
		return false
	}
	if len(session.Cookies) == 0 {
		return false
	}
	return true
}

func ClearSession(filename string) error {
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove session file: %w", err)
	}
	fmt.Printf("🗑️  Session cleared: %s\n", filename)
	return nil
}

func VerifyLogin(page *rod.Page, checkURL string) (bool, error) {
	if err := page.Navigate(checkURL); err != nil {
		return false, fmt.Errorf("failed to navigate: %w", err)
	}
	if err := page.WaitLoad(); err != nil {
		return false, fmt.Errorf("failed to wait for page load: %w", err)
	}
	time.Sleep(2 * time.Second)
	currentURL := page.MustInfo().URL
	if strings.Contains(currentURL, "/login") || strings.Contains(currentURL, "/uas/") {
		return false, nil
	}
	hasNav, _, _ := page.Has(`nav[aria-label="Primary Navigation"]`)
	if hasNav {
		return true, nil
	}
	hasFeed, _, _ := page.Has(`div[role="main"]`)
	if hasFeed {
		return true, nil
	}
	return false, nil
}
