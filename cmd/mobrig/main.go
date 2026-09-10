package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	mobrig "github.com/TheJenilDGohel/MobRig"
	"github.com/TheJenilDGohel/MobRig/internal/config"
	"github.com/TheJenilDGohel/MobRig/internal/database"
	mobrigmcp "github.com/TheJenilDGohel/MobRig/internal/mcp"
	"github.com/TheJenilDGohel/MobRig/internal/platform"
	"github.com/TheJenilDGohel/MobRig/internal/platform/android"
	"github.com/TheJenilDGohel/MobRig/internal/server"
	"github.com/TheJenilDGohel/MobRig/internal/setup"
)

func main() {
	// Handle version command
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("MobRig v%s\n", config.Version)
		return
	}

	// Handle setup command
	if len(os.Args) > 1 && (os.Args[1] == "setup" || os.Args[1] == "--setup") {
		runSetup()
		return
	}

	// Handle MCP server subcommand: stdio mode for AI agents
	if len(os.Args) > 1 && (os.Args[1] == "mcp" || os.Args[1] == "--mcp") {
		// Crucial for stdio transport: all logs MUST go to stderr so stdout is purely JSON-RPC
		log.SetOutput(os.Stderr)
		runMCP()
		return
	}

	// ── 1. Parse CLI flags ──────────────────────────────────────────
	var (
		flagPort    = flag.Int("port", 0, "HTTP API port (default: 8686)")
		flagDataDir = flag.String("data-dir", "", "Data directory path (default: ~/.mobrig)")
		flagDebug   = flag.Bool("debug", false, "Enable debug logging")
	)
	flag.Parse()

	// ── 2. Load config (defaults → file → env → CLI overrides) ─────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Apply CLI overrides on top of loaded config
	if *flagPort > 0 {
		cfg.HTTPPort = *flagPort
	}
	if *flagDataDir != "" {
		cfg.DataDir = *flagDataDir
		cfg.DBPath = filepath.Join(cfg.DataDir, config.DBFileName)
	}
	if *flagDebug {
		cfg.Debug = true
	}

	log.Printf("[mobrig] v%s starting...", cfg.Version)
	if cfg.Debug {
		log.Printf("[mobrig] config: data_dir=%s http=%s:%d", cfg.DataDir, cfg.HTTPHost, cfg.HTTPPort)
	}

	// ── 3. Ensure data directory exists ─────────────────────────────
	if err := cfg.EnsureDataDir(); err != nil {
		log.Fatalf("failed to create data directory %s: %v", cfg.DataDir, err)
	}

	// ── 4. Open database + run migrations ───────────────────────────
	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()
	log.Printf("[mobrig] database ready at %s", cfg.DBPath)

	repos := database.NewRepositories(db)

	// ── 5. Initialize platform adapter registry ─────────────────────
	registry := initAdapterRegistry(cfg, repos)

	// ── 6. Create the Wails-bound service ───────────────────────────
	svc := NewMobRigService(cfg, db, repos, registry)

	// ── 7. Start HTTP server in background ──────────────────────────
	httpServer := server.New(cfg, registry, repos)
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Printf("[mobrig] http server error: %v", err)
		}
	}()

	// ── 8. Set up graceful shutdown ─────────────────────────────────
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdownCh
		log.Println("[mobrig] shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("[mobrig] http shutdown error: %v", err)
		}
		if err := db.Close(); err != nil {
			log.Printf("[mobrig] db close error: %v", err)
		}

		log.Println("[mobrig] goodbye")
		os.Exit(0)
	}()

	// ── 9. Start Wails v3 desktop app ───────────────────────────────
	app := application.New(application.Options{
		Name:        "MobRig",
		Description: fmt.Sprintf("MobRig v%s — Mobile Device Automation for AI Agents", cfg.Version),
		Services: []application.Service{
			application.NewService(svc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(mobrig.Assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "MobRig",
		Width:  1200,
		Height: 800,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(5, 5, 5), // #050505
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatalf("[mobrig] wails app error: %v", err)
	}
}

func initAdapterRegistry(cfg *config.Config, repos *database.Repositories) *platform.AdapterRegistry {
	registry := platform.NewAdapterRegistry()

	adbPath, err := android.DetectADB(cfg.ADBPath)
	if err != nil {
		log.Printf("[mobrig] adb not found: %v (android devices will not be available)", err)
	} else {
		log.Printf("[mobrig] adb detected at: %s", adbPath)
		adbClient := android.NewADBClient(adbPath)

		var cli *android.AndroidCLI
		if cliPath, ok := android.DetectAndroidCLI(cfg.AndroidCLIPath); ok {
			log.Printf("[mobrig] google android CLI detected at: %s", cliPath)
			cli = android.NewAndroidCLI(cliPath)
		} else {
			log.Printf("[mobrig] google android CLI not found; using ADB fallback mode")
		}

		androidAdapter := android.NewAdapter(adbClient, cli, repos)
		registry.Register(androidAdapter)
		log.Printf("[mobrig] registered android adapter (dual-path active: cli=%v)", cli != nil)
	}

	return registry
}

func runMCP() {
	mcpFlags := flag.NewFlagSet("mcp", flag.ExitOnError)
	mcpFlags.SetOutput(os.Stderr)
	flagDataDir := mcpFlags.String("data-dir", "", "Data directory path (default: ~/.mobrig)")
	flagDebug := mcpFlags.Bool("debug", false, "Enable debug logging")
	if len(os.Args) > 2 {
		_ = mcpFlags.Parse(os.Args[2:])
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[mobrig-mcp] failed to load config: %v", err)
	}
	if *flagDataDir != "" {
		cfg.DataDir = *flagDataDir
		cfg.DBPath = filepath.Join(cfg.DataDir, config.DBFileName)
	}
	if *flagDebug {
		cfg.Debug = true
	}

	if err := cfg.EnsureDataDir(); err != nil {
		log.Fatalf("[mobrig-mcp] failed to create data directory %s: %v", cfg.DataDir, err)
	}

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("[mobrig-mcp] failed to open database: %v", err)
	}
	defer db.Close()

	repos := database.NewRepositories(db)
	registry := initAdapterRegistry(cfg, repos)

	srv, err := mobrigmcp.NewServer(registry, repos, cfg.Version)
	if err != nil {
		log.Fatalf("[mobrig-mcp] failed to create MCP server: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Printf("[mobrig-mcp] MCP server v%s running on stdio...", cfg.Version)
	if err := srv.ServeStdio(ctx); err != nil {
		log.Fatalf("[mobrig-mcp] MCP server terminated: %v", err)
	}
}

func runSetup() {
	cfg, _ := config.Load()
	report := setup.CheckEnvironment(cfg.ADBPath, cfg.AndroidCLIPath)

	fmt.Println("==================================================")
	fmt.Printf(" MobRig v%s — Environment Setup & Diagnostics\n", config.Version)
	fmt.Println("==================================================")

	if report.ADBFound {
		fmt.Printf(" [✓] Android Debug Bridge (ADB): %s\n", report.ADBPath)
		if report.ADBVersion != "" {
			fmt.Printf("     Version: %s\n", report.ADBVersion)
		}
	} else {
		fmt.Println(" [✗] Android Debug Bridge (ADB): NOT FOUND")
		fmt.Println("     Please install the Android SDK platform-tools or set ADB_PATH.")
	}

	if report.AndroidCLIFound {
		fmt.Printf(" [✓] Google Android CLI: %s\n", report.AndroidCLIPath)
	} else {
		fmt.Println(" [!] Google Android CLI: not installed (using fast ADB fallback)")
		fmt.Printf("     To install official CLI: %s\n", report.CLIInstallCmd)
	}

	fmt.Println("\n--- Configuring AI Agent Integration ---")
	execPath, err := os.Executable()
	if err != nil {
		execPath = "mobrig"
	}

	configured, errs := setup.GenerateAllConfigs(".", execPath)
	for _, p := range configured {
		fmt.Printf(" [✓] Generated/updated MCP config: %s\n", p)
	}
	for _, err := range errs {
		fmt.Printf(" [!] Notice: %v\n", err)
	}

	fmt.Println("\nSetup complete! You can now start MobRig or invoke `mobrig mcp` via Claude/Cursor.")
}
