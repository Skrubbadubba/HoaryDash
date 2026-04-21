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
	"github.com/mitchellh/reflectwalk"
	"go.yaml.in/yaml/v4"
)

type Theme struct {
	IsLight           bool
	Base              *string
	Vars              ThemeVars
	Background        *BackgroundLayer `yaml:"background"`
	BackgroundOverlay *BackgroundLayer `yaml:"background_overlay"`
	Colors            Colors
	Derived           derivedColors `yaml:"-"`

	Shapes     Shapes
	Typography Typography

	Custom template.CSS

	// Size multiplier
	Size float64

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

type ThemeVars map[string]template.CSS

type Typography struct {
	FontFamily    template.CSS `yaml:"font_family"`
	FontWeight    template.CSS `yaml:"font_weight"`
	FontStyle     template.CSS `yaml:"font_style"`
	TextTransform template.CSS `yaml:"text_transform"`
	LetterSpacing template.CSS `yaml:"letter_spacing"`
	FontXXS       template.CSS `yaml:"font_xxs"`
	FontXS        template.CSS `yaml:"font_xs"`
	FontSM        template.CSS `yaml:"font_sm"`
	FontMD        template.CSS `yaml:"font_md"`
	FontLG        template.CSS `yaml:"font_lg"`
	FontXL        template.CSS `yaml:"font_xl"`
	FontXXL       template.CSS `yaml:"font_xxl"`
	FontHero      template.CSS `yaml:"font_hero"`
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
	Size         float64      `yaml:"size"`
	Custom       template.CSS
}

func newPtr[T any](val T) *T {
	return &val
}

func getDefaultTheme() (Theme, error) {
	defaultTheme := Theme{
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
		Typography: Typography{
			FontXXS:  "0.55em",
			FontXS:   "0.65em",
			FontSM:   "0.7em",
			FontMD:   "1em",
			FontLG:   "1.2em",
			FontXL:   "1.7em",
			FontXXL:  "2.2em",
			FontHero: "6em",
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
	if err := defaultTheme.Finalize(); err != nil {
		return Theme{}, fmt.Errorf("Error finalizing default theme: %v", err)
	}
	return defaultTheme, nil
}

type ThemesMap map[string]Theme

type CSSResolver func(varName string, multiplier *float64) (string, error)

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

func buildTheme(named ThemesMap, ref ThemeRef) (*Theme, error) {
	theme, err := lookupThemeRef(named, ref)
	if err != nil {
		return nil, err
	}
	if theme == nil {
		return nil, nil
	}

	if isDev {
		name := "[inline]"
		if ref.Name != "" {
			name = ref.Name
		}
		log.Printf("--- constructing theme '%s', current state: ---\n%s", name, jsonStr(theme))
	}

	if defaultTheme, err := getDefaultTheme(); err == nil {
		theme.inheritVars(defaultTheme)
	}

	built, err := buildThemeRec(named, *theme, map[string]bool{})
	if err != nil {
		return nil, err
	}

	if isDev {
		log.Printf("final state:\n%s\n--- theme construction finished ---", jsonStr(built))
	}

	return &built, nil
}

func buildThemeRec(named ThemesMap, theme Theme, seen map[string]bool) (Theme, error) {
	if theme.Base != nil {
		name := *theme.Base

		if seen[name] {
			return Theme{}, fmt.Errorf("cyclic theme reference: %s", name)
		}
		seen[name] = true

		base, err := lookupThemeName(named, name)
		if err != nil {
			return Theme{}, err
		}

		baseTheme, err := buildThemeRec(named, base, seen)
		if err != nil {
			return Theme{}, fmt.Errorf("error calculating base theme '%s': %w", name, err)
		}

		theme.inheritTheme(baseTheme)
	}

	if err := theme.Finalize(); err != nil {
		return Theme{}, err
	}

	return theme, nil
}

func (t *Theme) inheritVars(base Theme) {
	if t.Vars == nil {
		t.Vars = map[string]template.CSS{}
	}

	for k, v := range base.Vars {
		if _, exists := t.Vars[k]; !exists {
			t.Vars[k] = v
		}
	}
}

func (t *Theme) inheritTheme(base Theme) {
	t.inheritVars(base)

	mergeBase := base.Clone()

	if t.Cards != nil {
		mergeBase.Entities = nil
		mergeBase.Sensors = nil
		mergeBase.Widgets = nil
	}

	*t = mergeOverride(mergeBase, *t)
}

func lookupThemeName(named ThemesMap, name string) (Theme, error) {
	spec, ok := named[name]
	if !ok {
		return Theme{}, fmt.Errorf("theme %q not found", name)
	}
	return spec.Clone(), nil
}

func lookupThemeRef(named ThemesMap, ref ThemeRef) (*Theme, error) {
	if ref.Name != "" {
		t, err := lookupThemeName(named, ref.Name)
		if err != nil {
			return nil, err
		}
		return &t, nil
	}

	if ref.Theme != nil {
		cloned := ref.Theme.Clone()
		return &cloned, nil
	}

	return nil, nil
}

func resolveThis(input template.CSS, this string) template.CSS {
	resolver := func(varName string, _ *float64) (string, error) {
		if varName == "this" {
			return this, nil
		}
		return varName, nil
	}

	if output, err := resolveCSS(input, resolver); err != nil {
		return input
	} else {
		return output
	}
}

func createVarResolver(vars ThemeVars) CSSResolver {
	return func(varName string, multiplier *float64) (string, error) {
		if varName == "this" {
			return "", nil
		}

		rawValue, ok := vars[varName]
		if !ok {
			return "", fmt.Errorf("variable %q not found", varName)
		}

		if multiplier == nil {
			return string(rawValue), nil
		}

		c, err := csscolorparser.Parse(string(rawValue))
		if err != nil {
			return fmt.Sprintf("calc(%s * %.4g)", rawValue, *multiplier), nil
		}
		c.A = *multiplier
		return toRGBAString(c), nil
	}
}

func resolveCSS(input template.CSS, resolver CSSResolver) (template.CSS, error) {
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
		var multiplier *float64 = nil
		if parts[2] != "" {
			multiplierConcrete, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				firstErr = fmt.Errorf("invalid multiplier for %s: %v", varName, err)
				return match
			}
			multiplier = &multiplierConcrete
		}

		resolved, err := resolver(varName, multiplier)

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

type WalkerCallback func(f reflect.StructField, v *reflect.Value) error

type cssWalker struct {
	cb  WalkerCallback
	err error
}

func (w *cssWalker) Struct(reflect.Value) error { return nil }
func (w *cssWalker) StructField(f reflect.StructField, v reflect.Value) error {
	if w.err != nil {
		return nil
	}
	if v.Type() == reflect.TypeOf((*Theme)(nil)) || v.Type() == reflect.TypeOf(Theme{}) {
		return reflectwalk.SkipEntry
	}
	if v.Type() != reflect.TypeOf(template.CSS("")) || v.String() == "" {
		return nil
	}
	w.err = w.cb(f, &v)
	return nil
}

func walkCSS(target any, cb WalkerCallback) error {
	w := &cssWalker{cb: cb}
	if err := reflectwalk.Walk(target, w); err != nil {
		return err
	}
	return w.err
}

func walkAndResolveCSS(target any, vars ThemeVars) error {
	resolver := createVarResolver(vars)
	return walkCSS(target, func(f reflect.StructField, v *reflect.Value) error {
		if !v.CanSet() {
			return nil
		}
		resolved, err := resolveCSS(template.CSS(v.String()), resolver)
		if err != nil {
			return fmt.Errorf("field %s: %w", f.Name, err)
		}
		v.Set(reflect.ValueOf(resolved))
		return nil
	})
}

func (t *Theme) seedSemanticVars() {
	if t.Vars == nil {
		t.Vars = ThemeVars{}
	}
	for _, strct := range []any{t.Colors, t.Shapes, t.Typography} {
		walkCSS(strct, func(f reflect.StructField, v *reflect.Value) error {
			var tag string
			yamlTag := f.Tag.Get("yaml")
			if yamlTag == "" || yamlTag == "-" {
				tag = strings.ToLower(f.Name)
			}
			tag = strings.SplitN(yamlTag, ",", 2)[0]
			if _, exists := t.Vars[tag]; !exists {
				t.Vars[tag] = template.CSS(v.String())
			}
			return nil
		})
	}
}

func (t *Theme) Finalize() error {
	if t == nil {
		return nil
	}
	t.seedSemanticVars()

	if err := walkAndResolveCSS(t, t.Vars); err != nil {
		return err
	}
	return t.ComputeDerivatives()
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
		log.Printf("parsed user themes: %+v", slices.Collect(maps.Keys(user)))
	}

	return &merged, nil
}
