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
		log.Printf(message+": %v", e)
		return
	}
}

type DomainClassSet map[string]map[string]struct{}

func (d DomainClassSet) add(domain, deviceClass string) {
	if d[domain] == nil {
		d[domain] = map[string]struct{}{}
	}
	d[domain][deviceClass] = struct{}{}
}

func (d DomainClassSet) has(domain, deviceClass string) bool {
	classes, ok := d[domain]
	if !ok {
		return false
	}
	_, ok = classes[deviceClass]
	return ok
}
