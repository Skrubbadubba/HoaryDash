package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/gorilla/websocket"
)

//go:embed mdi.json
var mdiData []byte

type ComponentIconMap map[string]map[string]ClassIcons

type ClassIcons struct {
	Default string                 `json:"default"`
	State   map[string]string      `json:"state"`
	Range   map[json.Number]string `json:"range"`
}

type getIconsResponse struct {
	Result struct {
		Resources map[string]map[string]ClassIcons `json:"resources"`
	} `json:"result"`
}

type ClassIconsSVG struct {
	Default template.HTML
	State   map[string]template.HTML
}

type ComponentIconMapSVG map[string]map[string]ClassIconsSVG

func fetchComponentIcons(ha HAConfig) (ComponentIconMap, error) {
	conn, _, err := websocket.DefaultDialer.Dial(ha.WSURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fetchComponentIcons: dial: %w", err)
	}
	defer conn.Close()

	if err := haAuth(conn, ha.Token); err != nil {
		return nil, fmt.Errorf("fetchComponentIcons: auth: %w", err)
	}

	req := map[string]any{
		"id":       1,
		"type":     "frontend/get_icons",
		"category": "entity_component",
	}
	if err := conn.WriteJSON(req); err != nil {
		return nil, fmt.Errorf("fetchComponentIcons: send request: %w", err)
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("fetchComponentIcons: read response: %w", err)
	}

	var resp getIconsResponse
	if err := json.Unmarshal(msg, &resp); err != nil {
		return nil, fmt.Errorf("fetchComponentIcons: unmarshal: %w", err)
	}

	return ComponentIconMap(resp.Result.Resources), nil
}

func filterIconMap(full ComponentIconMap, present DomainClassSet) ComponentIconMap {
	filtered := ComponentIconMap{}
	for dom, classes := range present {
		domainIcons, ok := full[dom]
		if !ok {
			continue
		}
		filtered[dom] = map[string]ClassIcons{}
		for class := range classes {
			if icons, ok := domainIcons[class]; ok {
				filtered[dom][class] = icons
			}
		}
		if icons, ok := domainIcons["_"]; ok {
			filtered[dom]["_"] = icons
		}
	}
	return filtered
}

func resolveIconMapSVG(icons ComponentIconMap, mdiIcons map[string]string) ComponentIconMapSVG {
	result := ComponentIconMapSVG{}
	for dom, classes := range icons {
		result[dom] = map[string]ClassIconsSVG{}
		for class, ci := range classes {
			svg := ClassIconsSVG{
				Default: mdiToSVG(ci.Default, mdiIcons),
			}
			if ci.State != nil {
				svg.State = make(map[string]template.HTML, len(ci.State))
				for state, iconName := range ci.State {
					svg.State[state] = mdiToSVG(iconName, mdiIcons)
				}
			}
			result[dom][class] = svg
		}
	}
	return result
}

func mdiToSVG(mdiName string, mdiIcons map[string]string) template.HTML {
	name := strings.TrimPrefix(mdiName, "mdi:")
	return iconToSVG(name, &mdiIcons)
}

func iconToSVG(name string, icons *map[string]string) template.HTML {
	iconMap := *icons
	path, ok := iconMap[name]
	if !ok {
		return ""
	}
	return template.HTML(fmt.Sprintf(
		`<svg class="icon" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="%s"/></svg>`,
		path,
	))
}

func loadMdiIcons() (map[string]string, error) {
	var mdiIcons map[string]string
	if err := json.Unmarshal(mdiData, &mdiIcons); err != nil {
		return nil, fmt.Errorf("error unmarshaling embedded mdi icons into json: %w", err)
	}
	return mdiIcons, nil
}
