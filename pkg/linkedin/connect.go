package linkedin

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/mithu/linkedin-bot/internal/storage"
	"github.com/mithu/linkedin-bot/pkg/stealth"
)

// ConnectOptions configures connection request behavior
type ConnectOptions struct {
	Message               string
	DelayBetweenRequests  int // seconds
	MaxRequestsPerSession int
}

// DefaultConnectOptions returns sensible defaults
func DefaultConnectOptions() *ConnectOptions {
	return &ConnectOptions{
		Message:               "", // No message by default
		DelayBetweenRequests:  30, // 30 seconds between requests
		MaxRequestsPerSession: 20, // Max 20 per session
	}
}

// SendConnectionRequest sends a connection request to a profile
func SendConnectionRequest(page *rod.Page, profileURL string, db *storage.Database, opts *ConnectOptions) error {
	if opts == nil {
		opts = DefaultConnectOptions()
	}

	// Check if already sent
	alreadySent, err := db.HasSentConnectionRequest(profileURL)
	if err != nil {
		return fmt.Errorf("failed to check database: %w", err)
	}

	if alreadySent {
		fmt.Printf("Skipping %s (already sent)\n", profileURL)
		return nil
	}

	fmt.Printf("Visiting profile: %s\n", profileURL)

	// Navigate to profile
	if err := page.Navigate(profileURL); err != nil {
		return fmt.Errorf("failed to navigate to profile: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for page load: %w", err)
	}

	stealth.RandomPageLoadDelay()

	// Record profile visit
	if err := db.RecordProfileVisit(profileURL); err != nil {
		fmt.Printf("Failed to record visit: %v\n", err)
	}

	// Human-like scrolling to view profile
	if err := scrollProfile(page); err != nil {
		fmt.Printf("Failed to scroll: %v\n", err)
	}

	// Find connect button
	connectButton, err := findConnectButton(page)
	if err != nil {
		return fmt.Errorf("failed to find connect button: %w", err)
	}

	// Click connect button with Bézier movement
	shape, err := connectButton.Shape()
	if err != nil {
		return fmt.Errorf("failed to get button box: %w", err)
	}

	centerX := (shape.Quads[0][0] + shape.Quads[0][2]) / 2
	centerY := (shape.Quads[0][1] + shape.Quads[0][5]) / 2

	fmt.Println("Clicking Connect button...")
	if err := stealth.MoveAndClick(page, centerX, centerY, stealth.DefaultMoveMouseOptions()); err != nil {
		return fmt.Errorf("failed to click connect button: %w", err)
	}

	time.Sleep(1 * time.Second)

	// Check if message modal appeared
	hasMessageModal, _, _ := page.Has(`div[role="dialog"]`)

	if hasMessageModal && opts.Message != "" {
		// Add a note
		if err := addConnectionNote(page, opts.Message); err != nil {
			fmt.Printf("Failed to add note: %v\n", err)
		}
	} else if hasMessageModal {
		// Send without note
		sendButton, err := page.Element(`button[aria-label="Send without a note"]`)
		if err != nil {
			// Try alternative selector
			sendButton, err = page.Element(`button.artdeco-button--primary`)
		}

		if err == nil {
			shape, _ := sendButton.Shape()
			centerX := (shape.Quads[0][0] + shape.Quads[0][2]) / 2
			centerY := (shape.Quads[0][1] + shape.Quads[0][5]) / 2
			stealth.MoveAndClick(page, centerX, centerY, stealth.DefaultMoveMouseOptions())
		}
	}

	time.Sleep(1 * time.Second)

	// Get profile name for database
	profileName, _ := getProfileName(page)

	// Save to database
	if err := db.CreateConnectionRequest(profileURL, profileName, opts.Message); err != nil {
		return fmt.Errorf("failed to save to database: %w", err)
	}

	fmt.Printf("Connection request sent to: %s\n", profileName)

	// Delay before next action
	if opts.DelayBetweenRequests > 0 {
		fmt.Printf("Waiting %d seconds before next action...\n", opts.DelayBetweenRequests)
		time.Sleep(time.Duration(opts.DelayBetweenRequests) * time.Second)
	}

	return nil
}

