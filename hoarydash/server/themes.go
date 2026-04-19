package main

import (
	"fmt"
	"html/template"
	"log"
	"maps"
	"math"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/mazznoer/csscolorparser"
	"go.yaml.in/yaml/v4"
)

type Theme struct {
	IsLight           bool
	Vars              map[string]template.CSS
	Background        *BackgroundLayer `yaml:"background"`
	BackgroundOverlay *BackgroundLayer `yaml:"background_overlay"`
	Colors
	Derived derivedColors `yaml:"-"`

	Shapes
	Typography

	Custom template.CSS

	// Font size
	FontSize float64 `yaml:"font_size"`

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

type Typography struct {
	FontFamily    template.CSS `yaml:"font_family"`
	FontWeight    template.CSS `yaml:"font_weight"`
	FontStyle     template.CSS `yaml:"font_style"`
	TextTransform template.CSS `yaml:"text_transform"`
	LetterSpacing template.CSS `yaml:"letter_spacing"`
}
type BackgroundLayer struct {
	Color     template.CSS `yaml:"color"`
	Image     template.CSS `yaml:"image"`
	Size      template.CSS `yaml:"size"`
	Position  template.CSS `yaml:"position"`
	Repeat    template.CSS `yaml:"repeat"`
	BlendMode template.CSS `yaml:"blend_mode"`
	Filter    template.CSS `yaml:"filter"`
	Opacity   *float64     `yaml:"opacity"`
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
	GapInner           template.CSS `yaml:"gap_inner"`
	GapOuter           template.CSS `yaml:"gap_outer"`
	PaddingInner       template.CSS `yaml:"padding_inner"`
}

type CardStyle struct {
	Borders      *bool
	BorderRadius template.CSS `yaml:"border_radius"`
	BorderWidth  template.CSS `yaml:"border_width"`
	Background   template.CSS
	Padding      template.CSS `yaml:"padding"`
	BorderColor  template.CSS `yaml:"border_color"`
	FontSize     float64      `yaml:"font_size"`
	Custom       template.CSS
}

func newPtr[T any](val T) *T {
	return &val
}

var defaultTheme = Theme{
	IsLight: false,
	Background: newPtr(BackgroundLayer{
		Color: "#0f0f0f",
	}),
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
		OnBackground:     "#ffffff",
		Accent:           "hsl(210, 90%, 65%)",
		Interactive:      "hsl(210, 90%, 65%)",
		StateOn:          "hsl(134, 60%, 45%)",
		StateOff:         "rgba(200,220,240,0.38)",
		StateDisabled:    "rgba(220, 240, 250, 0.25)",
		Positive:         "hsl(140, 60%, 55%)",
		Negative:         "hsl(0, 70%, 55%)",
	},
	FontSize: 1.0,
	Shapes: Shapes{
		GapInner:           "0.6em",
		GapOuter:           "1em",
		PaddingInner:       "0.4em",
		Borders:            newPtr(true),
		TightBorderRadius:  "0.2em",
		WideBorderRadius:   "2em",
		MediumBorderRadius: "0.75em",
		BorderThick:        "2px",
	},
	Entities: newPtr(CardStyle{
		Padding: "0.4em 0.75em",
	}),
	Cards: newPtr(CardStyle{
		Padding: "0.75em",
	}),
	Modals: newPtr(CardStyle{
		BorderColor: "rgba(130,185,255,0.15)",
	}),
	Badges: newPtr(CardStyle{
		BorderColor: "rgba(255,255,255,0.15)",
		Padding:     "0.35em 0.60em",
	}),
	BadgeButtons: newPtr(CardStyle{
		BorderColor: "rgba(255,255,255,0.15)",
		Padding:     "0.4em 0.7em",
	}),
	Buttons: newPtr(CardStyle{
		Padding: "1em",
		Borders: newPtr(false),
	}),
	Tooltips: newPtr(CardStyle{
		Padding: "1em",
	}),
}

type ThemesMap map[string]Theme

type VarResolver func(varName string, alpha *float64) (string, error)

func toRGBAString(c csscolorparser.Color) string {
	r := int(math.Round(c.R * 255))
	g := int(math.Round(c.G * 255))
	b := int(math.Round(c.B * 255))
	return fmt.Sprintf("rgba(%d, %d, %d, %.4g)", r, g, b, c.A)
}

