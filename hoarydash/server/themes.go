package main

import (
	"fmt"
	"html/template"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"dario.cat/mergo"
	"github.com/mazznoer/csscolorparser"
	"go.yaml.in/yaml/v4"
)

type Theme struct {
	IsLight    bool
	Vars       map[string]template.CSS
	Background template.CSS `yaml:"background"`
	Colors
	Derived derivedColors `yaml:"-"`

	Shapes
	// Font size
	FontSize int `yaml:"font_size"`

	// Structural styles
	Cards        *CardStyle
	Entities     *CardStyle
	Sensors      *CardStyle
	Widgets      *CardStyle
	Modals       *CardStyle
	Badges       *CardStyle
	BadgeButtons *CardStyle `yaml:"badge_buttons"`
	Buttons      *CardStyle
	Tooltips     *CardStyle
}

type Colors struct {
	// Surfaces
	Surface          template.CSS `yaml:"surface"`
	SurfaceOpaque    template.CSS `yaml:"surface_opaque"`
	SurfaceProminent template.CSS `yaml:"surface_prominent"`
	SurfaceSubtle    template.CSS `yaml:"surface_subtle"`
	SurfaceAlt       template.CSS `yaml:"surface_alt"`
	Highlight        template.CSS `yaml:"highlight"`
	Border           template.CSS `yaml:"border"`

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
}

// These colors are purely derived and can not be user themed
type derivedColors struct {
	PositiveMuted      template.CSS
	NegativeMuted      template.CSS
	StateOnMuted       template.CSS
	StateOffMuted      template.CSS
	StateDisabledMuted template.CSS
}
type Shapes struct {
	Borders            *bool
	TightBorderRadius  template.CSS `yaml:"tight_border_radius"`
	WideBorderRadius   template.CSS `yaml:"wide_border_radius"`
	MediumBorderRadius template.CSS `yaml:"medium_border_radius"`
	BorderThick        template.CSS `yaml:"border_thick"`
	BorderThin         template.CSS `yaml:"border_thin"`
	Padding            template.CSS `yaml:"padding"`
}

type CardStyle struct {
	Borders      *bool
	BorderRadius template.CSS `yaml:"border_radius"`
	BorderWidth  template.CSS `yaml:"border_width"`
	Padding      template.CSS `yaml:"padding"`
	BorderColor  template.CSS `yaml:"border_color"`
	FontSize     int          `yaml:"font_size"`
	Custom       template.CSS
}

func newBool(b bool) *bool {
	return &b
}

func newFalse() *bool {
	f := false
	return &f
}

func createDefaultModal() *CardStyle {
	ms := CardStyle{
		BorderColor: "rgba(130,185,255,0.15)",
	}
	return &ms
}

func createDefaultBadge() *CardStyle {
	bs := CardStyle{
		BorderColor: "rgba(255,255,255,0.15)",
		BorderWidth: "0.2px",
		Padding:     "0.35em 0.60em",
	}
	return &bs
}

func createDefaultBadgeButton() *CardStyle {
	bs := createDefaultBadge()
	bs.Padding = "0.4em 0.7em"
	return bs
}

func createDefaultButton() *CardStyle {
	return &CardStyle{
		Padding: "1em",
		Borders: newFalse(),
	}
}

func createDefaultTooltip() *CardStyle {
	return &CardStyle{
		Padding: "1em",
	}
}

var defaultTheme = Theme{
	IsLight:    false,
	Background: "#0f0f0f",
	Colors: Colors{
		Surface:          "rgba(18,18,18,0.75)",
		SurfaceOpaque:    "#1a1a1a",
		SurfaceProminent: "rgba(57, 57, 57, 0.85)",
		SurfaceSubtle:    "rgba(255,255,255,0.07)",
		SurfaceAlt:       "rgba(66, 61, 42, 0.85)",
		Border:           "rgba(130, 185, 255, 0.15)",
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
	},
	FontSize: 18,
	Shapes: Shapes{
		Padding:            "0",
		Borders:            newBool(true),
		TightBorderRadius:  "0.2em",
		WideBorderRadius:   "2em",
		MediumBorderRadius: "1em",
		BorderThick:        "1.8px",
		BorderThin:         "0.5px",
	},
	Modals:       createDefaultModal(),
	Badges:       createDefaultBadge(),
	BadgeButtons: createDefaultBadgeButton(),
	Buttons:      createDefaultButton(),
	Tooltips:     createDefaultTooltip(),
}

type ThemesMap map[string]Theme

func deriveAlpha(base template.CSS, alpha float64) template.CSS {
	if base == "" {
		return ""
	}
	c, err := csscolorparser.Parse(string(base))
	if err != nil {
		return base
	}
	c.A = c.A * alpha
	return template.CSS(c.RGBString())
}

