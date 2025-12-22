package linkedin

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/mithu/linkedin-bot/pkg/stealth"
)

const (
	LinkedInSearchURL = "https://www.linkedin.com/search/results/people/"
)

// SearchResult represents a profile from search results
type SearchResult struct {
	ProfileURL  string
	ProfileName string
	Headline    string
}

// Search performs a LinkedIn people search and returns profile URLs
func Search(page *rod.Page, query string, maxResults int) ([]SearchResult, error) {
	fmt.Printf("Searching for: %s\n", query)

	// Build search URL
	searchURL := fmt.Sprintf("%s?keywords=%s", LinkedInSearchURL, query)

	// Navigate to search page
	if err := page.Navigate(searchURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to search page: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("failed to wait for page load: %w", err)
	}

	stealth.RandomPageLoadDelay()

	var results []SearchResult
	currentPage := 1

	for len(results) < maxResults {
		fmt.Printf("Processing page %d...\n", currentPage)

		// Human-like scrolling before extracting results
		if err := humanLikeScroll(page); err != nil {
			fmt.Printf("Failed to scroll: %v\n", err)
		}

		// Wait for search results to load
		time.Sleep(2 * time.Second)

		// Extract profile URLs from current page
		pageResults, err := extractSearchResults(page)
		if err != nil {
			return nil, fmt.Errorf("failed to extract search results: %w", err)
		}

		if len(pageResults) == 0 {
			fmt.Println("No more results found")
			break
		}

		results = append(results, pageResults...)
		fmt.Printf("Found %d profiles (total: %d)\n", len(pageResults), len(results))

		// Check if we have enough results
		if len(results) >= maxResults {
			results = results[:maxResults]
			break
		}

		// Try to go to next page
		hasNext, err := goToNextPage(page)
		if err != nil {
			fmt.Printf("Error navigating to next page: %v\n", err)
			break
		}

		if !hasNext {
			fmt.Println("No more pages available")
			break
		}

		currentPage++
		stealth.RandomPageLoadDelay()
	}

	fmt.Printf("Search complete! Found %d profiles\n", len(results))
	return results, nil
}

// extractSearchResults extracts profile information from the current search page
func extractSearchResults(page *rod.Page) ([]SearchResult, error) {
	var results []SearchResult

	// Find all search result items
	// LinkedIn uses different selectors, we'll try multiple
	selectors := []string{
		`li.reusable-search__result-container`,
		`div.entity-result`,
		`li[class*="search-result"]`,
	}

	var elements rod.Elements
	var err error

	for _, selector := range selectors {
		elements, err = page.Elements(selector)
		if err == nil && len(elements) > 0 {
			break
		}
	}

	if len(elements) == 0 {
		return results, nil
	}

	// Extract data from each result
	for _, element := range elements {
		// Find profile link
		link, err := element.Element(`a[href*="/in/"]`)
		if err != nil {
			continue
		}

		profileURL, err := link.Attribute("href")
		if err != nil || profileURL == nil {
			continue
		}

		// Clean URL (remove query parameters)
		cleanURL := strings.Split(*profileURL, "?")[0]

		// Get profile name
		nameElement, err := element.Element(`span[aria-hidden="true"]`)
		if err != nil {
			nameElement, err = element.Element(`span.entity-result__title-text a span`)
		}

		profileName := ""
		if err == nil {
			name, _ := nameElement.Text()
			profileName = strings.TrimSpace(name)
		}

		// Get headline
		headlineElement, err := element.Element(`div.entity-result__primary-subtitle`)
		headline := ""
		if err == nil {
			hl, _ := headlineElement.Text()
			headline = strings.TrimSpace(hl)
		}

		results = append(results, SearchResult{
			ProfileURL:  cleanURL,
			ProfileName: profileName,
			Headline:    headline,
		})
	}

	return results, nil
}

// humanLikeScroll scrolls the page in a human-like manner
func humanLikeScroll(page *rod.Page) error {
	// Get page height
	pageHeight, err := page.Eval(`() => document.body.scrollHeight`)
	if err != nil {
		return err
	}

	height := pageHeight.Value.Num()

	// Scroll in chunks
	scrolls := 3 + (int(height) / 1000) // More scrolls for longer pages
	if scrolls > 8 {
		scrolls = 8 // Cap at 8 scrolls
	}

	for i := 0; i < scrolls; i++ {
		// Scroll down
		scrollAmount := 300.0 + float64(i)*100.0
		if err := stealth.ScrollWithVariation(page, scrollAmount); err != nil {
			return err
		}

		// Random pause
		stealth.RandomScrollDelay()
	}

	// Scroll back to top
	_, err = page.Eval(`() => window.scrollTo({ top: 0, behavior: 'smooth' })`)
	time.Sleep(1 * time.Second)

	return err
}

// goToNextPage attempts to navigate to the next page of search results
func goToNextPage(page *rod.Page) (bool, error) {
	// Look for "Next" button
	nextButton, err := page.Element(`button[aria-label="Next"]`)
	if err != nil {
		// Try alternative selector
		nextButton, err = page.Element(`button.artdeco-pagination__button--next`)
		if err != nil {
			return false, nil // No next button found
		}
	}

	// Check if button is disabled
	disabled, err := nextButton.Attribute("disabled")
	if err == nil && disabled != nil {
		return false, nil // Button is disabled
	}

	// Click next button with human-like movement
	shape, err := nextButton.Shape()
	if err != nil {
		return false, err
	}

	centerX := (shape.Quads[0][0] + shape.Quads[0][2]) / 2
	centerY := (shape.Quads[0][1] + shape.Quads[0][5]) / 2

	if err := stealth.MoveAndClick(page, centerX, centerY, stealth.DefaultMoveMouseOptions()); err != nil {
		return false, err
	}

	// Wait for page to load
	time.Sleep(2 * time.Second)

	return true, nil
}
