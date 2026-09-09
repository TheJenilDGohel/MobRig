package android

import (
	"testing"

	"github.com/TheJenilDGohel/MobRig/internal/platform"
)

const sampleXML = `<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy rotation="0">
  <node index="0" text="" resource-id="" class="android.widget.FrameLayout" package="com.android.settings" content-desc="" checkable="false" checked="false" clickable="false" enabled="true" focusable="false" focused="false" scrollable="false" long-clickable="false" password="false" selected="false" bounds="[0,0][1080,2400]">
    <node index="0" text="Settings" resource-id="com.android.settings:id/title" class="android.widget.TextView" package="com.android.settings" content-desc="" checkable="false" checked="false" clickable="false" enabled="true" focusable="false" focused="false" scrollable="false" long-clickable="false" password="false" selected="false" bounds="[48,150][450,220]" />
    <node index="1" text="Search settings" resource-id="com.android.settings:id/search_action_bar" class="android.widget.EditText" package="com.android.settings" content-desc="Search Box" checkable="false" checked="false" clickable="true" enabled="true" focusable="true" focused="false" scrollable="false" long-clickable="true" password="false" selected="false" bounds="[48,260][1032,380]" />
    <node index="2" text="Offscreen item" resource-id="" class="android.widget.TextView" package="com.android.settings" content-desc="" checkable="false" checked="false" clickable="false" enabled="true" focusable="false" focused="false" scrollable="false" bounds="[0,0][0,0]" />
  </node>
</hierarchy>`

func TestParseUITreeXML_Basic(t *testing.T) {
	tree, err := ParseUITreeXML(sampleXML, platform.UITreeOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tree.AppPackage != "com.android.settings" {
		t.Errorf("expected package 'com.android.settings', got %q", tree.AppPackage)
	}

	if len(tree.Elements) != 4 {
		t.Fatalf("expected 4 raw elements, got %d", len(tree.Elements))
	}

	// First node: FrameLayout
	e0 := tree.Elements[0]
	if e0.Type != "FrameLayout" {
		t.Errorf("expected type FrameLayout, got %q", e0.Type)
	}
	if len(e0.Children) != 3 {
		t.Errorf("expected 3 children on root, got %d", len(e0.Children))
	}

	// Node 1: TextView (Settings)
	e1 := tree.Elements[1]
	if e1.Text != "Settings" || e1.Type != "TextView" {
		t.Errorf("unexpected e1: %+v", e1)
	}
	if e1.Bounds.Left != 48 || e1.Bounds.Top != 150 || e1.Bounds.Right != 450 || e1.Bounds.Bottom != 220 {
		t.Errorf("unexpected bounds: %+v", e1.Bounds)
	}

	// Node 2: EditText (Search settings)
	e2 := tree.Elements[2]
	if e2.Type != "EditText" || !e2.Clickable || e2.ContentDesc != "Search Box" {
		t.Errorf("unexpected e2: %+v", e2)
	}
}

func TestParseUITreeXML_VisibleOnly(t *testing.T) {
	tree, err := ParseUITreeXML(sampleXML, platform.UITreeOptions{VisibleOnly: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The zero-dimension node [0,0][0,0] and empty FrameLayout should be filtered out!
	for _, e := range tree.Elements {
		if e.Bounds.Width() <= 0 || e.Bounds.Height() <= 0 {
			t.Errorf("found zero-dimension element in visible_only: %+v", e)
		}
		if isUninformativeContainer(e) {
			t.Errorf("found uninformative container in visible_only: %+v", e)
		}
	}

	// Should contain TextView (Settings) and EditText (Search settings)
	if len(tree.Elements) != 2 {
		t.Fatalf("expected 2 visible informative elements, got %d", len(tree.Elements))
	}
}

func TestParseUITreeXML_TextFilter(t *testing.T) {
	tree, err := ParseUITreeXML(sampleXML, platform.UITreeOptions{TextFilter: "search"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tree.Elements) != 1 {
		t.Fatalf("expected 1 element matching 'search', got %d", len(tree.Elements))
	}
	if tree.Elements[0].Text != "Search settings" {
		t.Errorf("expected 'Search settings', got %q", tree.Elements[0].Text)
	}
}

func TestParseBounds(t *testing.T) {
	b := ParseBounds("[48,150][450,220]")
	if b.Left != 48 || b.Top != 150 || b.Right != 450 || b.Bottom != 220 {
		t.Errorf("unexpected bounds: %+v", b)
	}
	if b.Width() != 402 || b.Height() != 70 {
		t.Errorf("unexpected dimensions: %dx%d", b.Width(), b.Height())
	}
	if b.CenterX() != 249 || b.CenterY() != 185 {
		t.Errorf("unexpected center: (%d, %d)", b.CenterX(), b.CenterY())
	}
}

func TestShortenClassName(t *testing.T) {
	cases := map[string]string{
		"android.widget.TextView":             "TextView",
		"android.widget.Button":               "Button",
		"com.google.android.material.Button":  "Button",
		"androidx.recyclerview.widget.RecyclerView": "RecyclerView",
		"":                                    "View",
	}

	for full, expected := range cases {
		got := shortenClassName(full)
		if got != expected {
			t.Errorf("shortenClassName(%q) = %q, expected %q", full, got, expected)
		}
	}
}
