package gojson

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/json-iterator/go"
)

var (
	//ConfigCompatibleWithStandardLibrary+UseNumber

	MaxInt = int64(^uint(0) >> 1)

	JsonIter = jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
		UseNumber:              true,
	}.Froze()
	bFalse = []byte{'f', 'a', 'l', 's', 'e'}
	bTrue  = []byte{'t', 'r', 'u', 'e'}

	ErrEmpty    = errors.New("empty")
	ErrNotFound = errors.New("not found")
	ErrOverflow = errors.New("overflow")
)

type JsonObject struct {
	kvs map[string]*jsoniter.RawMessage
}

func NewObject(r io.Reader) (j *JsonObject, e error) {
	j = &JsonObject{}
	if r == nil {
		return
	}
	e = JsonIter.NewDecoder(r).Decode(&j.kvs)
	return
}

func (j *JsonObject) Read(key string, obj interface{}) error {
	raw, ok := j.kvs[key]
	if !ok {
		return ErrNotFound
	}
	if raw != nil {
		return JsonIter.Unmarshal(*raw, obj)
	}
	return nil
}

func (j *JsonObject) ReadString(key string) (ret string, err error) {
	raw, ok := j.kvs[key]
	if !ok {
		err = ErrNotFound
		return
	}
	if raw == nil {
		return
	}
	if (*raw)[0] == '"' {
		JsonIter.Unmarshal(*raw, &ret)
	} else {
		err = fmt.Errorf("error parse string:%s", *raw)
	}
	return
}

func (j *JsonObject) ReadBool(key string) (ret bool, err error) {
	raw, ok := j.kvs[key]
	if !ok {
		err = ErrNotFound
		return
	}
	if raw == nil {
		return
	}
	if bytes.Equal(*raw, bTrue) {
		ret = true
	} else if bytes.Equal(*raw, bFalse) {
		ret = false
	} else {
		err = fmt.Errorf("error parse bool:%s", *raw)
	}
	return
}

//ReadInt convert the number to int,return ErrOverflow when overflow
func (j *JsonObject) ReadInt(key string) (int, error) {
	ret, err := j.ReadInt64(key)
	if err == nil && ret > MaxInt {
		err = ErrOverflow
	}
	return int(ret), err
}

func (j *JsonObject) ReadInt64(key string) (ret int64, err error) {
	raw, ok := j.kvs[key]
	if !ok {
		err = ErrNotFound
		return
	}
	if raw != nil {
		ret, err = strconv.ParseInt(string(*raw), 10, 64)
	}
	return
}

func (j *JsonObject) ReadUint64(key string) (ret uint64, err error) {
	raw, ok := j.kvs[key]
	if !ok {
		err = ErrNotFound
		return
	}
	if raw != nil {
		ret, err = strconv.ParseUint(string(*raw), 10, 64)
	}
	return
}

func (j *JsonObject) ReadInt32(key string) (int32, error) {
	ret, err := j.ReadInt64(key)
	if err == nil && ret > math.MaxInt32 {
		err = ErrOverflow
	}
	return int32(ret), err
}

func (j *JsonObject) ReadUint32(key string) (uint32, error) {
	ret, err := j.ReadUint64(key)
	if err == nil && ret > math.MaxUint32 {
		err = ErrOverflow
	}
	return uint32(ret), err
}

func (j *JsonObject) ReadFloat(key string) (ret float64, err error) {
	raw, ok := j.kvs[key]
	if !ok {
		err = ErrNotFound
		return
	}
	if raw != nil {
		ret, err = strconv.ParseFloat(string(*raw), 64)
	}
	return
}

func (j *JsonObject) ReadRawValueAsString(key string) (ret string, err error) {
	raw, ok := j.kvs[key]
	if !ok {
		err = ErrNotFound
		return
	}
	if raw != nil {
		ret = string(*raw)
	}
	return
}

//GetAsObject read the key to JsonObject,return empty object if the value is null
//this library is ame at dynimac marshal map to final entity struct,so path chain get value is not considered,
//because each time call GetAsObject we will serilize the data as no memory.
func (j *JsonObject) GetAsObject(key string) (obj *JsonObject, err error) {
	raw, ok := j.kvs[key]
	if !ok {
		err = ErrNotFound
		return
	}
	if raw == nil {
		obj = &JsonObject{kvs: map[string]*jsoniter.RawMessage{}}
		return
	}
	return NewObject(bytes.NewReader(*raw))
}
