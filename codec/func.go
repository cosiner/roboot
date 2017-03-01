package codec

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"

	"github.com/cosiner/roboot"
)

type FuncEncoder struct {
	EncodeFunc func(io.Writer, interface{}) error
	Writer     io.Writer
}

func (f FuncEncoder) Encode(v interface{}) error {
	return f.EncodeFunc(f.Writer, v)
}

type FuncDecoder struct {
	DecodeFunc func(io.Reader, interface{}) error
	Reader     io.Reader
}

func (f FuncDecoder) Decode(v interface{}) error {
	return f.DecodeFunc(f.Reader, v)
}

type FuncCodec struct {
	ContentTyp     string
	MarshalFunc    func(interface{}) ([]byte, error)
	NewEncoderFunc func(io.Writer) roboot.Encoder
	EncodeFunc     func(io.Writer, interface{}) error
	UnmarshalFunc  func([]byte, interface{}) error
	NewDecoderFunc func(io.Reader) roboot.Decoder
	DecodeFunc     func(io.Reader, interface{}) error
}

var (
	ErrMarshalUnimplemented   = errors.New("marshal is not implemented")
	ErrUnmarshalUnimplemented = errors.New("unmarshal is not implemented")
)

func (f *FuncCodec) ContentType() string {
	return f.ContentTyp
}

func (f *FuncCodec) Encode(w io.Writer, v interface{}) error {
	if f.EncodeFunc != nil {
		return f.EncodeFunc(w, v)
	}
	if f.NewEncoderFunc != nil {
		return f.NewEncoderFunc(w).Encode(v)
	}
	if f.MarshalFunc != nil {
		bytes, err := f.MarshalFunc(v)
		if err != nil {
			return err
		}
		_, err = w.Write(bytes)
		return err
	}
	return ErrMarshalUnimplemented
}

func (f *FuncCodec) Marshal(v interface{}) ([]byte, error) {
	const defaultBufsize = 2048
	if f.MarshalFunc != nil {
		return f.MarshalFunc(v)
	}

	if f.EncodeFunc != nil || f.NewEncoderFunc != nil {
		buffer := bytes.NewBuffer(make([]byte, 0, defaultBufsize))
		err := f.Encode(buffer, v)
		return buffer.Bytes(), err
	}

	return nil, ErrMarshalUnimplemented
}

func (f *FuncCodec) NewEncoder(w io.Writer) roboot.Encoder {
	if f.NewEncoderFunc != nil {
		return f.NewEncoderFunc(w)
	}
	return FuncEncoder{
		EncodeFunc: f.Encode,
		Writer:     w,
	}
}

func (f *FuncCodec) Decode(r io.Reader, v interface{}) error {
	if f.DecodeFunc != nil {
		return f.DecodeFunc(r, v)
	}
	if f.NewDecoderFunc != nil {
		return f.NewDecoderFunc(r).Decode(v)
	}
	if f.UnmarshalFunc != nil {
		bytes, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		return f.UnmarshalFunc(bytes, v)
	}
	return ErrUnmarshalUnimplemented
}

func (f *FuncCodec) Unmarshal(b []byte, v interface{}) error {
	if f.UnmarshalFunc != nil {
		return f.UnmarshalFunc(b, v)
	}

	return f.Decode(bytes.NewReader(b), v)
}

func (f *FuncCodec) NewDecoder(r io.Reader) roboot.Decoder {
	if f.NewDecoderFunc != nil {
		return f.NewDecoderFunc(r)
	}
	return FuncDecoder{
		DecodeFunc: f.Decode,
		Reader:     r,
	}
}
