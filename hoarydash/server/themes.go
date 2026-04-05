package main

import (
	"html/template"
	"os"

	"dario.cat/mergo"
	"go.yaml.in/yaml/v4"
)

type Theme struct {
	Background         template.CSS
	OpaqueBackground   template.CSS `yaml:"opaque_background"`
	Cards              CardStyle    // Default for widgets, entities and sensors
	Entities           CardStyle
	Sensors            CardStyle
	Widgets            CardStyle
	ButtonBackground   template.CSS `yaml:"button_background"`
	FontColor          template.CSS `yaml:"font_color"`
	SecondaryFontColor template.CSS `yaml:"secondary_font_color"`
	IconColor          template.CSS `yaml:"icon_color"`
	DisabledIconColor  template.CSS `yaml:"disabled_icon_color"`
	FontSize           int          `yaml:"font_size"`
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

var defaultTheme = Theme{
	Background:         "#000000",
	OpaqueBackground:   "#0f0f0f",
	FontColor:          "#ffffff",
	SecondaryFontColor: "#ffffffa2",
	IconColor:          "hsl(0, 0%, 90%)",
	DisabledIconColor:  "hsla(0, 0%, 90%, )",
	FontSize:           18,
	Cards: CardStyle{
		Borders:      newTrue(),
		BorderColor:  "rgba(130,185,255,0.15)",
		BorderRadius: "0.5em",
		Background:   "rgba(18,18,18,0.75)",
	},
}

type ThemesMap map[string]Theme

func mergeTheme(base, override Theme) Theme {
	result := base
	mergo.Merge(&result, override, mergo.WithOverride)
	return result
}

func resolveTheme(named ThemesMap, layers ...ThemeRef) Theme {
	result := defaultTheme
	for _, ref := range layers {
		if ref.Name != "" {
			if namedTheme, ok := named[ref.Name]; ok {
				result = mergeTheme(result, namedTheme)
			}
		}
		result = mergeTheme(result, ref.Theme)
	}
	return result
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
