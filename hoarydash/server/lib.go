package main

import (
	"encoding/json"
	"log"

	"dario.cat/mergo"
)

func clonePtr[T any](ptr *T) *T {
	if ptr == nil {
		return nil
	}
	b := *ptr
	return &b
}

func mergeOverride[T any, U any](base T, override U) T {
	mergo.Merge(&base, override, mergo.WithOverride)
	return base
}

func nilfunc() any { return nil }

func jsonStr(j interface{}) string { // For debugging
	var out []byte
	out, err := json.Marshal(j)
	if err != nil {
		return ""
	}
	return string(out)
}

func check(e error, message string, v ...any) {
	if e != nil {
		log.Print(e)
		return
	}
	if isDev {
		log.Printf(message, v...)
	}
}
