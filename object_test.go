package gojson

import (
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
	obj, err := NewObject(strings.NewReader(s))
	if err != nil {
		t.Fail()
	}
	str, _ := obj.ReadString("Str")
	assert.True(t, "hello" == str)
	boolV, _ := obj.ReadBool("Bool")
	assert.True(t, boolV)
	int64V, _ := obj.ReadInt64("Int")
	assert.True(t, int64V == math.MaxInt64)
	uintV, _ := obj.ReadUint64("Uint")
	assert.True(t, math.MaxUint64 == uintV)
	//test int overflow
	intV, err := obj.ReadInt32("Int")
	assert.True(t, err == ErrOverflow)
	assert.True(t, intV == -1)
	uint32V, err := obj.ReadUint32("Int")
	assert.True(t, err == ErrOverflow)
	assert.True(t, uint32V == math.MaxUint32)

	floatV, _ := obj.ReadFloat("Float")
	assert.Equal(t, 9.125e+30, floatV)
	//test type convert under rule
	floatV, _ = obj.ReadFloat("Uint")
	assert.Equal(t, float64(math.MaxUint64), floatV)
	//test unmatch type read
	str, err = obj.ReadString("Float")
	assert.True(t, err != nil)
	//read array
	var arrays []interface{}
	obj.Read("Array", &arrays)
	assert.True(t, len(arrays) == 4)
	assert.True(t, arrays[2] == "world")
	//read object
	var elem *testStruct
	err = obj.Read("Obj", &elem)
	assert.True(t, err == nil)
	assert.Equal(t, `simple`, elem.Str)
	var objAsMap map[string]interface{}
	err = obj.Read("Obj", &objAsMap)
	assert.True(t, err == nil)
	//read raw bytes
	str, _ = obj.ReadRawValueAsString("Array")
	assert.Equal(t, `[1,true,"world",null]`, str)
}

func Test_ReadAsObject(t *testing.T) {
	obj, err := NewObject(strings.NewReader(s))
	if err != nil {
		t.Fail()
	}
	sub, err := obj.GetAsObject("Obj")
	if err != nil {
		t.Fail()
	}
	str, _ := sub.ReadString("Str")
	assert.Equal(t, `simple`, str)
	//map
	sub, err = obj.GetAsObject("Map")
	if err != nil {
		t.Fail()
	}
	var elem *testStruct
	err = sub.Read("key3", &elem)
	assert.True(t, err == nil)
	assert.Equal(t, `shadow`, elem.Str)
	//null
	sub, err = obj.GetAsObject("Nil")
	if err != nil {
		t.Fail()
	}
	assert.True(t, 0 == len(sub.kvs))
	//read other type as object
	_, err = obj.GetAsObject("Array")
	assert.True(t, err != nil)
	_, err = obj.GetAsObject("Str")
	assert.True(t, err != nil)
}
