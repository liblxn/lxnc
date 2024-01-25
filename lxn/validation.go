package lxn

import (
	"fmt"

	"github.com/liblxn/lxnc/schema"
)

type Validator struct {
	Warn func(msg string)
}

func Validate(c schema.Catalog, v Validator) {
	warnf := func(format string, args ...any) {
		if v.Warn != nil {
			v.Warn(fmt.Sprintf(format, args...))
		}
	}

	messageKeys := make(map[string]map[string]struct{}) // section => key set
	warnedDuplicates := make(map[string]struct{})       // (section, message key) set
	for _, msg := range c.Messages {
		keys, has := messageKeys[msg.Section]
		if !has {
			keys = make(map[string]struct{})
			messageKeys[msg.Section] = keys
		}

		if _, has = keys[msg.Key]; has {
			s := msg.Section + "." + msg.Key
			if _, warned := warnedDuplicates[s]; !warned {
				if msg.Section == "" {
					warnf("duplicate message key %q", msg.Key)
				} else {
					warnf("duplicate message key %q for section %q", msg.Key, msg.Section)
				}
				warnedDuplicates[s] = struct{}{}
			}
		}
		keys[msg.Key] = struct{}{}
	}
}
