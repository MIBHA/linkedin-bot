package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-rod/rod"
	"github.com/mithu/linkedin-bot/internal/config"
	"github.com/mithu/linkedin-bot/internal/storage"
	"github.com/mithu/linkedin-bot/pkg/browser"
	"github.com/mithu/linkedin-bot/pkg/linkedin"
)

func main() {
	fmt.Println("LinkedIn Automation Bot")
	fmt.Println("============================")

	// Load configuration
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	dbConfig := &storage.Config{
		DBPath: cfg.DatabasePath,
	}
	db, err := storage.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize browser
	browserConfig := browser.DefaultConfig()
	browserConfig.Headless = cfg.Headless
	browserConfig.SlowMotion = time.Duration(cfg.SlowMotion) * time.Millisecond

	browserManager, err := browser.NewManager(browserConfig)
	if err != nil {
		log.Fatalf("Failed to create browser manager: %v", err)
	}

	// Launch browser
	if err := browserManager.Launch(); err != nil {
		log.Fatalf("Failed to launch browser: %v", err)
	}
	defer browserManager.Close()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\nShutting down gracefully...")

		// Save session before exit
		if page := browserManager.GetBrowser().MustPage(""); page != nil {
			browser.SaveSession(page, cfg.SessionFile)
		}

		browserManager.Close()
		db.Close()
		os.Exit(0)
	}()

	// Create new page
	page, err := browserManager.NewPage("")
	if err != nil {
		log.Fatalf("Failed to create page: %v", err)
	}

	// Login
	fmt.Println("\nLogging in to LinkedIn...")
	if err := linkedin.LoginWithSession(page, cfg.SessionFile, cfg.LinkedInEmail, cfg.LinkedInPassword); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	// Display statistics
	displayStats(db)

	// Main menu
	for {
		fmt.Println("\nMain Menu")
		fmt.Println("1. Search and Connect")
		fmt.Println("2. View Statistics")
		fmt.Println("3. Exit")
		fmt.Print("\nSelect an option: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			runSearchAndConnect(page, db, cfg)
		case 2:
			displayStats(db)
		case 3:
			fmt.Println("Goodbye!")

			// Save session
			browser.SaveSession(page, cfg.SessionFile)
			return
		default:
			fmt.Println("Invalid option")
		}
	}
}

func runSearchAndConnect(page *rod.Page, db *storage.Database, cfg *config.Config) {
	// Get search query
	query := cfg.SearchQuery
	if query == "" {
		fmt.Print("\nEnter search query: ")
		fmt.Scanln(&query)
	}

	if query == "" {
		fmt.Println("Search query cannot be empty")
		return
	}

	// Search for profiles
	results, err := linkedin.Search(page, query, cfg.MaxResults)
	if err != nil {
		fmt.Printf("Search failed: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("No results found")
		return
	}

	// Ask for confirmation
	fmt.Printf("\nFound %d profiles. Send connection requests? (y/n): ", len(results))
	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "y" && confirm != "Y" {
		fmt.Println("Cancelled")
		return
	}

	// Get connection message
	fmt.Print("\nEnter connection message (press Enter for no message): ")
	var message string
	fmt.Scanln(&message)

	// Configure connect options
	opts := &linkedin.ConnectOptions{
		Message:               message,
		DelayBetweenRequests:  cfg.DelayBetweenActions,
		MaxRequestsPerSession: cfg.MaxConnectionsPerDay,
	}

	// Send connection requests
	if err := linkedin.BulkConnect(page, results, db, opts); err != nil {
		fmt.Printf("Bulk connect failed: %v\n", err)
		return
	}

	// Display updated statistics
	displayStats(db)
}

func displayStats(db *storage.Database) {
	stats, err := db.GetConnectionRequestStats()
	if err != nil {
		fmt.Printf("Failed to get statistics: %v\n", err)
		return
	}

	fmt.Println("\nStatistics")
	fmt.Println("─────────────────────────────")
	fmt.Printf("Total Requests:    %d\n", stats["total"])
	fmt.Printf("Pending:      %d\n", stats["pending"])
	fmt.Printf("Accepted:     %d\n", stats["accepted"])
	fmt.Printf("Rejected:     %d\n", stats["rejected"])
	fmt.Printf("Sent Today:        %d\n", stats["today"])
	fmt.Println("─────────────────────────────")
}
