package android

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TheJenilDGohel/MobRig/internal/platform"
)

// rawXMLNode represents a <node> tag in uiautomator XML dump.
type rawXMLNode struct {
	XMLName     xml.Name     `xml:"node"`
	Index       int          `xml:"index,attr"`
	Text        string       `xml:"text,attr"`
	ResourceID  string       `xml:"resource-id,attr"`
	Class       string       `xml:"class,attr"`
	Package     string       `xml:"package,attr"`
	ContentDesc string       `xml:"content-desc,attr"`
	Checkable   bool         `xml:"checkable,attr"`
	Checked     bool         `xml:"checked,attr"`
	Clickable   bool         `xml:"clickable,attr"`
	Enabled     bool         `xml:"enabled,attr"`
	Focusable   bool         `xml:"focusable,attr"`
	Focused     bool         `xml:"focused,attr"`
	Scrollable  bool         `xml:"scrollable,attr"`
	Selected    bool         `xml:"selected,attr"`
	Bounds      string       `xml:"bounds,attr"`
	Children    []rawXMLNode `xml:"node"`
}

// rawXMLHierarchy represents the root <hierarchy> tag in uiautomator dump.
type rawXMLHierarchy struct {
	XMLName  xml.Name     `xml:"hierarchy"`
	Rotation int          `xml:"rotation,attr"`
	Nodes    []rawXMLNode `xml:"node"`
}

// ParseUIAttributes parses raw XML hierarchy into a compact UITree.
func ParseUITreeXML(xmlContent string, opts platform.UITreeOptions) (*platform.UITree, error) {
	var hier rawXMLHierarchy
	if err := xml.Unmarshal([]byte(xmlContent), &hier); err != nil {
		return nil, fmt.Errorf("failed to parse uiautomator xml: %w", err)
	}

	tree := &platform.UITree{
		Elements:  []platform.UIElement{},
		Format:    "compact",
		Timestamp: time.Now(),
	}

	var flatList []platform.UIElement
	appPackage := ""

	var walk func(node rawXMLNode, parentIdx int, depth int)
	walk = func(node rawXMLNode, parentIdx int, depth int) {
		if opts.MaxDepth > 0 && depth > opts.MaxDepth {
			return
		}

		if appPackage == "" && node.Package != "" {
			appPackage = node.Package
		}

		b := ParseBounds(node.Bounds)
		shortType := shortenClassName(node.Class)

		elem := platform.UIElement{
			Index:       len(flatList),
			Text:        node.Text,
			Type:        shortType,
			ResourceID:  node.ResourceID,
			ContentDesc: node.ContentDesc,
			Bounds:      b,
			Clickable:   node.Clickable,
			Scrollable:  node.Scrollable,
			Focused:     node.Focused,
			Enabled:     node.Enabled,
			Selected:    node.Selected,
			ParentIndex: parentIdx,
			Depth:       depth,
		}

		if node.Checkable {
			checkedVal := node.Checked
			elem.Checked = &checkedVal
		}

		currentIdx := elem.Index
		flatList = append(flatList, elem)

		// Record child index on parent
		if parentIdx >= 0 && parentIdx < len(flatList) {
			flatList[parentIdx].Children = append(flatList[parentIdx].Children, currentIdx)
		}

		for _, child := range node.Children {
			walk(child, currentIdx, depth+1)
		}
	}

	for _, rootNode := range hier.Nodes {
		walk(rootNode, -1, 0)
	}

	tree.AppPackage = appPackage

	// Filter elements according to opts
	filtered := filterElements(flatList, opts)
	tree.Elements = filtered
	tree.TokenCount = estimateTokenCount(filtered)

	return tree, nil
}

// ParseBounds converts bounds strings like "[0,0][1080,2400]" into a platform.Bounds struct.
func ParseBounds(boundsStr string) platform.Bounds {
	b := platform.Bounds{}
	if boundsStr == "" {
		return b
	}

	// Format: [x1,y1][x2,y2]
	boundsStr = strings.ReplaceAll(boundsStr, "][", ",")
	boundsStr = strings.Trim(boundsStr, "[]")
	parts := strings.Split(boundsStr, ",")

	if len(parts) == 4 {
		b.Left, _ = strconv.Atoi(parts[0])
		b.Top, _ = strconv.Atoi(parts[1])
		b.Right, _ = strconv.Atoi(parts[2])
		b.Bottom, _ = strconv.Atoi(parts[3])
	}

	return b
}

// shortenClassName converts verbose Android package classes into simple type names.
// e.g. "android.widget.TextView" -> "TextView"
func shortenClassName(fullClass string) string {
	if fullClass == "" {
		return "View"
	}
	parts := strings.Split(fullClass, ".")
	return parts[len(parts)-1]
}

// filterElements applies VisibleOnly and TextFilter options.
func filterElements(elements []platform.UIElement, opts platform.UITreeOptions) []platform.UIElement {
	if !opts.VisibleOnly && opts.TextFilter == "" {
		return elements
	}

	textFilterLower := strings.ToLower(opts.TextFilter)
	var result []platform.UIElement

	for _, e := range elements {
		// VisibleOnly check: element must have non-zero dimensions
		if opts.VisibleOnly {
			if e.Bounds.Width() <= 0 || e.Bounds.Height() <= 0 {
				continue
			}
			// Skip pure layout containers that carry zero user information and aren't interactive
			if isUninformativeContainer(e) {
				continue
			}
		}

		// TextFilter check
		if textFilterLower != "" {
			hasText := strings.Contains(strings.ToLower(e.Text), textFilterLower)
			hasDesc := strings.Contains(strings.ToLower(e.ContentDesc), textFilterLower)
			hasResID := strings.Contains(strings.ToLower(e.ResourceID), textFilterLower)
			if !hasText && !hasDesc && !hasResID {
				continue
			}
		}

		result = append(result, e)
	}

	// Re-index filtered elements sequentially so agents have clean indices
	for i := range result {
		result[i].Index = i
	}

	return result
}

// isUninformativeContainer checks if a node is an empty invisible/transparent layout container.
func isUninformativeContainer(e platform.UIElement) bool {
	switch e.Type {
	case "FrameLayout", "LinearLayout", "RelativeLayout", "ViewGroup", "CoordinatorLayout":
		// Only uninformative if it has no text, no desc, and is not clickable/focusable
		return e.Text == "" && e.ContentDesc == "" && !e.Clickable && !e.Scrollable && !e.Focused
	default:
		return false
	}
}

// estimateTokenCount provides a lightweight estimation of token consumption for LLMs.
// Rough rule: ~4 chars per token for typical compact structured JSON.
func estimateTokenCount(elements []platform.UIElement) int {
	totalChars := 0
	for _, e := range elements {
		totalChars += len(e.Text) + len(e.Type) + len(e.ResourceID) + len(e.ContentDesc) + 30 // overhead per node
	}
	tokens := totalChars / 4
	if tokens < 10 && len(elements) > 0 {
		return 10
	}
	return tokens
}
