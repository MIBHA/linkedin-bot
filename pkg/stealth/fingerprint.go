package stealth

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// FingerprintConfig holds configuration for browser fingerprint masking
type FingerprintConfig struct {
	UserAgent    string
	Viewport     *proto.EmulationSetDeviceMetricsOverride
	Language     string
	Timezone     string
	WebRTCPolicy string
	DisableWebGL bool
	CanvasNoise  bool
}

// UserAgents contains a list of realistic user agents
var UserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
}

// CommonViewports contains realistic viewport dimensions
var CommonViewports = []struct {
	Width  int
	Height int
}{
	{1920, 1080},
	{1366, 768},
	{1536, 864},
	{1440, 900},
	{1280, 720},
}

// NewRandomFingerprintConfig generates a random but realistic fingerprint configuration
func NewRandomFingerprintConfig() *FingerprintConfig {
	rand.Seed(time.Now().UnixNano())

	// Select random user agent
	userAgent := UserAgents[rand.Intn(len(UserAgents))]

	// Select random viewport
	viewport := CommonViewports[rand.Intn(len(CommonViewports))]

	// Add slight randomization to viewport (±50 pixels)
	viewportWidth := viewport.Width + rand.Intn(100) - 50
	viewportHeight := viewport.Height + rand.Intn(100) - 50

	return &FingerprintConfig{
		UserAgent: userAgent,
		Viewport: &proto.EmulationSetDeviceMetricsOverride{
			Width:             viewportWidth,
			Height:            viewportHeight,
			DeviceScaleFactor: 1,
			Mobile:            false,
		},
		Language:     "en-US,en;q=0.9",
		Timezone:     "America/New_York",
		WebRTCPolicy: "default_public_interface_only",
		DisableWebGL: false,
		CanvasNoise:  true,
	}
}

// ApplyFingerprint applies fingerprint masking to a Rod page
func ApplyFingerprint(page *rod.Page, config *FingerprintConfig) error {
	if config == nil {
		config = NewRandomFingerprintConfig()
	}

	// Set user agent
	if err := page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: config.UserAgent,
	}); err != nil {
		return fmt.Errorf("failed to set user agent: %w", err)
	}

	// Set viewport
	if config.Viewport != nil {
		if err := page.SetViewport(config.Viewport); err != nil {
			return fmt.Errorf("failed to set viewport: %w", err)
		}
	}

	// Inject scripts to mask automation detection
	if err := injectStealthScripts(page, config); err != nil {
		return fmt.Errorf("failed to inject stealth scripts: %w", err)
	}

	return nil
}

// injectStealthScripts injects JavaScript to remove automation indicators
func injectStealthScripts(page *rod.Page, config *FingerprintConfig) error {
	// Script to remove webdriver property and other automation indicators
	stealthScript := `
		// Remove webdriver property
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined
		});
		
		// Override plugins to appear more realistic
		Object.defineProperty(navigator, 'plugins', {
			get: () => [
				{
					0: {type: "application/x-google-chrome-pdf", suffixes: "pdf", description: "Portable Document Format"},
					description: "Portable Document Format",
					filename: "internal-pdf-viewer",
					length: 1,
					name: "Chrome PDF Plugin"
				},
				{
					0: {type: "application/pdf", suffixes: "pdf", description: "Portable Document Format"},
					description: "Portable Document Format",
					filename: "mhjfbmdgcfjbbpaeojofohoefgiehjai",
					length: 1,
					name: "Chrome PDF Viewer"
				}
			]
		});
		
		// Override languages
		Object.defineProperty(navigator, 'languages', {
			get: () => ['en-US', 'en']
		});
		
		// Override permissions
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({ state: Notification.permission }) :
				originalQuery(parameters)
		);
		
		// Add chrome runtime (makes it look like a real Chrome browser)
		window.chrome = {
			runtime: {}
		};
		
		// Override the toString of functions to hide modifications
		const originalToString = Function.prototype.toString;
		Function.prototype.toString = function() {
			if (this === window.navigator.permissions.query) {
				return 'function query() { [native code] }';
			}
			return originalToString.call(this);
		};
	`

	// Canvas fingerprint noise injection
	if config.CanvasNoise {
		stealthScript += `
		// Add noise to canvas fingerprinting
		const originalGetContext = HTMLCanvasElement.prototype.getContext;
		HTMLCanvasElement.prototype.getContext = function(type, ...args) {
			const context = originalGetContext.call(this, type, ...args);
			
			if (type === '2d') {
				const originalFillText = context.fillText;
				context.fillText = function(...args) {
					// Add tiny random noise
					const noise = Math.random() * 0.0001;
					context.globalAlpha = 1 - noise;
					return originalFillText.apply(this, args);
				};
			}
			
			return context;
		};
		`
	}

	// WebGL fingerprint masking
	if !config.DisableWebGL {
		stealthScript += `
		// Mask WebGL fingerprinting
		const getParameter = WebGLRenderingContext.prototype.getParameter;
		WebGLRenderingContext.prototype.getParameter = function(parameter) {
			// Randomize UNMASKED_VENDOR_WEBGL and UNMASKED_RENDERER_WEBGL
			if (parameter === 37445) {
				return 'Intel Inc.';
			}
			if (parameter === 37446) {
				return 'Intel Iris OpenGL Engine';
			}
			return getParameter.call(this, parameter);
		};
		`
	}

	// Wrap in IIFE to prevent scope issues and ensure correct execution
	finalScript := fmt.Sprintf(`(() => {
%s
})();`, stealthScript)

	// Inject the script so it runs on every page load (persistence)
	_, err := page.EvalOnNewDocument(finalScript)
	return err
}

// SetupStealthBrowser configures a browser instance with anti-detection measures
func SetupStealthBrowser(browser *rod.Browser) error {
	// This function can be used to set browser-level configurations
	// Most stealth measures are applied per-page via ApplyFingerprint
	return nil
}

// RandomDelay adds a random human-like delay
func RandomDelay(minMs, maxMs int) {
	delay := minMs + rand.Intn(maxMs-minMs)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// RandomScrollDelay adds a delay typical after scrolling
func RandomScrollDelay() {
	RandomDelay(500, 1500)
}

// RandomClickDelay adds a delay typical after clicking
func RandomClickDelay() {
	RandomDelay(300, 800)
}

// RandomPageLoadDelay adds a delay typical after page load
func RandomPageLoadDelay() {
	RandomDelay(1000, 3000)
}
