package linkedin

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/mithu/linkedin-bot/pkg/browser"
	"github.com/mithu/linkedin-bot/pkg/stealth"
)

const (
	LinkedInLoginURL = "https://www.linkedin.com/login"
	LinkedInFeedURL  = "https://www.linkedin.com/feed/"
)

// Login performs LinkedIn login with human-like behavior
func Login(page *rod.Page, email, password string) error {
	fmt.Println("Starting LinkedIn login...")

	// Navigate to login page
	if err := page.Navigate(LinkedInLoginURL); err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for page load: %w", err)
	}

	stealth.RandomPageLoadDelay()

	// Find email input
	emailInput, err := page.Element("#username")
	if err != nil {
		return fmt.Errorf("failed to find email input: %w", err)
	}

	// Type email with human-like behavior
	fmt.Println("Entering email...")
	if err := stealth.ClearAndType(page, emailInput, email, stealth.DefaultTypingOptions()); err != nil {
		return fmt.Errorf("failed to type email: %w", err)
	}

	stealth.RandomDelay(300, 800)

	// Find password input
	passwordInput, err := page.Element("#password")
	if err != nil {
		return fmt.Errorf("failed to find password input: %w", err)
	}

	// Type password with human-like behavior
	fmt.Println("🔑 Entering password...")
	if err := stealth.ClearAndType(page, passwordInput, password, stealth.DefaultTypingOptions()); err != nil {
		return fmt.Errorf("failed to type password: %w", err)
	}

	stealth.RandomDelay(500, 1200)

	// Find and click login button
	loginButton, err := page.Element(`button[type="submit"]`)
	if err != nil {
		return fmt.Errorf("failed to find login button: %w", err)
	}

	// Get button position and click with Bézier movement
	shape, err := loginButton.Shape()
	if err != nil {
		return fmt.Errorf("failed to get button box: %w", err)
	}

	centerX := (shape.Quads[0][0] + shape.Quads[0][2]) / 2
	centerY := (shape.Quads[0][1] + shape.Quads[0][5]) / 2

	fmt.Println("Clicking login button...")
	if err := stealth.MoveAndClick(page, centerX, centerY, stealth.DefaultMoveMouseOptions()); err != nil {
		return fmt.Errorf("failed to click login button: %w", err)
	}

	// Wait for navigation
	time.Sleep(3 * time.Second)

	// Check for 2FA or verification
	currentURL := page.MustInfo().URL

	if contains(currentURL, "/checkpoint") || contains(currentURL, "/challenge") {
		fmt.Println("2FA or verification detected!")
		if err := browser.WaitForUserAction("Please complete 2FA/verification in the browser", 300); err != nil {
			return err
		}
	}

	// Wait for page to load after login
	time.Sleep(2 * time.Second)

	// Verify login success
	currentURL = page.MustInfo().URL
	if contains(currentURL, "/feed") || contains(currentURL, "/mynetwork") {
		fmt.Println("Login successful!")
		return nil
	}

	// Check if still on login page (login failed)
	if contains(currentURL, "/login") {
		return fmt.Errorf("login failed - still on login page")
	}

	fmt.Println("Login completed!")
	return nil
}

// LoginWithSession attempts to login using saved session, falls back to credentials
func LoginWithSession(page *rod.Page, sessionFile, email, password string) error {
	// Try to load session
	if browser.IsSessionValid(sessionFile) {
		fmt.Println("Loading saved session...")

		// Navigate to LinkedIn first
		if err := page.Navigate(LinkedInFeedURL); err != nil {
			return fmt.Errorf("failed to navigate: %w", err)
		}

		// Load session
		if err := browser.LoadSession(page, sessionFile); err != nil {
			fmt.Printf("Failed to load session: %v\n", err)
		} else {
			// Refresh page to apply cookies
			if err := page.Reload(); err != nil {
				return fmt.Errorf("failed to reload page: %w", err)
			}

			if err := page.WaitLoad(); err != nil {
				return fmt.Errorf("failed to wait for page load: %w", err)
			}

			time.Sleep(2 * time.Second)

			// Verify login
			isLoggedIn, err := browser.VerifyLogin(page, LinkedInFeedURL)
			if err != nil {
				fmt.Printf("Failed to verify login: %v\n", err)
			} else if isLoggedIn {
				fmt.Println("Session is valid, logged in successfully!")
				return nil
			}
		}
	}

	// Session invalid or doesn't exist, perform normal login
	fmt.Println("Session invalid or not found, performing login...")
	if err := Login(page, email, password); err != nil {
		return err
	}

	// Save session after successful login
	fmt.Println("Saving session...")
	if err := browser.SaveSession(page, sessionFile); err != nil {
		fmt.Printf("Failed to save session: %v\n", err)
	}

	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
