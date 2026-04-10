package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"reflect"
	"strings"
)

//go:embed mdi.json
var mdiData []byte

var mdiIcons map[string]string

type TemplateData struct {
	Dashboard
	Config
	Name  string
	IsDev bool
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

func nilfunc() any { return nil }

var uid = 0

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
	"css": func(val any) template.CSS {
		return template.CSS(fmt.Sprintf("%v", val))
	},
	"enabledByDefault": func(v *bool) bool {
		if v == nil {
			return true
		}
		return *v
	},
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
	"icon": func(name string) template.HTML {
		path := mdiIcons[name]
		return template.HTML(fmt.Sprintf(
			`<svg class="icon" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="%s"/></svg>`,
			path,
		))
	},
	"isEmoji": func(s string) bool {
		for _, r := range s {
			return r > 127
		}
		return false
	},
	"json": func(j interface{}) string { // For debugging
		var out []byte
		out, err := json.Marshal(j)
		if err != nil {
			return ""
		}
		return string(out)
	},
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
	"uid": func() int {
		uid++
		return uid
	},
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
}

func BuildDash() {
	cfg, err := parseConfig()
	if err != nil {
		log.Printf("Could not load config when building dashboard")
		return
	}

	parsedThemes, err := parseThemes()
	if err != nil {
		log.Printf("Could not load config when building dashboard")
		return
	}
	allNamed := ThemesMap{}
	for k, v := range *parsedThemes {
		allNamed[k] = v
	}
	for k, v := range cfg.Themes {
		allNamed[k] = v
	}

	var tmpl *template.Template

	tmpl, err = template.New("").Funcs(funcMap).ParseGlob(frontendPath + "/templates/*.html.tmpl")
	if err != nil {
		log.Printf("Could not return root level templates %v", err)
		return
	}

	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/css/*.html.tmpl")
	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/css/*.css.tmpl")
	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/entities/*.html.tmpl")
	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/widgets/*.html.tmpl")
	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/controllers/*.html.tmpl")
	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/navbar-styles/*.html.tmpl")
	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/layouts/*.html.tmpl")
	tmpl, err = tmpl.ParseGlob(frontendPath + "/templates/common/*.html.tmpl")
	check(err, "Created template object")

	type builtDash struct {
		tmpl *template.Template
		data TemplateData
	}

	// First pass, to enrich template data and create function enclosures for each dashboard.
	// Function enclosures break if template (along with its funcmap)
	// is copied after it has been executed
	built := make(map[string]builtDash)
	for name, dash := range cfg.Dashboards {
		for i, screen := range dash.Screens {
			resolvedTheme, err := resolveTheme(allNamed, screen.ThemeRef)
			if err != nil {
				log.Printf("Error reading theme at screen[%d] (%s): %v", i, screen.Name, err)
				return
			}
			dash.Screens[i].Theme = resolvedTheme
		}
		resolvedTheme, err := resolveTheme(allNamed, dash.ThemeRef)
		if err != nil {
			log.Printf("Error reading theme for dashboard %s: %v", name, err)
			return
		}
		dereffedTheme := Theme{}
		if resolvedTheme != nil {
			dereffedTheme = *resolvedTheme
		}
		dash.Theme = mergeTheme(defaultTheme, dereffedTheme)

		funcMap["once"] = makeOnceFunc()

		lang := "en"
		if cfg.Localization.Locale != "" {
			lang = strings.Split(cfg.Localization.Locale, "-")[0]
		}

		globals := map[string]any{
			"Animations": dash.Animations,
			"Lang":       lang,
		}
		funcMap["globals"] = makeGlobals(globals)

		funcMap["translate"] = makeTranslate(lang)
		funcMap["domainTranslations"] = makeDomainTranslations(lang)

		dashTmpl, err := tmpl.Clone()
		if err != nil {
			log.Printf("Could not clone template for %s: %v", name, err)
			continue
		}

		built[name] = builtDash{
			tmpl: dashTmpl.Funcs(funcMap),
			data: TemplateData{dash, cfg.Config, name, isDev},
		}
	}

	for name, b := range built {
		outputDir := frontendPath + "/static/" + name
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Printf("Could not create output dir for %s: %v", name, err)
			continue
		}
		out, err := os.Create(outputDir + "/index.html")
		if err != nil {
			log.Printf("Could not create output file for %s: %v", name, err)
			continue
		}
		if err := b.tmpl.ExecuteTemplate(out, "main.html.tmpl", b.data); err != nil {
			log.Printf("Could not execute template for %s: %v", name, err)
		}
		out.Sync()
		out.Close()
	}
}
