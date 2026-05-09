package gobx

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"sync"
)

// Codec is used to encode/decode a type using gob. The type header metadata
// usually included with a stream of gob objects are not included in the
// output. Since the type paramater restricts which type is decoded, the header
// isn't necessary and is useful when working with single objects.
type Codec[T any] struct {
	dec    *gob.Decoder
	enc    *gob.Encoder
	wb, rb bytes.Buffer

	once sync.Once
}

func (e *Codec[T]) init() error {
	if rt := reflect.TypeFor[T](); rt.Kind() == reflect.Pointer {
		return fmt.Errorf("must provide non-pointer type, received: %v", rt)
	}
	e.enc = gob.NewEncoder(&e.wb)
	var zero T
	if err := e.enc.Encode(zero); err != nil {
		return err
	}
	// write the type metadata for use by the decoder
	e.rb.Write(e.wb.Bytes())
	e.wb.Reset()
	e.dec = gob.NewDecoder(&e.rb)
	if err := e.dec.Decode(&zero); err != nil {
		return err
	}
	return nil
}

func (e *Codec[T]) Encode(v T) ([]byte, error) {
	var err error
	e.once.Do(func() {
		err = e.init()
	})
	if err != nil {
		return nil, err
	}
	if err := e.enc.Encode(v); err != nil {
		return nil, err
	}
	data := e.wb.Bytes()
	e.wb.Reset()
	return data, nil
}

func (e *Codec[T]) Decode(data []byte) (T, error) {
	var err error
	e.once.Do(func() {
		err = e.init()
	})
	if err != nil {
		return *new(T), err
	}
	e.rb.Reset()
	e.rb.Write(data)
	var v T
	if err := e.dec.Decode(&v); err != nil {
		return v, err
	}
	return v, nil
}
