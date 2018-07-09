package ptk

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io"
)

// M is a QoL shortcut for map[string]interface{}
type M map[string]interface{}

// ToJSON returns a string json representation of M, mostly for debugging.
func (m M) ToJSON(indent bool) string {
	if m == nil {
		return "{}"
	}
	j, _ := MarshalJSON(indent, m)
	return j
}

func MarshalJSON(indent bool, v interface{}) (string, error) {
	var (
		j   []byte
		err error
	)
	if indent {
		j, err = json.MarshalIndent(v, "", "\t")
	} else {
		j, err = json.Marshal(v)
	}
	return string(j), err
}

// PipeJSONObject uses an io.Pipe to stream an M rather than keeping it in memory.
// if the type of value is io.Reader, it will be directly streamed (so it must be valid JSON).
// if the type is func(io.Writer|*bufio.Writer) error, it will pass the underlying writer.
// otherwise json.Marshal will be called on the value.
func PipeJSONObject(obj M) io.Reader {
	pr, pw := io.Pipe()

	go func() {
		bw := bufio.NewWriter(pw)
		if err := WriteJSONObject(bw, obj); err != nil {
			pw.CloseWithError(err)
		} else {
			pw.CloseWithError(bw.Flush())
		}
		// pr.Close()

	}()
	return pr
}

func WriteJSONObject(w io.Writer, obj M) (err error) {
	var (
		b  []byte
		ln = len(obj)
		bw *bufio.Writer
	)

	if ww, ok := w.(*bufio.Writer); ok {
		bw = ww
	} else {
		bw = bufio.NewWriter(w)
		defer bw.Flush()
	}

	bw.WriteByte('{')
	for k, v := range obj {
		ln--

		bw.WriteByte('"')
		bw.WriteString(k)
		bw.WriteString(`":`)

		switch v := v.(type) {
		case io.Reader:
			if _, err = io.Copy(bw, v); err != nil {
				return
			}
		case M:
			if err = WriteJSONObject(bw, v); err != nil {
				return
			}
		case map[string]interface{}:
			if err = WriteJSONObject(bw, v); err != nil {
				return
			}
		case func(w io.Writer) error:
			if err = v(bw); err != nil {
				return
			}
		case func(w *bufio.Writer) error:
			if err = v(bw); err != nil {
				return
			}
		default:
			if b, err = json.Marshal(v); err != nil {
				return
			}
			bw.Write(b)
		}

		if ln > 0 {
			bw.WriteByte(',')
		}
	}
	bw.WriteByte('}')

	return
}

// Base64ToJSON allows converting a reader into a base64 encoded string and streaming it directly without using buffers.
// if enc is nil, it uses base64.StdEncoding by default.
// Example: WriteJSONObject(w, M{"image": Base64ToJSON(nil, "image/png", r)})
func Base64ToJSON(enc *base64.Encoding, mimeType string, r io.Reader) func(w io.Writer) error {
	if enc == nil {
		enc = base64.StdEncoding
	}

	return func(w io.Writer) error {
		var bw *bufio.Writer
		if ww, ok := w.(*bufio.Writer); ok {
			bw = ww
		} else {
			bw = bufio.NewWriter(w)
			defer bw.Flush()
		}

		bw.WriteString(`"data:` + mimeType + `;base64,`)
		enc := base64.NewEncoder(enc, bw)
		_, err := io.Copy(enc, r)
		enc.Close()
		bw.WriteByte('"')
		return err
	}
}