func deriveAlpha(base template.CSS, alpha float64) template.CSS {
	if base == "" {
		return ""
	}
	c, err := csscolorparser.Parse(string(base))
	if err != nil {
		return base
	}
	c.A = c.A * alpha
	return template.CSS(toRGBAString(c))
}

func (t *Theme) ComputeDerivatives() error {
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
	if isDev {
		// log.Printf("[derive] StateOn=%q Positive=%q Negative=%q StateOff=%q StateDisabled=%q",
		// 	t.Colors.StateOn,
		// 	t.Colors.Positive,
		// 	t.Colors.Negative,
		// 	t.Colors.StateOff,
		// 	t.Colors.StateDisabled,
		// )
	}
	t.Derived.PositiveMuted = deriveAlpha(t.Colors.Positive, 0.5)
	t.Derived.NegativeMuted = deriveAlpha(t.Colors.Negative, 0.5)
	t.Derived.StateOnMuted = deriveAlpha(t.Colors.StateOn, 0.3)
	t.Derived.StateOffMuted = deriveAlpha(t.Colors.StateOff, 0.3)
	t.Derived.StateDisabledMuted = deriveAlpha(t.Colors.StateDisabled, 0.3)
	if isDev {
		// log.Printf("[derive] result: StateOnMuted=%q StateOffMuted=%q PositiveMuted=%q",
		// 	t.Derived.StateOnMuted, t.Derived.StateOffMuted, t.Derived.PositiveMuted,
		// )
	}

	if t.Shapes.BorderThin == "" && t.Shapes.BorderThick != "" {
		css, err := computeNumericCSS(t.Shapes.BorderThick, func(val float64) float64 { return val / 2 })
		if err != nil {
			return fmt.Errorf("could not derive BorderThin when finalizing theme: %v", err)
		}
		t.Shapes.BorderThin = css
	}

	if t.Shapes.GapInner == "" && t.Shapes.GapOuter != "" {
		css, err := computeNumericCSS(t.Shapes.GapOuter, func(val float64) float64 { return val / 2 })
		if err != nil {
			return fmt.Errorf("could not derive MarginInner when finalizing theme: %v", err)
		}
		t.Shapes.GapInner = css
	}

	return nil
}

func computeNumericCSS(str template.CSS, calculation func(float64) float64) (template.CSS, error) {
	var cssValueRegex = regexp.MustCompile(`^([0-9\.]+)\s*([a-zA-Z%]*)$`)
	matches := cssValueRegex.FindStringSubmatch(strings.TrimSpace(string(str)))

	if len(matches) != 3 {
		return "", fmt.Errorf("could not match a value and unit in string '%s'", str)
	}
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", fmt.Errorf("could not parse float from '%s':%w", str, err)
	}
	thinValue := calculation(value)
	unit := matches[2]
	return template.CSS(fmt.Sprintf("%g%s", thinValue, unit)), nil
}

func clonePtr[T any](ptr *T) *T {
	if ptr == nil {
		return nil
	}
	b := *ptr
	return &b
}

func (t *Theme) Clone() Theme {
	clone := *t

	clone.Background = clonePtr(t.Background)
	clone.BackgroundOverlay = clonePtr(t.BackgroundOverlay)

	clone.Cards = clonePtr(t.Cards)
	clone.Entities = clonePtr(t.Entities)
	clone.Sensors = clonePtr(t.Sensors)
	clone.Widgets = clonePtr(t.Widgets)
	clone.Badges = clonePtr(t.Badges)
	clone.BadgeButtons = clonePtr(t.BadgeButtons)
	clone.Tooltips = clonePtr(t.Tooltips)
	clone.Modals = clonePtr(t.Modals)
	clone.Buttons = clonePtr(t.Buttons)

	if t.Shapes.Borders != nil {
		clone.Shapes.Borders = newPtr(*t.Shapes.Borders)
	}

	return clone
}

func mergeTheme(base, override Theme) Theme {
	if override.Cards != nil {
		base.Entities = nil
		base.Sensors = nil
		base.Widgets = nil
	}
	return mergeOverride(base, override)
}

func resolveTheme(named ThemesMap, ref ThemeRef) (*Theme, error) {
	if ref.Name != "" {
		name := ref.Name
		if namedTheme, ok := named[name]; ok {
			cloned := namedTheme.Clone()
			return &cloned, nil
		}
		return nil, fmt.Errorf("theme %q not found", name)
	}

	if ref.Theme == nil {
		return nil, nil
	}

	cloned := ref.Theme.Clone()
	if err := cloned.Finalize(); err != nil {
		return nil, err
	}
	return &cloned, nil
}

