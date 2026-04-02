package config

import (
	"context"
	"log"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/TOsmanov/go-pdf/internal/lib/utils"
	"github.com/chromedp/chromedp"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env               string        `yaml:"environment" env-default:"prod"`
	Limit             int           `yaml:"limit" env-default:"1"`
	ChromeTimeout     time.Duration `yaml:"chrome-timeout" env-default:"60s"`
	UrlsLimit         int           `yaml:"urls-limit" env-default:"30"`
	TrustedHosts      []string      `yaml:"trusted-hosts"`
	HTTPSOnly         bool          `yaml:"https-only"`
	Debug             bool
	Face              `yaml:"face"`
	Core              `yaml:"core"`
	HTTPServer        `yaml:"http-server"`
	FaceTemplate      *template.Template // TODO: move to another structure
	TitlePageTemplate *template.Template // TODO: move to another structure
	ToCPageTemplate   *template.Template // TODO: move to another structure
	Reload            chan error         // TODO: move to another structure
	// SizeLimit         float64       `yaml:"size-limit" env-default:"32"`
	Browser BrowserConfig
}

type Face struct {
	FaceEnable bool   `yaml:"enable" env-default:"false"`
	FacePage   string `yaml:"path" env-default:"face/index.html"`
	Prefix     string `yaml:"prefix"`
}

type Core struct {
	Docx `yaml:"docx"`
	Pdf  `yaml:"pdf"`
}

type HTTPServer struct {
	Address         string        `yaml:"address" env-default:"localhost:8080"`
	Timeout         time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout     time.Duration `yaml:"idleTimeout" env-default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdownTimeout" env-default:"10s"`
}

type Pdf struct {
	PdfSelector      string `yaml:"selector" env-default:"body"`
	PdfTitleSelector string `yaml:"title-selector" env-default:"title"`
	CacheDir         string `yaml:"cache-dir" env-default:"cache"`
	CustomScripts    `yaml:"custom-scripts"`
	ServicePages     `yaml:"service-pages"`
}

type CustomScripts struct {
	GrabScript  string `yaml:"grab-script" env-default:""`
	MergeScript string `yaml:"merge-script" env-default:""`
}

type Docx struct {
	DocxSelector string `yaml:"selector-class"`
}

type ServicePages struct {
	LogoPath  string `yaml:"logo-path"`
	Templates `yaml:"templates"`
}

type Templates struct {
	TocPage   string `yaml:"toc-page" env-default:"templates/toc-page.html"`
	TitlePage string `yaml:"title-page" env-default:"templates/title-page.html"`
}

type BrowserConfig struct {
	Ctx           *context.Context
	isRunning     bool
	allocCancel   context.CancelFunc
	browserCancel context.CancelFunc
}

func MustLoad() *Config {
	configPath := utils.GetEnv("CONFIG_PATH", "./config/local.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist: %s", configPath)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Cannot read config: %s", err)
	}
	switch cfg.Env {
	case "local":
		cfg.Debug = true
		cfg.HTTPSOnly = false
	case "docker-debug":
		cfg.Debug = true
		cfg.HTTPSOnly = false
	case "prod":
		cfg.Debug = false
		cfg.HTTPSOnly = true
	default:
		cfg.Debug = false
		cfg.HTTPSOnly = true
	}

	cfg.Limit = 1

	cfg.CreateBrowser()

	return &cfg
}

func (cfg *Config) CreateBrowser() error {
	if cfg.Browser.isRunning {
		cfg.CloseBrowser()
		cfg.Browser.isRunning = false
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir("/tmp/chrome-cache"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("single-process", true),
		chromedp.Flag("no-zygote", true),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("max-old-space-size", "128"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)

	cfg.Browser.Ctx = &browserCtx
	cfg.Browser.allocCancel = allocCancel
	cfg.Browser.browserCancel = browserCancel
	cfg.Browser.isRunning = true

	log.Println("The browser is running successfully")
	return nil
}

func (cfg *Config) CloseBrowser() {
	if !cfg.Browser.isRunning {
		return
	}

	log.Println("Close browser...")

	if cfg.Browser.browserCancel != nil {
		cfg.Browser.browserCancel()
	}

	if cfg.Browser.allocCancel != nil {
		cfg.Browser.allocCancel()
	}

	time.Sleep(2 * time.Second)

	cfg.forceCleanupProcesses()

	cfg.Browser.isRunning = false
	cfg.Browser.Ctx = nil
	cfg.Browser.allocCancel = nil
	cfg.Browser.browserCancel = nil

	log.Println("Browser is closed")
}

func (cfg *Config) forceCleanupProcesses() {
	// graceful shutdown
	exec.Command("pkill", "-15", "-f", "chrome").Run() // SIGTERM
	exec.Command("pkill", "-15", "-f", "chromium").Run()
	exec.Command("pkill", "-15", "-f", "headless-shell").Run()

	time.Sleep(2 * time.Second)

	// force shutdown
	exec.Command("pkill", "-9", "-f", "chrome").Run()
	exec.Command("pkill", "-9", "-f", "chromium").Run()
	exec.Command("pkill", "-9", "-f", "headless-shell").Run()

	exec.Command("rm", "-rf", "/tmp/.org.chromium.Chromium*").Run()
	exec.Command("rm", "-rf", "/tmp/chromedp-*").Run()

	exec.Command("rm", "-rf", "/tmp/chrome-cache").Run()
}

func (cfg *Config) InitTemplates() error {
	var err error
	cfg.FaceTemplate, err = template.ParseFiles(cfg.Face.FacePage)
	if err != nil {
		return err
	}
	cfg.TitlePageTemplate, err = template.ParseFiles(cfg.Templates.TitlePage)
	if err != nil {
		return err
	}
	cfg.ToCPageTemplate, err = template.ParseFiles(cfg.Templates.TocPage)
	if err != nil {
		return err
	}
	return nil
}
