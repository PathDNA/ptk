package ptk

import (
	"bufio"
	"encoding/json"
	"io"

	"github.com/missionMeteora/apiserv"
)

type M = apiserv.M

// PipeJSONObject uses an io.Pipe to stream an M rather than keeping it in memory.
// if the type of value is io.Reader, it will be directly streamed (so it must be valid JSON).
// otherwise json.Marshal will be called on the value.
func PipeJSONObject(obj M) io.Reader {
	pr, pw := io.Pipe()

	go func() {
		bw := bufio.NewWriter(pw)
		if err := pipeJSONObject(bw, obj); err != nil {
			pw.CloseWithError(err)
		} else {
			pw.CloseWithError(bw.Flush())
		}
		pr.Close()

	}()
	return pr
}

func pipeJSONObject(bw *bufio.Writer, obj M) (err error) {
	var (
		b  []byte
		ln = len(obj)
	)
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
			if err = pipeJSONObject(bw, v); err != nil {
				return
			}
		case map[string]interface{}:
			if err = pipeJSONObject(bw, v); err != nil {
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
