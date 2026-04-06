package main

import (
	"fmt"
	"html/template"
	"os"

	"dario.cat/mergo"
	"go.yaml.in/yaml/v4"
)

type Theme struct {
	// Surfaces
	Background       template.CSS `yaml:"background"`
	Surface          template.CSS `yaml:"surface"`
	SurfaceOpaque    template.CSS `yaml:"surface_opaque"`
	SurfaceProminent template.CSS `yaml:"surface_prominent"`
	SurfaceAlt       template.CSS `yaml:"surface_alt"`
	Highlight        template.CSS `yaml:"highlight"`

	// Text + icons
	OnSurface       template.CSS `yaml:"on_surface"`
	OnSurfaceMuted  template.CSS `yaml:"on_surface_muted"`
	OnSurfaceSubtle template.CSS `yaml:"on_surface_subtle"`
	OnBackground    template.CSS `yaml:"on_background"`

	// Accent
	Accent      template.CSS `yaml:"accent"`
	AccentMuted template.CSS `yaml:"accent_muted"`

	// Interactive elements (buttons, sliders)
	Interactive         template.CSS `yaml:"interactive"`
	InteractiveMuted    template.CSS `yaml:"interactive_muted"`
	InteractiveDisabled template.CSS `yaml:"interactive_disabled"`

	// State
	StateOn       template.CSS `yaml:"state_on"`
	StateOff      template.CSS `yaml:"state_off"`
	StateDisabled template.CSS `yaml:"state_disabled"`

	// Semantic states
	Positive template.CSS `yaml:"positive"`
	Negative template.CSS `yaml:"negative"`

	// Font size
	FontSize int `yaml:"font_size"`

	// Structural styles
	Cards    *CardStyle
	Entities *CardStyle
	Sensors  *CardStyle
	Widgets  *CardStyle
	Modals   *CardStyle
}

type CardStyle struct {
	Borders      *bool
	BorderColor  template.CSS `yaml:"border_color"`
	BorderRadius template.CSS `yaml:"border_radius"`
	Background   template.CSS
	FontSize     int `yaml:"font_size"`
}

func newTrue() *bool {
	b := true
	return &b
}

func newDefaultCard() *CardStyle {
	style := CardStyle{
		Borders:      newTrue(),
		BorderColor:  "rgba(130,185,255,0.15)",
		BorderRadius: "0.5em",
		Background:   "rgba(18,18,18,0.75)",
	}
	return &style
}

func newDefaultModal() *CardStyle {
	style := newDefaultCard()
	style.Background = "rgb1(20,20,20)"
	return style
}

var defaultTheme = Theme{
	Background:       "#0f0f0f",
	Surface:          "rgba(18,18,18,0.75)",
	SurfaceOpaque:    "#1a1a1a",
	SurfaceProminent: "rgba(57, 57, 57, 0.85)",
	SurfaceAlt:       "rgba(66, 61, 42, 0.85)",
	Highlight:        "rgba(255, 255, 255, 0.12)",
	OnSurface:        "#ffffff",
	OnSurfaceMuted:   "#dbdbdb",
	OnSurfaceSubtle:  "#989898",
	Accent:           "hsl(210, 90%, 65%)",
	Interactive:      "hsl(210, 90%, 65%)",
	StateOn:          "hsl(134, 60%, 45%)",
	StateOff:         "rgba(200,220,240,0.38)",
	StateDisabled:    "rgba(220, 240, 250, 0.25)",
	Positive:         "hsl(140, 60%, 55%)",
	Negative:         "hsl(0, 70%, 55%)",
	FontSize:         18,
	Cards:            newDefaultCard(),
	Modals:           newDefaultModal(),
}

type ThemesMap map[string]Theme

func mergeTheme(base, override Theme) Theme {
	result := base
	mergo.Merge(&result, override, mergo.WithOverride)
	return result
}

func resolveTheme(named ThemesMap, ref ThemeRef) (*Theme, error) {
	if ref.Name != "" {
		if namedTheme, ok := named[ref.Name]; ok {
			return &namedTheme, nil
		}
		return nil, fmt.Errorf("Theme %q not found", ref.Name)
	}
	return ref.Theme, nil
}

func parseThemes() (*ThemesMap, error) {
	yaml_file, err := os.ReadFile(yamlPath + "/themes.yaml")
	parsed := ThemesMap{}
	if err != nil {
		return &parsed, err
	}
	err = yaml.Unmarshal(yaml_file, &parsed)
	for _, theme := range parsed {
		mergeTheme(defaultTheme, theme)
	}
	return &parsed, err
}