func resolveThis(input template.CSS, this string) template.CSS {
	resolver := func(varName string, _ *float64) (string, error) {
		if varName == "this" {
			return this, nil
		}
		return varName, nil
	}

	if output, err := replaceVars(input, resolver); err != nil {
		return input
	} else {
		return output
	}
}

func createResolveVar(vars map[string]template.CSS) VarResolver {
	return func(varName string, alpha *float64) (string, error) {
		if varName == "this" {
			return "", nil
		}

		rawColor, ok := vars[varName]
		if !ok {
			return "", fmt.Errorf("variable %q not found", varName)
		}

		if alpha == nil {
			return string(rawColor), nil
		}

		c, err := csscolorparser.Parse(string(rawColor))
		if err != nil {
			return "", fmt.Errorf("could not parse color %s: %v", rawColor, err)
		}
		c.A = *alpha
		return toRGBAString(c), nil
	}
}

func replaceVars(input template.CSS, replacer VarResolver) (template.CSS, error) {
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
		var alpha *float64 = nil
		if parts[2] != "" {
			alphaConcrete, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				firstErr = fmt.Errorf("invalid alpha for %s: %v", varName, err)
				return match
			}
			alpha = &alphaConcrete
		}

		resolved, err := replacer(varName, alpha)

		if err != nil {
			firstErr = err
		}

		if resolved != "" {
			return resolved
		} else {
			return match
		}

	})

	return template.CSS(result), firstErr
}

func resolveCSSFields(ptr interface{}, repl VarResolver) error {
	v := reflect.ValueOf(ptr).Elem()
	t := v.Type()
	cssType := reflect.TypeOf(template.CSS(""))

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() || field.Type() != cssType {
			continue
		}
		resolved, err := replaceVars(template.CSS(field.String()), repl)
		if err != nil {
			return fmt.Errorf("field %s: %w", t.Field(i).Name, err)
		}
		field.Set(reflect.ValueOf(resolved))
	}
	return nil
}

func (t *Theme) Finalize() error {
	if t.Vars != nil {
		resolver := createResolveVar(t.Vars)
		if err := resolveCSSFields(&t.Colors, resolver); err != nil {
			return err
		}
		if err := resolveCSSFields(&t.Shapes, resolver); err != nil {
			return err
		}
		if err := resolveCSSFields(&t.Typography, resolver); err != nil {
			return err
		}
		if t.Background != nil {
			if err := resolveCSSFields(t.Background, resolver); err != nil {
				return err
			}
		}
		if t.BackgroundOverlay != nil {
			if err := resolveCSSFields(t.BackgroundOverlay, resolver); err != nil {
				return err
			}
		}

		// CardStyles
		for _, cs := range []*CardStyle{t.Cards, t.Entities, t.Sensors, t.Widgets,
			t.Modals, t.Badges, t.BadgeButtons, t.Buttons, t.Tooltips} {
			if cs == nil {
				continue
			}
			if err := resolveCSSFields(cs, resolver); err != nil {
				return err
			}
		}

		if t.Custom != "" {
			resolved, err := replaceVars(t.Custom, resolver)
			if err != nil {
				return err
			}
			t.Custom = resolved
		}
	}

	t.ComputeDerivatives()
	return nil
}

func parseThemes() (*ThemesMap, error) {
	merged := ThemesMap{}

	bundled, err := os.ReadFile(appPath + "/themes.bundled.yaml")
	if err != nil {
		return nil, fmt.Errorf("bundled themes missing: %w", err)
	}
	if err := yaml.Unmarshal(bundled, &merged); err != nil {
		return nil, err
	}

	userFilePath := configPath + "/themes.yaml"
	if isDev {
		userFilePath = configPath + "/themes.dev.yaml"
	}
	userFile, err := os.ReadFile(userFilePath)
	if err == nil {
		log.Printf("Found user defined themes.yaml")
		user := ThemesMap{}
		if err := yaml.Unmarshal(userFile, &user); err != nil {
			return nil, err
		}
		for name, theme := range user {
			merged[name] = theme
		}
		log.Printf("parsed themes: %+v", slices.Collect(maps.Keys(user)))
	}

	for name, theme := range merged {
		if err := theme.Finalize(); err != nil {
			return nil, fmt.Errorf("theme %s: %w", name, err)
		}
		merged[name] = theme
	}
	return &merged, nil
}
