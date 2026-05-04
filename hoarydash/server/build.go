package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type dashBuilder struct {
	cfg          Yaml
	themes       ThemesMap
	defaultTheme Theme
	mdiIcons     map[string]string
	iconMap      ComponentIconMap
}
type TemplateData struct {
	Dashboard
	Config
	Name                    string
	DomainClassStateIconMap ComponentIconMapSVG
}

type HAState struct {
	EntityID   string `json:"entity_id"`
	Attributes struct {
		FriendlyName string `json:"friendly_name"`
		Icon         string `json:"icon"`
		Class        string `json:"device_class"`
	} `json:"attributes"`
}

func domain(entityID string) string {
	parts := strings.SplitN(entityID, ".", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

func makeOnceFunc() func(string) bool {
	seen := make(map[string]bool)
	return func(key string) bool {
		if seen[key] {
			return false
		}
		seen[key] = true
		return true
	}
}

func makeGlobals(globals map[string]any) func(string) (any, error) {
	return func(key string) (any, error) {
		if val, ok := globals[key]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("Global value '%s' not found", key)
	}
}

func makeTranslate(lang string) func(key string) (string, error) {
	return func(key string) (string, error) {
		return translate(key, lang)
	}
}

func makeDomainTranslations(lang string) func(domain string) (map[string]string, error) {
	return func(domain string) (map[string]string, error) {
		return domainTranslations(domain, lang)
	}
}

func makeUid() func() int {
	uid := 0
	return func() int {
		uid++
		return uid
	}
}

func newDashBuilder(cfg Yaml) (*dashBuilder, error) {
	themes, err := loadThemes(&cfg)
	if err != nil {
		return nil, fmt.Errorf("load themes: %w", err)
	}
	defaultTheme, err := getDefaultTheme()
	if err != nil {
		return nil, fmt.Errorf("default theme: %w", err)
	}
	iconMap, err := fetchComponentIcons(cfg.HA)
	if err != nil {
		return nil, fmt.Errorf("fetch component icons: %w", err)
	}
	mdiIcons, err := loadMdiIcons()
	if err != nil {
		return nil, fmt.Errorf("loading mdi icons: %w", err)
	}
	return &dashBuilder{
		cfg:          cfg,
		themes:       themes,
		defaultTheme: defaultTheme,
		mdiIcons:     mdiIcons,
		iconMap:      iconMap,
	}, nil
}

func friendlyName(raw string, locale string) string {
	raw = strings.ReplaceAll(raw, "_", " ")
	tag := language.Make(locale)
	return cases.Title(tag).String(raw)
}

func applyState(v *reflect.Value, e Entity, state HAState, locale string, icons ComponentIconMap) {
	if e.Label == "" && state.Attributes.FriendlyName != "" {
		v.FieldByName("Label").SetString(friendlyName(state.Attributes.FriendlyName, locale))
	}

	class := "_"
	if state.Attributes.Class != "" {
		class = state.Attributes.Class
	}
	v.FieldByName("Class").SetString(class)

	if e.Icon == "" {
		icon := state.Attributes.Icon
		if icon == "" {
			if domainIcons, ok := icons[domain(e.EntityID)]; ok {
				if classIcons, ok := domainIcons[class]; ok {
					icon = classIcons.Default
				} else if fallback, ok := domainIcons["_"]; ok {
					icon = fallback.Default
				}
			}
		}
		if icon != "" {
			icon = strings.TrimPrefix(icon, "mdi:")
			v.FieldByName("Icon").SetString(icon)
		}
	}
}

func enrichEntities(dashboard Dashboard, cfg Yaml, icons ComponentIconMap) (map[string]HAState, DomainClassSet, error) {
	states := map[string]HAState{}
	domainClasses := DomainClassSet{}

	err := walkEntities(dashboard, func(f reflect.StructField, v *reflect.Value) error {
		e := v.Interface().(Entity)
		if e.EntityID == "" {
			return nil
		}

		var state HAState
		if cached, already := states[e.EntityID]; already {
			state = cached
		} else {
			fetched, err := fetchEntityState(e.EntityID, cfg.HA)
			if err != nil {
				return err
			}
			states[e.EntityID] = fetched
			state = fetched
		}

		applyState(v, e, state, cfg.Localization.Locale, icons)

		domainClasses.add(domain(state.EntityID), state.Attributes.Class)

		return nil
	})

	return states, domainClasses, err
}

func fetchEntityState(id string, ha HAConfig) (HAState, error) {
	req, err := http.NewRequest("GET", ha.HTTPURL+"/api/states/"+id, nil)
	if err != nil {
		return HAState{}, err
	}
	req.Header.Set("Authorization", "Bearer "+ha.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return HAState{}, err
	}
	defer resp.Body.Close()
	var state HAState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return HAState{}, err
	}
	return state, nil
}

// stateIconMap runs the full icon pipeline for one dashboard:
// enrich entities → collect domain/class set → filter full icon map → resolve to SVGs.
func (b *dashBuilder) stateIconMap(dash Dashboard) ComponentIconMapSVG {
	_, domainClasses, _ := enrichEntities(dash, b.cfg, b.iconMap)
	return resolveIconMapSVG(filterIconMap(b.iconMap, domainClasses), b.mdiIcons)
}

// resolveThemes applies dashboard- and screen-level theme resolution,
// mutating a copy of the dashboard and returning it.
func (b *dashBuilder) resolveThemes(name string, dash Dashboard) (Dashboard, error) {
	resolvedDashTheme, err := buildTheme(b.themes, dash.ThemeRef)
	if err != nil {
		return dash, fmt.Errorf("dashboard %s theme: %w", name, err)
	}
	dashTheme := Theme{}
	if resolvedDashTheme != nil {
		dashTheme = *resolvedDashTheme
	}
	dashTheme.inheritTheme(b.defaultTheme)
	dash.Theme = dashTheme

	for i, screen := range dash.Screens {
		resolvedScreenTheme, err := buildTheme(b.themes, screen.ThemeRef)
		if err != nil {
			return dash, fmt.Errorf("screen[%d] (%s) theme: %w", i, screen.Name, err)
		}
		dash.Screens[i].Theme = resolvedScreenTheme

		effectiveTheme := resolvedScreenTheme
		if effectiveTheme == nil {
			effectiveTheme = &dash.Theme
		} else {
			effectiveTheme.inheritVars(dash.Theme)
		}
		if err := walkAndResolveCSS(&dash.Screens[i], effectiveTheme.Vars); err != nil {
			return dash, fmt.Errorf("screen[%d] (%s) css resolution: %w", i, screen.Name, err)
		}
	}

	return dash, nil
}

// prepareData produces the full TemplateData for one dashboard.
func (b *dashBuilder) prepareData(name string, dash Dashboard) (TemplateData, error) {
	enrichedDash, err := b.resolveThemes(name, dash)
	if err != nil {
		return TemplateData{}, err
	}
	return TemplateData{
		Dashboard:               enrichedDash,
		Config:                  b.cfg.Config,
		Name:                    name,
		DomainClassStateIconMap: b.stateIconMap(enrichedDash),
	}, nil
}

var funcMap = template.FuncMap{
	"default": func(def any, val any) any {
		if val == nil {
			return def
		}
		v := reflect.ValueOf(val)

		if v.Kind() == reflect.Ptr && v.IsNil() {
			return def
		}
		if v.Kind() == reflect.String && v.String() == "" {
			return def
		}
		if v.Kind() == reflect.Int && v.Int() == 0 {
			return def
		}

		return val
	},
	"mergeDateclock": mergeOverride[map[string]any, Dateclock],
	"bool": func(val *bool) bool {
		if val != nil {
			return *val
		}
		return false
	},
	"css": func(val any) template.CSS {
		return template.CSS(fmt.Sprintf("%v", val))
	},
	"enabledByDefault": enabledByDefault,
	"disabledByDefault": func(v *bool) bool {
		if v == nil {
			return false
		}
		return *v
	},
	"dict": func(values ...any) map[string]any {
		m := map[string]any{}
		for i := 0; i < len(values); i += 2 {
			key := values[i].(string)
			m[key] = values[i+1]
		}
		return m
	},
	"domainIn": func(entityID string, domains ...string) bool {
		domain := domain(entityID)
		for _, d := range domains {
			if domain == d {
				return true
			}
		}
		return false
	},
	"anyOfIn": func(anyOf []string, in ...string) bool { // O(n) is n^2 but the lists are tiny so its fine
		for _, is := range anyOf {
			for _, of := range in {
				if is == of {
					return true
				}
			}
		}
		return false
	},
	"domain": domain,
	"domains": func(entityIDs []string) []string {
		var out []string
		for _, id := range entityIDs {
			out = append(out, domain(id))
		}
		return out
	},
	"entityIDs": func(entities []Card) []string {
		out := make([]string, len(entities))
		for i, e := range entities {
			out[i] = e.EntityID
		}
		return out
	},
	"isEmoji": func(s string) bool {
		for _, r := range s {
			return r > 127
		}
		return false
	},
	"json": jsonStr,
	"merge": func(maps ...any) (map[string]any, error) {
		result := map[string]any{}
		for _, m := range maps {
			switch v := m.(type) {
			case map[string]any:
				for k, val := range v {
					result[k] = val
				}
			default:
				rv := reflect.ValueOf(m)
				if rv.Kind() == reflect.Ptr {
					rv = rv.Elem()
				}
				if rv.Kind() != reflect.Struct {
					return nil, fmt.Errorf("merge: unsupported type %T", m)
				}
				rt := rv.Type()
				for i := 0; i < rv.NumField(); i++ {
					f := rt.Field(i)
					if f.IsExported() {
						result[f.Name] = rv.Field(i).Interface()
					}
				}
			}
		}
		return result, nil
	},
	"concat": func(values ...any) string {
		var out strings.Builder
		for _, v := range values {
			out.WriteString(fmt.Sprint(v))
		}
		return out.String()
	},
	"prevScreen": func(d Dashboard, i int) *Screen {
		if i > 0 {
			return &d.Screens[i-1]
		}
		return nil
	},
	"nextScreen": func(d Dashboard, i int) *Screen {
		if i+1 < len(d.Screens) {
			return &d.Screens[i+1]
		}
		return nil
	},
	"once":    nilfunc,
	"globals": nilfunc,
	"uid":     nilfunc,
	"icon":    nilfunc,
	"replace": func(old string, new string, s string) string {
		return strings.ReplaceAll(s, old, new)
	},
	"isFull": func(layout string) bool {
		sArr := strings.Split(layout, "-")
		if len(sArr) > 1 {
			return sArr[1] == "full"
		}
		return false
	},
	"translate":          nilfunc,
	"domainTranslations": nilfunc,
	"add": func(vals ...float64) float64 {
		var res float64
		for _, val := range vals {
			res += val
		}
		return res
	},
	"resolveThis": resolveThis,
}

// makeFuncMap returns a fresh FuncMap with all per-dashboard closures bound.
// Called once per dashboard, after data is prepared, so closures capture
// the correct per-dashboard state.
func (b *dashBuilder) makeFuncMap(cfg Yaml, name string) template.FuncMap {
	lang := "en"
	if cfg.Localization.Locale != "" {
		lang = strings.Split(cfg.Localization.Locale, "-")[0]
	}
	globals := map[string]any{
		"Animations": enabledByDefault(cfg.Dashboards[name].Animations),
		"Lang":       lang,
		"IsDev":      isDev,
	}

	fm := make(template.FuncMap, len(funcMap))
	for k, v := range funcMap {
		fm[k] = v
	}
	fm["once"] = makeOnceFunc()
	fm["globals"] = makeGlobals(globals)
	fm["translate"] = makeTranslate(lang)
	fm["domainTranslations"] = makeDomainTranslations(lang)
	fm["uid"] = makeUid()
	fm["icon"] = func(name string) template.HTML {
		return iconToSVG(name, &b.mdiIcons)
	}
	return fm
}

// loadTemplates parses all template globs and returns the root template.
func loadTemplates() (*template.Template, error) {
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(frontendPath + "/templates/*.html.tmpl")
	if err != nil {
		return nil, fmt.Errorf("root templates: %w", err)
	}
	globs := []string{
		"/templates/css/*.html.tmpl",
		"/templates/css/*.css.tmpl",
		"/templates/entities/*.html.tmpl",
		"/templates/widgets/*.html.tmpl",
		"/templates/controllers/*.html.tmpl",
		"/templates/navbar-styles/*.html.tmpl",
		"/templates/layouts/*.html.tmpl",
		"/templates/common/*.html.tmpl",
	}
	for _, g := range globs {
		if tmpl, err = tmpl.ParseGlob(frontendPath + g); err != nil {
			return nil, fmt.Errorf("templates glob %s: %w", g, err)
		}
	}
	return tmpl, nil
}

type builtDash struct {
	tmpl *template.Template
	data TemplateData
}

func writeOutput(name string, b builtDash) {
	outputDir := frontendPath + "/static/" + name
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Could not create output dir for %s: %v", name, err)
		return
	}
	out, err := os.Create(outputDir + "/index.html")
	if err != nil {
		log.Printf("Could not create output file for %s: %v", name, err)
		return
	}
	defer out.Close()
	if err := b.tmpl.ExecuteTemplate(out, "main.html.tmpl", b.data); err != nil {
		log.Printf("Could not execute template for %s: %v", name, err)
		return
	}
	out.Sync()
}