func (t *Theme) ComputeDerivatives() {
	// Theme colors
	if t.Colors.SurfaceProminent == "" && t.Colors.Surface != "" {
		t.Colors.SurfaceProminent = t.Colors.Surface
	}
	if t.Colors.SurfaceSubtle == "" && t.Colors.Surface != "" {
		t.Colors.SurfaceSubtle = deriveAlpha(t.Colors.Surface, 0.09)
	}
	if t.Colors.OnSurfaceMuted == "" && t.Colors.OnSurface != "" {
		t.Colors.OnSurfaceMuted = deriveAlpha(t.Colors.OnSurface, 0.6)
	}
	if t.Colors.OnSurfaceSubtle == "" && t.Colors.OnSurface != "" {
		t.Colors.OnSurfaceSubtle = deriveAlpha(t.Colors.OnSurface, 0.35)
	}
	if t.Colors.OnBackground == "" && t.Colors.OnSurface != "" {
		t.Colors.OnBackground = t.Colors.OnSurface
	}

	if t.Colors.AccentMuted == "" && t.Colors.Accent != "" {
		t.Colors.AccentMuted = deriveAlpha(t.Colors.Accent, 0.5)
	}
	if t.Colors.InteractiveMuted == "" && t.Colors.Interactive != "" {
		t.Colors.InteractiveMuted = deriveAlpha(t.Colors.Interactive, 0.6)
	}
	if t.Colors.InteractiveDisabled == "" && t.Colors.Interactive != "" {
		t.Colors.InteractiveDisabled = deriveAlpha(t.Colors.Interactive, 0.5)
	}

	// Precalculated derivations
	t.Derived.PositiveMuted = deriveAlpha(t.Colors.Positive, 0.5)
	t.Derived.NegativeMuted = deriveAlpha(t.Colors.Negative, 0.5)
	t.Derived.StateOnMuted = deriveAlpha(t.Colors.StateOn, 0.35)
	t.Derived.StateOffMuted = deriveAlpha(t.Colors.StateOff, 0.35)
	t.Derived.StateDisabledMuted = deriveAlpha(t.Colors.StateDisabled, 0.35)
}

func (t *Theme) Clone() Theme {
	clone := *t

	cloneCard := func(cs *CardStyle) *CardStyle {
		if cs == nil {
			return nil
		}
		c := *cs
		return &c
	}

	clone.Cards = cloneCard(t.Cards)
	clone.Entities = cloneCard(t.Entities)
	clone.Sensors = cloneCard(t.Sensors)
	clone.Widgets = cloneCard(t.Widgets)
	clone.Badges = cloneCard(t.Badges)
	clone.BadgeButtons = cloneCard(t.BadgeButtons)
	clone.Tooltips = cloneCard(t.Tooltips)
	clone.Modals = cloneCard(t.Modals)
	clone.Buttons = cloneCard(t.Buttons)

	clone.Shapes.Borders = newBool(*t.Shapes.Borders)

	return clone
}

func mergeTheme(base, override Theme) Theme {
	result := base.Clone()
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

func resolveColor(input template.CSS, vars map[string]template.CSS) (template.CSS, error) {
	val := string(input)
	if !strings.Contains(val, "$") {
		return input, nil
	}

	result := val
	re := regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_]*)(?::([0-9]*\.?[0-9]+))?`)
	var firstErr error

	result = re.ReplaceAllStringFunc(result, func(match string) string {
		if firstErr != nil {
			return match
		}
		parts := re.FindStringSubmatch(match)
		varName := parts[1]

		rawColor, ok := vars[varName]
		if !ok {
			firstErr = fmt.Errorf("variable %q not found", varName)
			return match
		}

		if parts[2] == "" {
			return string(rawColor)
		}

		alpha, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			firstErr = fmt.Errorf("invalid alpha for %s: %v", varName, err)
			return match
		}

		c, err := csscolorparser.Parse(string(rawColor))
		if err != nil {
			firstErr = fmt.Errorf("could not parse color %s: %v", rawColor, err)
			return match
		}
		c.A = alpha
		return c.RGBString()
	})

	return template.CSS(result), firstErr
}

func resolveCSS(ptr interface{}, vars map[string]template.CSS) error {
	v := reflect.ValueOf(ptr).Elem()
	t := v.Type()
	cssType := reflect.TypeOf(template.CSS(""))

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() || field.Type() != cssType {
			continue
		}
		resolved, err := resolveColor(template.CSS(field.String()), vars)
		if err != nil {
			return fmt.Errorf("field %s: %w", t.Field(i).Name, err)
		}
		field.Set(reflect.ValueOf(resolved))
	}
	return nil
}

func (t *Theme) Resolve() error {
	if t.Vars == nil {
		return nil
	}

	if err := resolveCSS(&t.Colors, t.Vars); err != nil {
		return err
	}
	if err := resolveCSS(&t.Shapes, t.Vars); err != nil {
		return err
	}
	if t.Background != "" {
		resolved, err := resolveColor(t.Background, t.Vars)
		if err != nil {
			return err
		}
		t.Background = resolved
	}

	// CardStyles
	for _, cs := range []*CardStyle{t.Cards, t.Entities, t.Sensors, t.Widgets,
		t.Modals, t.Badges, t.BadgeButtons, t.Buttons, t.Tooltips} {
		if cs == nil {
			continue
		}
		if err := resolveCSS(cs, t.Vars); err != nil {
			return err
		}
	}

	t.ComputeDerivatives()
	return nil
}

func parseThemes() (*ThemesMap, error) {
	yamlFile, err := os.ReadFile(yamlPath + "/themes.yaml")
	if err != nil {
		return nil, err
	}

	parsed := ThemesMap{}
	err = yaml.Unmarshal(yamlFile, &parsed)
	if err != nil {
		return nil, err
	}

	for name, theme := range parsed {
		merged := mergeTheme(defaultTheme, theme)

		err := merged.Resolve()
		if err != nil {
			return nil, fmt.Errorf("theme %s: %w", name, err)
		}

		parsed[name] = merged
	}
	return &parsed, nil
}
