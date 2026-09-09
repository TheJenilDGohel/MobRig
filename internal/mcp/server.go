package mcp

import (
	"context"

	"github.com/TheJenilDGohel/MobRig/internal/database"
	"github.com/TheJenilDGohel/MobRig/internal/platform"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the official Model Context Protocol server.
type Server struct {
	mcpServer *mcp.Server
	registry  *platform.AdapterRegistry
	repos     *database.Repositories
}

// NewServer initializes the MobRig MCP server and registers all device automation tools.
// Every tool is registered under both its primary name and its mobile_ prefixed alias.
func NewServer(registry *platform.AdapterRegistry, repos *database.Repositories, version string) (*Server, error) {
	impl := &mcp.Implementation{
		Name:        "mobrig",
		Title:       "MobRig Mobile Engine",
		Description: "MobRig high-speed mobile device control and automation engine for AI agents",
		Version:     version,
	}

	mcpServer := mcp.NewServer(impl, nil)
	s := &Server{
		mcpServer: mcpServer,
		registry:  registry,
		repos:     repos,
	}

	s.registerTools()

	return s, nil
}

// UnderlyingServer returns the wrapped *mcp.Server handle.
func (s *Server) UnderlyingServer() *mcp.Server {
	return s.mcpServer
}

// ServeStdio starts serving MCP requests over standard I/O (stdin/stdout).
// This call blocks until ctx is cancelled or standard input closes.
func (s *Server) ServeStdio(ctx context.Context) error {
	transport := &mcp.StdioTransport{}
	return s.mcpServer.Run(ctx, transport)
}

func (s *Server) registerTools() {
	// Helper to register both canonical name and mobile_ prefixed alias
	registerDual := func(name, desc string, fn func(toolName string)) {
		fn(name)
		fn("mobile_" + name)
	}

	// 1. list_devices
	listDevicesH := handleListDevices(s.registry)
	registerDual("list_devices", "List all currently connected mobile devices and emulators", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "List all connected mobile devices"}, listDevicesH)
	})

	// 2. get_device_info
	devInfoH := handleGetDeviceInfo(s.registry)
	registerDual("get_device_info", "Get detailed specs (model, OS, screen dimensions) for a device", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Get hardware and OS info for a device"}, devInfoH)
	})

	// 3. get_screenshot
	screenshotH := handleGetScreenshot(s.registry)
	registerDual("get_screenshot", "Capture a PNG screenshot of the device display", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Capture a PNG screenshot of the device display"}, screenshotH)
	})

	// 4. get_ui_tree
	uiTreeH := handleGetUITree(s.registry)
	registerDual("get_ui_tree", "Read the compact UI hierarchy with element indices, types, and bounds", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Read the compact UI hierarchy with element indices, types, and bounds"}, uiTreeH)
	})

	// 5. find_element
	findElementH := handleFindElement(s.registry)
	registerDual("find_element", "Find a specific element on screen by text, description, or selector", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Find an element by text or selector"}, findElementH)
	})

	// 6. wait_for_element
	waitForElementH := handleWaitForElement(s.registry)
	registerDual("wait_for_element", "Poll the screen until an element with matching text or selector appears", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Wait for an element to appear on screen"}, waitForElementH)
	})

	// 7. tap
	tapH := handleTap(s.registry)
	registerDual("tap", "Tap an element by index (from get_ui_tree) or screen coordinates (x, y)", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Tap an element by index or (x, y) coordinates"}, tapH)
	})

	// 8. double_tap
	doubleTapH := handleDoubleTap(s.registry)
	registerDual("double_tap", "Double tap an element by index or coordinates", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Double tap an element by index or coordinates"}, doubleTapH)
	})

	// 9. long_press
	longPressH := handleLongPress(s.registry)
	registerDual("long_press", "Long press an element by index or coordinates with duration in ms", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Long press an element by index or coordinates"}, longPressH)
	})

	// 10. swipe
	swipeH := handleSwipe(s.registry)
	registerDual("swipe", "Swipe gesture: direction ('up' to scroll down, 'down' to scroll up) or coordinates", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Perform a swipe or scroll gesture"}, swipeH)
	})

	// 11. type_text
	typeTextH := handleTypeText(s.registry)
	registerDual("type_text", "Type text into the focused field (optionally tap field index first)", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Type text into focused input field"}, typeTextH)
	})

	// 12. press_button
	pressButtonH := handlePressButton(s.registry)
	registerDual("press_button", "Press a hardware/system button by keycode: 3=Home, 4=Back, 66=Enter", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Press a keycode (Back=4, Home=3, Enter=66)"}, pressButtonH)
	})

	// 13. go_home
	goHomeH := handleGoHome(s.registry)
	registerDual("go_home", "Navigate to the device home screen", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Navigate to the device home screen"}, goHomeH)
	})

	// 14. launch_app
	launchAppH := handleLaunchApp(s.registry)
	registerDual("launch_app", "Launch an installed app by package name (e.g. com.android.settings)", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Launch an installed app by package name"}, launchAppH)
	})

	// 15. terminate_app
	terminateAppH := handleTerminateApp(s.registry)
	registerDual("terminate_app", "Force-stop an application by package name", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Force-stop an application"}, terminateAppH)
	})

	// 16. install_app
	installAppH := handleInstallApp(s.registry)
	registerDual("install_app", "Install an application from a local APK path", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Install an APK from local path"}, installAppH)
	})

	// 17. uninstall_app
	uninstallAppH := handleUninstallApp(s.registry)
	registerDual("uninstall_app", "Uninstall an application by package name", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Uninstall an app by package name"}, uninstallAppH)
	})

	// 18. open_url
	openURLH := handleOpenURL(s.registry)
	registerDual("open_url", "Open a web URL or deep link on the device", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Open a URL or deep link on device"}, openURLH)
	})

	// 19. clear_app_data
	clearAppDataH := handleClearAppData(s.registry)
	registerDual("clear_app_data", "Reset all data and cache for an application", func(n string) {
		mcp.AddTool(s.mcpServer, &mcp.Tool{Name: n, Description: "Reset all app data and cache"}, clearAppDataH)
	})
}
