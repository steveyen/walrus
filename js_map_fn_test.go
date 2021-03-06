//  Copyright (c) 2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package walrus

import (
	"github.com/sdegutis/go.assert"
	"testing"
)

// Just verify that the calls to the emit() fn show up in the output.
func TestEmitFunction(t *testing.T) {
	mapper, err := NewJSMapFunction(`function(doc) {emit("key", "value"); emit("k2","v2")}`)
	assertNoError(t, err, "Couldn't create mapper")
	rows, err := mapper.callMapper(`{}`, "doc1")
	assertNoError(t, err, "callMapper failed")
	assert.Equals(t, len(rows), 2)
	assert.DeepEquals(t, rows[0], ViewRow{Key: "key", Value: "value"})
	assert.DeepEquals(t, rows[1], ViewRow{Key: "k2", Value: "v2"})
	// (callMapper doesn't set the ID field.)
}

func testMap(t *testing.T, mapFn string, doc string) []ViewRow {
	mapper, err := NewJSMapFunction(mapFn)
	assertNoError(t, err, "Couldn't create mapper")
	rows, err := mapper.callMapper(doc, "doc1")
	assertNoError(t, err, "callMapper failed")
	return rows
}

// Now just make sure the input comes through intact
func TestInputParse(t *testing.T) {
	rows := testMap(t, `function(doc) {emit(doc.key, doc.value);}`,
		`{"key": "k", "value": "v"}`)
	assert.Equals(t, len(rows), 1)
	assert.DeepEquals(t, rows[0], ViewRow{Key: "k", Value: "v"})
}

// Test different types of keys/values:
func TestKeyTypes(t *testing.T) {
	rows := testMap(t, `function(doc) {emit(doc.key, doc.value);}`,
		`{"key": true, "value": false}`)
	assert.DeepEquals(t, rows[0], ViewRow{Key: true, Value: false})
	rows = testMap(t, `function(doc) {emit(doc.key, doc.value);}`,
		`{"key": null, "value": 0}`)
	assert.DeepEquals(t, rows[0], ViewRow{Key: nil, Value: float64(0)})
	rows = testMap(t, `function(doc) {emit(doc.key, doc.value);}`,
		`{"key": ["foo", 23, []], "value": [null]}`)
	assert.DeepEquals(t, rows[0],
		ViewRow{Key: []interface{}{"foo", 23.0, []interface{}{}},
			Value: []interface{}{nil}})
}

// Empty/no-op map fn
func TestEmptyJSMapFunction(t *testing.T) {
	mapper, err := NewJSMapFunction(`function(doc) {}`)
	assertNoError(t, err, "Couldn't create mapper")
	rows, err := mapper.callMapper(`{"key": "k", "value": "v"}`, "doc1")
	assertNoError(t, err, "callMapper failed")
	assert.Equals(t, len(rows), 0)
}

// Test the public API
func TestPublicJSMapFunction(t *testing.T) {
	mapper, err := NewJSMapFunction(`function(doc) {emit(doc.key, doc.value);}`)
	assertNoError(t, err, "Couldn't create mapper")
	rows, err := mapper.CallFunction(`{"key": "k", "value": "v"}`, "doc1")
	assertNoError(t, err, "CallFunction failed")
	assert.Equals(t, len(rows), 1)
	assert.DeepEquals(t, rows[0], ViewRow{ID: "doc1", Key: "k", Value: "v"})
	mapper.Stop()
}
