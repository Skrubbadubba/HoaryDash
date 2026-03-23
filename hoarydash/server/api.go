package main

import (
	"encoding/json"
	"net/http"
)

var weatherTranslations = map[string]map[string]string{
	"en": {
		"clear-night":     "Clear night",
		"cloudy":          "Cloudy",
		"exceptional":     "Exceptional",
		"fog":             "Foggy",
		"hail":            "Hail",
		"lightning":       "Lightning",
		"lightning-rainy": "Lightning, rainy",
		"partlycloudy":    "Partly cloudy",
		"pouring":         "Pouring",
		"rainy":           "Rainy",
		"snowy":           "Snowy",
		"snowy-rainy":     "Snowy, rainy",
		"sunny":           "Sunny",
		"windy":           "Windy",
		"windy-variant":   "Windy",
	},
	"sv": {
		"clear-night":     "Klart, natt",
		"cloudy":          "Molnigt",
		"exceptional":     "Exceptionellt",
		"fog":             "Dimma",
		"hail":            "Hagel",
		"lightning":       "Åska",
		"lightning-rainy": "Åska, regnigt",
		"partlycloudy":    "Delvis molnigt",
		"pouring":         "Ösregn",
		"rainy":           "Regnigt",
		"snowy":           "Snöigt",
		"snowy-rainy":     "Snöigt, regnigt",
		"sunny":           "Soligt",
		"windy":           "Blåsigt",
		"windy-variant":   "Blåsigt",
	},
	"de": {
		"clear-night":     "Klare Nacht",
		"cloudy":          "Bewölkt",
		"exceptional":     "Außergewöhnlich",
		"fog":             "Neblig",
		"hail":            "Hagel",
		"lightning":       "Gewitter",
		"lightning-rainy": "Gewitter, regnerisch",
		"partlycloudy":    "Teilweise bewölkt",
		"pouring":         "Starkregen",
		"rainy":           "Regnerisch",
		"snowy":           "Schneefall",
		"snowy-rainy":     "Schneeregen",
		"sunny":           "Sonnig",
		"windy":           "Windig",
		"windy-variant":   "Windig, bewölkt",
	},
}

var widgetTranslations = map[string]map[string]map[string]string{
	"weather": weatherTranslations,
	// "media_player": mediaPlayerTranslations,
}

func translationsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		widget := r.PathValue("widget")
		lang := r.PathValue("lang")

		widgetMap, ok := widgetTranslations[widget]
		if !ok {
			http.Error(w, "unknown widget", http.StatusNotFound)
			return
		}

		translations, ok := widgetMap[lang]
		if !ok {
			translations = widgetMap["en"] // fallback to english
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "max-age=86400")
		json.NewEncoder(w).Encode(translations)
	}
}
