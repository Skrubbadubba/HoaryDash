package main

import (
	"encoding/json"
	"log"
	"net"
	"reflect"

	"dario.cat/mergo"
	"github.com/mitchellh/reflectwalk"
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

func enabledByDefault(v *bool) bool {
	if v == nil {
		return true
	}
	return *v
}

type NodeWalkerCallback[T any] func(f reflect.StructField, v *T) error

type nodeWalker[T any] struct {
	target reflect.Type
	cb     NodeWalkerCallback[T]
	err    error
}

func (w *nodeWalker[T]) Struct(v reflect.Value) error { return nil }

func (w *nodeWalker[T]) StructField(f reflect.StructField, v reflect.Value) error {
	if w.err != nil {
		return nil
	}
	if v.Type() != w.target {
		return nil
	}
	ptr := v.Addr().Interface().(*T)
	w.err = w.cb(f, ptr)
	return nil
}

func makeNodeWalker[T any]() func(target any, cb NodeWalkerCallback[T]) error {
	t := reflect.TypeOf((*T)(nil)).Elem()
	return func(target any, cb NodeWalkerCallback[T]) error {
		w := &nodeWalker[T]{target: t, cb: cb}
		if err := reflectwalk.Walk(target, w); err != nil {
			return err
		}
		return w.err
	}
}

func ipOnly(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
