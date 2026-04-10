package main

import (
	"fmt"
	"strings"
)

var translations = map[string]map[string]map[string]string{
	"en": {
		"weather": {
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
		"media": {
			"sources": "Sources",
			"browse":  "Browse",
		},
	},
	"sv": {
		"weather": {
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
		"media": {
			"sources": "Källor",
			"browse":  "Bläddra",
		},
	},
	"de": {
		"weather": {
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
	},
}

func translate(key string, lang string) (string, error) {
	langMap, ok := translations[lang]
	if !ok {
		return "", fmt.Errorf("could not find language '%s'", lang)
	}
	keyParts := strings.Split(key, ":")

	if !(len(keyParts) == 2) {
		return "", fmt.Errorf("could not read key specifier '%s', make sure it is of format domain:key")
	}
	dKey := keyParts[0]
	tKey := keyParts[1]

	domain, ok := langMap[dKey]
	if !ok {
		return "", fmt.Errorf("could not find domain '%s' for language '%s'", dKey, lang)
	}

	t, ok := domain[tKey]
	if !ok {
		return "", fmt.Errorf("could not find specifier '%s' in domain '%s' for language '%s'", tKey, dKey, lang)
	}
	return t, nil
}

func domainTranslations(domain string, lang string) (map[string]string, error) {
	langMap, ok := translations[lang]
	if !ok {
		return nil, fmt.Errorf("could not find language '%s'", lang)
	}
	dTranslations, ok := langMap[domain]
	if !ok {
		return nil, fmt.Errorf("could not find domain '%s' for language '%s'", domain, lang)
	}

	return dTranslations, nil
}
