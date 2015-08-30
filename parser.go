// Copyright 2015 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"sort"
	"strings"
)

const (
	maxLevel = 255
)

type parser struct {
	fields map[uint8]records
	static map[string]Handle
}

type record struct {
	key    uint16
	handle Handle
	parts  []string
}

type records []*record

func (n records) Len() int           { return len(n) }
func (n records) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n records) Less(i, j int) bool { return n[i].key < n[j].key }

func newParser() *parser {
	return &parser{
		fields: make(map[uint8]records),
		static: make(map[string]Handle),
	}
}

func (p *parser) register(path string, handle Handle) bool {
	if parts, ok := split(path); ok {
		var static, dynamic uint16
		for _, value := range parts {
			if strings.HasPrefix(value, ":") {
				dynamic++
			} else {
				static++
			}
		}
		if dynamic == 0 {
			p.static["/"+strings.Join(parts, "/")] = handle
		} else {
			level := uint8(len(parts))
			p.fields[level] = append(p.fields[level], &record{key: dynamic<<8 + static, handle: handle, parts: parts})
			sort.Sort(records(p.fields[level]))
		}
		return true
	}

	return false
}

func (p *parser) get(path string) (handle Handle, result []Param, ok bool) {
	if handle, ok := p.static[path]; ok {
		return handle, nil, true
	}
	if parts, ok := split(path); ok {
		if handle, ok := p.static["/"+strings.Join(parts, "/")]; ok {
			return handle, nil, true
		}
		if data := p.fields[uint8(len(parts))]; data != nil {
			for _, nds := range data {
				values := nds.parts
				result = nil
				found := true
				for idx, value := range values {
					if value != parts[idx] && !(strings.HasPrefix(value, ":")) {
						found = false
						break
					} else {
						if strings.HasPrefix(value, ":") {
							result = append(result, Param{Key: value, Value: parts[idx]})
						}
					}
				}
				if found {
					return nds.handle, result, true
				}
			}
		}
	}

	return nil, nil, false
}

func split(path string) (result []string, ok bool) {
	sdata := strings.Split(strings.Trim(path, "/"), "/")
	if len(sdata) != 0 && len(sdata) < maxLevel {
		for _, value := range sdata {
			if v := strings.Trim(value, " "); v == "" {
				continue
			} else {
				result = append(result, v)
			}
		}
		if len(result) != 0 {
			ok = true
		}
	}

	return
}