// BulkConnect sends connection requests to multiple profiles
func BulkConnect(page *rod.Page, profiles []SearchResult, db *storage.Database, opts *ConnectOptions) error {
	if opts == nil {
		opts = DefaultConnectOptions()
	}

	successCount := 0
	errorCount := 0
	skippedCount := 0

	fmt.Printf("Starting bulk connect to %d profiles...\n", len(profiles))

	for i, profile := range profiles {
		// Check session limit
		if opts.MaxRequestsPerSession > 0 && successCount >= opts.MaxRequestsPerSession {
			fmt.Printf("Reached session limit (%d requests)\n", opts.MaxRequestsPerSession)
			break
		}

		fmt.Printf("\n[%d/%d] Processing: %s\n", i+1, len(profiles), profile.ProfileName)

		// Check if already sent
		alreadySent, _ := db.HasSentConnectionRequest(profile.ProfileURL)
		if alreadySent {
			fmt.Println("Already sent, skipping...")
			skippedCount++
			continue
		}

		// Send connection request
		err := SendConnectionRequest(page, profile.ProfileURL, db, opts)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			errorCount++

			// Continue to next profile
			time.Sleep(5 * time.Second)
			continue
		}

		successCount++
	}

	fmt.Printf("\nBulk connect complete!\n")
	fmt.Printf("Sent: %d\n", successCount)
	fmt.Printf("Skipped: %d\n", skippedCount)
	fmt.Printf("Errors: %d\n", errorCount)

	return nil
}

// findConnectButton finds the connect button on a profile page
func findConnectButton(page *rod.Page) (*rod.Element, error) {
	// Try multiple selectors
	selectors := []string{
		`button[aria-label*="Connect"]`,
		`button.pvs-profile-actions__action:has-text("Connect")`,
		`button:has-text("Connect")`,
	}

	for _, selector := range selectors {
		btn, err := page.Element(selector)
		if err == nil {
			return btn, nil
		}
	}

	return nil, fmt.Errorf("connect button not found")
}

// scrollProfile scrolls through a profile page in a human-like manner
func scrollProfile(page *rod.Page) error {
	// Scroll down to view profile sections
	for i := 0; i < 3; i++ {
		if err := stealth.ScrollWithVariation(page, 400); err != nil {
			return err
		}
		stealth.RandomScrollDelay()
	}

	// Scroll back to top
	_, err := page.Eval(`() => window.scrollTo({ top: 0, behavior: 'smooth' })`)
	time.Sleep(1 * time.Second)

	return err
}

// addConnectionNote adds a personalized note to the connection request
func addConnectionNote(page *rod.Page, message string) error {
	// Find "Add a note" button
	addNoteButton, err := page.Element(`button[aria-label="Add a note"]`)
	if err != nil {
		return err
	}

	// Click add note button
	shape, _ := addNoteButton.Shape()
	centerX := (shape.Quads[0][0] + shape.Quads[0][2]) / 2
	centerY := (shape.Quads[0][1] + shape.Quads[0][5]) / 2
	stealth.MoveAndClick(page, centerX, centerY, stealth.DefaultMoveMouseOptions())

	time.Sleep(500 * time.Millisecond)

	// Find message textarea
	messageBox, err := page.Element(`textarea[name="message"]`)
	if err != nil {
		return err
	}

	// Type message
	if err := stealth.ClearAndType(page, messageBox, message, stealth.DefaultTypingOptions()); err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)

	// Click send button
	sendButton, err := page.Element(`button[aria-label="Send now"]`)
	if err != nil {
		sendButton, err = page.Element(`button.artdeco-button--primary`)
		if err != nil {
			return err
		}
	}

	shape, _ = sendButton.Shape()
	centerX = (shape.Quads[0][0] + shape.Quads[0][2]) / 2
	centerY = (shape.Quads[0][1] + shape.Quads[0][5]) / 2
	return stealth.MoveAndClick(page, centerX, centerY, stealth.DefaultMoveMouseOptions())
}

// getProfileName extracts the profile name from the page
func getProfileName(page *rod.Page) (string, error) {
	nameElement, err := page.Element(`h1.text-heading-xlarge`)
	if err != nil {
		return "", err
	}

	return nameElement.Text()
}
