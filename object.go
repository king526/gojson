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

const (
	typeNotArray = 1
	typeNotMap   = 2
	MaxInt       = int64(^uint(0) >> 1)
)

var (
	//ConfigCompatibleWithStandardLibrary+UseNumber

	JsonIter = jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
		UseNumber:              true,
	}.Froze()
	bFalse = []byte{'f', 'a', 'l', 's', 'e'}
	bTrue  = []byte{'t', 'r', 'u', 'e'}

	ErrNotFound = errors.New("not found")
	ErrDecode   = errors.New("wrong decode")
	ErrOverflow = errors.New("overflow")
)

type Json struct {
	raw *jsoniter.RawMessage
	kvs map[string]*jsoniter.RawMessage
	arr []*jsoniter.RawMessage
	typ uint8
	err error
}

func New(r io.Reader) (j *Json) {
	j = &Json{}
	if r == nil {
		return
	}
	j.err = JsonIter.NewDecoder(r).Decode(&j.raw)
	return
}

func (j *Json) Err() error {
	return j.err
}

//Data return the raw data as string
func (j *Json) Data() (ret string, err error) {
	if j.err != nil {
		err = j.err
		return
	}
	if j.raw != nil {
		ret = string(*j.raw)
	}
	return
}

func (j *Json) IsNull() (ok bool, err error) {
	if j.err != nil {
		err = j.err
		return
	}
	ok = j.raw == nil && j.arr == nil && j.kvs == nil
	return
}

//Get return Json of the key,return ErrWrongType if caller not a JsonObject
func (j *Json) Get(key string) *Json {
	if j.err != nil {
		return j
	}
	if j.kvs != nil {
		return j.get(key)
	}
	if j.typ&typeNotMap != 0 {
		return &Json{err: ErrDecode}
	}
	if j.raw == nil {
		j.typ |= typeNotMap & typeNotArray
		return &Json{err: ErrDecode}
	}
	err := JsonIter.NewDecoder(bytes.NewReader(*j.raw)).Decode(&j.kvs)
	if err != nil {
		j.typ |= typeNotMap
		return &Json{err: ErrDecode}
	} else {
		j.typ |= typeNotArray
	}
	return j.get(key)
}

//Get return Json of the key,return ErrWrongType if caller not a JsonArray
func (j *Json) Index(idx int) *Json {
	if j.err != nil {
		return j
	}
	if j.arr != nil {
		return j.index(idx)
	}
	if j.typ&typeNotArray != 0 {
		return &Json{err: ErrDecode}
	}
	if j.raw == nil {
		j.typ |= typeNotMap & typeNotArray
		return &Json{err: ErrDecode}
	}
	err := JsonIter.NewDecoder(bytes.NewReader(*j.raw)).Decode(&j.arr)
	if err != nil {
		j.typ |= typeNotArray
		return &Json{err: ErrDecode}
	} else {
		j.typ |= typeNotMap
	}
	return j.index(idx)
}

func (j *Json) get(key string) *Json {
	raw, ok := j.kvs[key]
	if !ok {
		return &Json{err: ErrNotFound}
	}
	return &Json{raw: raw}
}

func (j *Json) index(idx int) *Json {
	if len(j.arr)+1 < idx {
		return &Json{err: ErrNotFound}
	}
	return &Json{raw: j.arr[idx]}
}

func (j *Json) Read(obj interface{}) error {
	if j.err != nil || j.raw == nil {
		return j.err
	}
	return JsonIter.Unmarshal(*j.raw, obj)
}

//Interface unmarshal to an interface{} ,same as Read
func (j *Json) Interface() (it interface{}, err error) {
	if j.err != nil || j.raw == nil {
		return nil, j.err
	}
	err = JsonIter.Unmarshal(*j.raw, &it)
	return
}

//String convert the value to string
func (j *Json) String() (ret string, err error) {
	if j.err != nil {
		err = j.err
		return
	}
	if j.raw == nil {
		return
	}
	if (*j.raw)[0] == '"' {
		err = JsonIter.Unmarshal(*j.raw, &ret)
	} else {
		err = fmt.Errorf("error parse string:%s", *j.raw)
	}
	return
}

func (j *Json) Bool() (ret bool, err error) {
	if j.err != nil {
		err = j.err
		return
	}
	if j.raw == nil {
		return
	}
	if bytes.Equal(*j.raw, bTrue) {
		ret = true
	} else if bytes.Equal(*j.raw, bFalse) {
		ret = false
	} else {
		err = fmt.Errorf("error parse bool:%s", *j.raw)
	}
	return
}

//Int convert the number to int,return ErrOverflow when overflow
func (j *Json) Int() (int, error) {
	ret, err := j.Int64()
	if err == nil && ret > MaxInt {
		err = ErrOverflow
	}
	return int(ret), err
}

func (j *Json) Int64() (ret int64, err error) {
	if j.err != nil {
		err = j.err
		return
	}
	if j.raw != nil {
		ret, err = strconv.ParseInt(string(*j.raw), 10, 64)
	}
	return
}

func (j *Json) Uint64() (ret uint64, err error) {
	if j.err != nil {
		err = j.err
		return
	}
	if j.raw != nil {
		ret, err = strconv.ParseUint(string(*j.raw), 10, 64)
	}
	return
}

func (j *Json) Int32() (int32, error) {
	ret, err := j.Int64()
	if err == nil && ret > math.MaxInt32 {
		err = ErrOverflow
	}
	return int32(ret), err
}

func (j *Json) Uint32() (uint32, error) {
	ret, err := j.Uint64()
	if err == nil && ret > math.MaxUint32 {
		err = ErrOverflow
	}
	return uint32(ret), err
}

func (j *Json) Float() (ret float64, err error) {
	if j.err != nil {
		err = j.err
		return
	}
	if j.raw != nil {
		ret, err = strconv.ParseFloat(string(*j.raw), 64)
	}
	return
}
