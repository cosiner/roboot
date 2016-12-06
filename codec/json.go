package codec

import (
	"encoding/json"

	"io"

	"github.com/cosiner/roboot"
)

var JSON roboot.Codec = &FuncCodec{
	MarshalFunc: json.Marshal,
	NewEncoderFunc: func(w io.Writer) roboot.Encoder {
		return json.NewEncoder(w)
	},
	UnmarshalFunc: json.Unmarshal,
	NewDecoderFunc: func(r io.Reader) roboot.Decoder {
		return json.NewDecoder(r)
	},
}
