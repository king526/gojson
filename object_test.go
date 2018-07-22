package gojson

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Str   string
	Int   int64
	Uint  uint64
	Float float64
	Bool  bool
	Array []interface{}
	Map   map[string]interface{}
	Obj   *testStruct
	Nil   interface{}
}

var (
	s, _ = JsonIter.MarshalToString(&testStruct{
		Str:   "hello",
		Int:   math.MaxInt64,
		Uint:  math.MaxUint64,
		Float: 9.125e30,
		Bool:  true,
		Array: []interface{}{1, true, "world", nil},
		Obj:   &testStruct{Str: "simple"},
		Map: map[string]interface{}{
			"key1": 123,
			"key2": "val2",
			"key3": &testStruct{Str: "shadow"},
		},
	})
)

func Test_Read(t *testing.T) {
	obj := New(strings.NewReader(s))
	if obj.Err() != nil {
		t.Fatal(obj.Err)
	}
	fmt.Println(string(*obj.raw))
	it, _ := obj.Get("Str").Interface()
	assert.True(t, "hello" == it)
	str, _ := obj.Get("Str").String()
	assert.True(t, "hello" == str)
	boolV, _ := obj.Get("Bool").Bool()
	assert.True(t, boolV)
	int64V, _ := obj.Get("Int").Int64()
	assert.True(t, int64V == math.MaxInt64)
	uintV, _ := obj.Get("Uint").Uint64()
	assert.True(t, math.MaxUint64 == uintV)
	//test int overflow
	intV, err := obj.Get("Int").Int32()
	assert.True(t, err == ErrOverflow)
	assert.True(t, intV == -1)
	uint32V, err := obj.Get("Int").Uint32()
	assert.True(t, err == ErrOverflow)
	assert.True(t, uint32V == math.MaxUint32)

	floatV, _ := obj.Get("Float").Float()
	assert.Equal(t, 9.125e+30, floatV)
	//test type convert under rule
	floatV, _ = obj.Get("Uint").Float()
	assert.Equal(t, float64(math.MaxUint64), floatV)
	//test unmatch type read
	str, err = obj.Get("Float").String()
	assert.True(t, err != nil)
	//read array
	var arrays []interface{}
	obj.Get("Array").Read(&arrays)
	assert.True(t, len(arrays) == 4)
	assert.True(t, arrays[2] == "world")
	str, _ = obj.Get("Array").Index(2).String()
	assert.Equal(t, "world", str)
	//read object
	var elem *testStruct
	err = obj.Get("Obj").Read(&elem)
	assert.True(t, err == nil)
	assert.Equal(t, `simple`, elem.Str)
	var objAsMap map[string]interface{}
	err = obj.Get("Obj").Read(&objAsMap)
	assert.True(t, err == nil)
	//read raw bytes
	str, _ = obj.Get("Array").Data()
	assert.Equal(t, `[1,true,"world",null]`, str)

}

func Test_PathGet(t *testing.T) {
	obj := New(strings.NewReader(s))
	if obj.Err() != nil {
		t.Fatal(obj.Err())
	}
	str, err := obj.Get("Obj").Get("Str").String()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `simple`, str)
	//map
	sub := obj.Get("Map")
	if obj.Err() != nil {
		t.Fatal(sub.Err())
	}
	str, err = sub.Get("key2").String()
	assert.Equal(t, str, "val2")
	var elem *testStruct
	err = sub.Get("key3").Read(&elem)
	assert.True(t, err == nil)
	assert.Equal(t, `shadow`, elem.Str)
	//null
	sub = obj.Get("Nil").Get("2")
	assert.Equal(t, ErrDecode, sub.Err())
	//read other type as object
	_, err = obj.Get("Array").Get("Arr").String()
	assert.Equal(t, ErrDecode, err)
	_, err = obj.Get("Ztr").Get("Str").String()
	assert.Equal(t, ErrNotFound, err)
}
