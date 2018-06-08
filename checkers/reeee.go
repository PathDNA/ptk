package checkers

import (
	"regexp"
)

var (
	emails = regexp.MustCompile(`^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`)
)

// Email returns a copy
func Email() *regexp.Regexp { return emails.Copy() }

// MatchEmail matches against the main regexp, not recommended for high concurrency.
func MatchEmail(email string) bool { return emails.MatchString(email) }
