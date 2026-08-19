package olx

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

const extractJS = `(() => {
  const hrefs = Array.from(document.querySelectorAll("a[href*='/d/']"))
    .map((a) => a.href)
    .filter((h) => h.includes("olx.ua"))
    .slice(0, 20);
  const title = (document.querySelector("h1") && document.querySelector("h1").textContent || "").trim();
  const priceText = document.body.innerText.match(/([0-9][0-9 ]{3,})\s*(грн|₴|USD|\$)/i);
  let price = 0;
  let currency = "UAH";
  if (priceText) {
    price = Number(priceText[1].replace(/\s/g, ""));
    currency = /usd|\$/i.test(priceText[2]) ? "USD" : "UAH";
  }
  const idMatch = location.href.match(/-ID([A-Za-z0-9]+)/) || location.href.match(/\/(\d+)(?:\.html)?$/);
  return {
    hrefs,
    title,
    price,
    currency,
    id: idMatch ? idMatch[1] : location.href,
  };
})()`

// Chrome drives a local Chrome/Chromium via DevTools (no Node/npm).
// Concurrency is 1: one tab at a time, human-like jitter, captcha = stop.
type Chrome struct {
	mu     sync.Mutex
	exec   string
	alloc  context.Context
	cancel context.CancelFunc
}

func NewChrome() *Chrome {
	return &Chrome{exec: FindChrome()}
}

func (c *Chrome) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		c.cancel()
		c.alloc = nil
		c.cancel = nil
	}
}

func (c *Chrome) ready(ctx context.Context) error {
	if c.alloc != nil {
		return nil
	}
	if c.exec == "" {
		c.exec = FindChrome()
	}
	if c.exec == "" {
		return fmt.Errorf("chrome/chromium not found (set CHROME_PATH)")
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(c.exec),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("lang", "uk-UA"),
		chromedp.WindowSize(1280, 800),
		chromedp.UserAgent(userAgent),
	)
	alloc, cancel := chromedp.NewExecAllocator(ctx, opts...)
	c.alloc = alloc
	c.cancel = cancel
	return nil
}

func (c *Chrome) Extract(ctx context.Context, pageURL string) (Extract, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ready(ctx); err != nil {
		return Extract{}, err
	}
	tab, cancel := chromedp.NewContext(c.alloc)
	defer cancel()
	tab, cancelTO := context.WithTimeout(tab, 50*time.Second)
	defer cancelTO()

	sleep(ctx, 400, 1200)

	var html string
	var data map[string]any
	err := chromedp.Run(tab,
		chromedp.Navigate(pageURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(context.Context) error {
			sleep(ctx, 600, 1800)
			return nil
		}),
		chromedp.Evaluate(`window.scrollBy(0, 400)`, nil),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
		chromedp.Evaluate(extractJS, &data),
	)
	if err != nil {
		return Extract{}, err
	}
	if htmlBlocked(html) {
		return Extract{Blocked: true}, CaptchaError{Msg: "olx captcha or waf"}
	}
	return Extract{HTML: html, JSON: data}, nil
}

func htmlBlocked(html string) bool {
	t := strings.ToLower(html)
	return strings.Contains(t, "captcha") || strings.Contains(t, "cf-challenge") || strings.Contains(t, "access denied")
}

func sleep(ctx context.Context, minMS, maxMS int) {
	d := time.Duration(minMS+rand.IntN(maxMS-minMS+1)) * time.Millisecond
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

// FindChrome locates a system Chrome/Chromium binary. No npm/playwright downloads.
func FindChrome() string {
	if p := os.Getenv("CHROME_PATH"); p != "" {
		return p
	}
	var candidates []string
	switch runtime.GOOS {
	case "darwin":
		candidates = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		}
	case "windows":
		candidates = []string{
			filepath.Join(os.Getenv("PROGRAMFILES"), "Google/Chrome/Application/chrome.exe"),
			filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Google/Chrome/Application/chrome.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Google/Chrome/Application/chrome.exe"),
		}
	default:
		candidates = []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "chrome"}
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if filepath.IsAbs(c) {
			if _, err := os.Stat(c); err == nil {
				return c
			}
			continue
		}
		if p, err := exec.LookPath(c); err == nil {
			return p
		}
	}
	return ""
}