func BuildDash() {
	cfg, err := loadConfig()
	if err != nil {
		log.Printf("Could not load config when building dashboard: %v", err)
		return
	}
	BuildDashFromConfig(cfg)
}

func BuildDashFromConfig(cfg Yaml) {
	builder, err := newDashBuilder(cfg)
	if err != nil {
		log.Printf("Build setup failed: %v", err)
		return
	}

	tmpl, err := loadTemplates()
	if err != nil {
		log.Printf("Could not load templates: %v", err)
		return
	}

	// First pass: prepare data and clone templates.
	//
	// The two-pass approach is load-bearing. FuncMap closures (once, uid, globals, etc.)
	// are per-dashboard and must be bound before cloning. If we clone-then-execute in the
	// same loop, the next iteration mutates the funcmap on an already-cloned template,
	// breaking the closures for previously cloned dashboards.
	built := make(map[string]builtDash, len(cfg.Dashboards))
	for name, dash := range cfg.Dashboards {
		if isDev {
			log.Printf("=== Preprocessing dashboard '%s' ===", name)
		}

		data, err := builder.prepareData(name, dash)
		if err != nil {
			log.Printf("Could not prepare data for %s: %v", name, err)
			continue
		}

		fm := builder.makeFuncMap(cfg, name)

		t, err := tmpl.Clone()
		if err != nil {
			log.Printf("Could not clone template for %s: %v", name, err)
			continue
		}

		built[name] = builtDash{
			tmpl: t.Funcs(fm),
			data: data,
		}

		if isDev {
			log.Printf("=== Finished preprocessing '%s' ===", name)
		}
	}

	for name, b := range built {
		writeOutput(name, b)
	}
}
