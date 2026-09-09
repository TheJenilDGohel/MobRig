package platform

import "time"

// DeviceStatus represents the connection state of a device.
type DeviceStatus string

const (
	StatusConnected    DeviceStatus = "connected"
	StatusDisconnected DeviceStatus = "disconnected"
	StatusOffline      DeviceStatus = "offline"
	StatusBusy         DeviceStatus = "busy"
	StatusUnauthorized DeviceStatus = "unauthorized"
)

// ConnectionType describes how a device is connected.
type ConnectionType string

const (
	ConnectionUSB       ConnectionType = "usb"
	ConnectionWiFi      ConnectionType = "wifi"
	ConnectionEmulator  ConnectionType = "emulator"
	ConnectionSimulator ConnectionType = "simulator"
)

// DevicePlatform identifies the device OS.
type DevicePlatform string

const (
	PlatformAndroid DevicePlatform = "android"
	PlatformIOS     DevicePlatform = "ios"
)

// DeviceInfo holds metadata about a connected device.
type DeviceInfo struct {
	ID             string         `json:"id"`
	Platform       DevicePlatform `json:"platform"`
	Model          string         `json:"model"`
	Manufacturer   string         `json:"manufacturer"`
	OSVersion      string         `json:"os_version"`
	APILevel       int            `json:"api_level,omitempty"` // Android only
	ScreenWidth    int            `json:"screen_width"`
	ScreenHeight   int            `json:"screen_height"`
	ScreenDensity  float64        `json:"screen_density,omitempty"`
	RAMMB          int            `json:"ram_mb,omitempty"`
	Status         DeviceStatus   `json:"status"`
	ConnectionType ConnectionType `json:"connection_type"`
	TransportID    string         `json:"transport_id,omitempty"` // ADB transport ID
}

// Screenshot holds captured screen data.
type Screenshot struct {
	Data      []byte    `json:"-"`        // raw PNG bytes (not serialized to JSON)
	Base64    string    `json:"base64"`   // base64-encoded PNG for JSON transport
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Format    string    `json:"format"`   // "png"
	Timestamp time.Time `json:"timestamp"`
	SizeBytes int       `json:"size_bytes"`
}

// UITreeOptions controls how the UI tree is read.
type UITreeOptions struct {
	VisibleOnly      bool   `json:"visible_only"`       // only return visible elements
	TextFilter       string `json:"text_filter"`        // filter elements by text content
	DiffFromPrevious bool   `json:"diff_from_previous"` // return only changed elements
	MaxDepth         int    `json:"max_depth"`          // limit tree depth (0 = unlimited)
}

// UITree represents the device's UI hierarchy in compact format.
type UITree struct {
	Elements   []UIElement `json:"elements"`
	TokenCount int         `json:"token_count"` // estimated LLM token count
	Format     string      `json:"format"`      // "compact" or "raw"
	ScreenName string      `json:"screen_name,omitempty"`
	AppPackage string      `json:"app_package,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
}

// UIElement represents a single UI element in the tree.
type UIElement struct {
	Index       int       `json:"index"`
	Text        string    `json:"text,omitempty"`
	Type        string    `json:"type"`                   // e.g., "Button", "TextView", "EditText"
	ResourceID  string    `json:"resource_id,omitempty"`
	ContentDesc string    `json:"content_desc,omitempty"`
	Bounds      Bounds    `json:"bounds"`
	Clickable   bool      `json:"clickable,omitempty"`
	Scrollable  bool      `json:"scrollable,omitempty"`
	Focused     bool      `json:"focused,omitempty"`
	Enabled     bool      `json:"enabled,omitempty"`
	Selected    bool      `json:"selected,omitempty"`
	Checked     *bool     `json:"checked,omitempty"` // pointer to distinguish false from absent
	Children    []int     `json:"children,omitempty"` // indices of child elements
	ParentIndex int       `json:"parent_index,omitempty"`
	Depth       int       `json:"depth,omitempty"`
}

// Bounds represents the screen rectangle of a UI element.
type Bounds struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
}

// CenterX returns the horizontal center of the bounds.
func (b Bounds) CenterX() int {
	return (b.Left + b.Right) / 2
}

// CenterY returns the vertical center of the bounds.
func (b Bounds) CenterY() int {
	return (b.Top + b.Bottom) / 2
}

// Width returns the width of the bounds.
func (b Bounds) Width() int {
	return b.Right - b.Left
}

// Height returns the height of the bounds.
func (b Bounds) Height() int {
	return b.Bottom - b.Top
}

// ActionResult is the outcome of a device interaction.
// Message is always an NLP-friendly summary (per R010).
type ActionResult struct {
	Success  bool          `json:"success"`
	Message  string        `json:"message"`  // NLP-friendly: "Tapped the Login button at (152, 23)"
	Duration time.Duration `json:"duration"` // how long the action took
	Error    string        `json:"error,omitempty"`
}
