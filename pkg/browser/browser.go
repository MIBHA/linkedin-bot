package browser

import (
	"fmt"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/mithu/linkedin-bot/pkg/stealth"
)

type Config struct {
	Headless      bool
	DisableImages bool
	UserDataDir   string
	ProxyServer   string
	WindowWidth   int
	WindowHeight  int
	Timeout       time.Duration
	SlowMotion    time.Duration
}

func DefaultConfig() *Config {
	return &Config{Headless: false, DisableImages: false, WindowWidth: 1366, WindowHeight: 768, Timeout: 30 * time.Second, SlowMotion: 0}
}

type Manager struct {
	browser     *rod.Browser
	launcher    *launcher.Launcher
	config      *Config
	fingerprint *stealth.FingerprintConfig
}

func NewManager(config *Config) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}
	return &Manager{config: config, fingerprint: stealth.NewRandomFingerprintConfig()}, nil
}

func (m *Manager) Launch() error {
	// Disable Leakless to avoid AV false positives on Windows
	l := launcher.New().
		Leakless(false).
		Headless(m.config.Headless).
		NoSandbox(true).
		Devtools(false)
	if m.config.UserDataDir != "" {
		l = l.UserDataDir(m.config.UserDataDir)
	}
	if m.config.ProxyServer != "" {
		l = l.Proxy(m.config.ProxyServer)
	}
	l = l.Set("disable-blink-features", "AutomationControlled").Set("excludeSwitches", "enable-automation").Set("disable-infobars").Set("start-maximized")
	url, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}
	m.launcher = l
	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("failed to connect to browser: %w", err)
	}
	if m.config.SlowMotion > 0 {
		browser = browser.SlowMotion(m.config.SlowMotion)
	}
	browser = browser.Timeout(m.config.Timeout)
	m.browser = browser
	return nil
}

func (m *Manager) NewPage(url string) (*rod.Page, error) {
	if m.browser == nil {
		return nil, fmt.Errorf("browser not launched")
	}
	page, err := m.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}
	if err := stealth.ApplyFingerprint(page, m.fingerprint); err != nil {
		return nil, fmt.Errorf("failed to apply fingerprint: %w", err)
	}
	if m.config.DisableImages {
		router := page.HijackRequests()
		router.MustAdd("*", func(ctx *rod.Hijack) {
			if ctx.Request.Type() == proto.NetworkResourceTypeImage {
				ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
				return
			}
			ctx.ContinueRequest(&proto.FetchContinueRequest{})
		})
		go router.Run()
	}
	if url != "" {
		if err := page.Navigate(url); err != nil {
			return nil, fmt.Errorf("failed to navigate to %s: %w", url, err)
		}
		if err := page.WaitLoad(); err != nil {
			return nil, fmt.Errorf("failed to wait for page load: %w", err)
		}
		stealth.RandomPageLoadDelay()
	}
	return page, nil
}

func (m *Manager) GetBrowser() *rod.Browser {
	return m.browser
}

func (m *Manager) Close() error {
	if m.browser != nil {
		if err := m.browser.Close(); err != nil {
			return fmt.Errorf("failed to close browser: %w", err)
		}
	}
	if m.launcher != nil {
		m.launcher.Cleanup()
	}
	return nil
}

func WaitForUserAction(message string, timeoutSeconds int) error {
	fmt.Printf("\n %s\n", message)
	fmt.Printf("Press Enter to continue (timeout: %d seconds)...\n", timeoutSeconds)
	done := make(chan bool)
	go func() {
		var input string
		fmt.Scanln(&input)
		done <- true
	}()
	select {
	case <-done:
		fmt.Println("Continuing...")
		return nil
	case <-time.After(time.Duration(timeoutSeconds) * time.Second):
		fmt.Println("Timeout reached, continuing...")
		return nil
	}
}

func Screenshot(page *rod.Page, filename string) error {
	data, err := page.Screenshot(true, &proto.PageCaptureScreenshot{Format: proto.PageCaptureScreenshotFormatPng})
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to save screenshot: %w", err)
	}
	fmt.Printf("screenshot saved: %s\n", filename)
	return nil
}
