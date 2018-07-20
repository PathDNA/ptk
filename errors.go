package ptk

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

func ErrorWithLine(err error) error {
	return Error{Msg: RuntimeLine(1), Err: err}
}

type Error struct {
	Msg string
	Err error
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}

type Errors []error

func (es *Errors) Push(err error) (pushed bool) {
	if pushed = err != nil; pushed {
		*es = append(*es, err)
	}
	return
}

func (es *Errors) PushWithLine(err error) (pushed bool) {
	if pushed = err != nil; pushed {
		*es = append(*es, Error{Msg: RuntimeLine(1), Err: err})
	}
	return
}

func (es *Errors) Wrap(msg string, err error) (pushed bool) {
	if pushed = err != nil; pushed {
		*es = append(*es, Error{Msg: msg, Err: err})
	}
	return
}

func (es *Errors) WrapWithLine(msg string, err error) (pushed bool) {
	if pushed = err != nil; pushed {
		*es = append(*es, Error{Msg: RuntimeLine(1), Err: Error{Msg: msg, Err: err}})
	}
	return
}

func (es Errors) Err() error {
	if len(es) > 0 {
		return es
	}

	return nil
}

func (es Errors) Error() string {
	msgs := make([]string, 0, len(es))
	for _, e := range es {
		msgs = append(msgs, e.Error())
	}

	return strings.Join(msgs, "\n")
}

func RuntimeLine(callerIdx int) string {
	_, file, line, ok := runtime.Caller(callerIdx + 1)
	if !ok {
		file = "???"
		line = 0
	}

	// make it output the package owning the file
	parts := strings.Split(file, "/")
	if len(parts) > 2 {
		parts = parts[len(parts)-2:]
	}

	return strings.Join(parts, "/") + ":" + strconv.Itoa(line)
}
