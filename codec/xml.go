package codec

import (
	"encoding/xml"
	"io"

	"github.com/cosiner/roboot"
)

var XML roboot.Codec = &FuncCodec{
	ContentTyp:  "application/xml",
	MarshalFunc: xml.Marshal,
	NewEncoderFunc: func(w io.Writer) roboot.Encoder {
		return xml.NewEncoder(w)
	},
	UnmarshalFunc: xml.Unmarshal,
	NewDecoderFunc: func(r io.Reader) roboot.Decoder {
		return xml.NewDecoder(r)
	},
}
