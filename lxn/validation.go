package lxn

import (
	"fmt"
)

type Validator interface {
	Warn(msg string)
}

func ValidateMessages(messages []Message, v Validator) {
	warnf := func(format string, args ...any) {
		if v != nil {
			v.Warn(fmt.Sprintf(format, args...))
		}
	}

	messageKeys := make(map[string]map[string]struct{}) // section => key set
	warnedDuplicates := make(map[string]struct{})       // (section, message key) set
	for _, msg := range messages {
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
