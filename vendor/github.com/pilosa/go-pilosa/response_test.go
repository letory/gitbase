// Copyright 2017 Pilosa Corp.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
// 1. Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright
// notice, this list of conditions and the following disclaimer in the
// documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its
// contributors may be used to endorse or promote products derived
// from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
// CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
// INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
// CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
// BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY,
// WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
// NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH
// DAMAGE.

package pilosa

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"testing"

	pbuf "github.com/pilosa/go-pilosa/gopilosa_pbuf"
)

func TestNewBitmapResultFromInternal(t *testing.T) {
	targetAttrs := map[string]interface{}{
		"name":       "some string",
		"age":        int64(95),
		"registered": true,
		"height":     1.83,
	}
	targetBits := []uint64{5, 10}
	attrs := []*pbuf.Attr{
		{Key: "name", StringValue: "some string", Type: 1},
		{Key: "age", IntValue: 95, Type: 2},
		{Key: "registered", BoolValue: true, Type: 3},
		{Key: "height", FloatValue: 1.83, Type: 4},
	}
	bitmap := &pbuf.Bitmap{
		Attrs: attrs,
		Bits:  []uint64{5, 10},
	}
	result, err := newBitmapResultFromInternal(bitmap)
	if err != nil {
		t.Fatalf("Failed with error: %s", err)
	}
	// assertMapEquals(t, targetAttrs, result.Attributes)
	if !reflect.DeepEqual(targetAttrs, result.Attributes) {
		t.Fatal()
	}
	if !reflect.DeepEqual(targetBits, result.Bits) {
		t.Fatal()
	}
}

func TestNewQueryResponseFromInternal(t *testing.T) {
	targetAttrs := map[string]interface{}{
		"name":       "some string",
		"age":        int64(95),
		"registered": true,
		"height":     1.83,
	}
	targetBits := []uint64{5, 10}
	targetCountItems := []CountResultItem{
		{ID: 10, Count: 100},
	}
	attrs := []*pbuf.Attr{
		{Key: "name", StringValue: "some string", Type: 1},
		{Key: "age", IntValue: 95, Type: 2},
		{Key: "registered", BoolValue: true, Type: 3},
		{Key: "height", FloatValue: 1.83, Type: 4},
	}
	bitmap := &pbuf.Bitmap{
		Attrs: attrs,
		Bits:  []uint64{5, 10},
	}
	pairs := []*pbuf.Pair{
		{ID: 10, Count: 100},
	}
	response := &pbuf.QueryResponse{
		Results: []*pbuf.QueryResult{
			{Type: QueryResultTypeBitmap, Bitmap: bitmap},
			{Type: QueryResultTypePairs, Pairs: pairs},
		},
		Err: "",
	}
	qr, err := newQueryResponseFromInternal(response)
	if err != nil {
		t.Fatalf("Failed with error: %s", err)
	}
	if qr.ErrorMessage != "" {
		t.Fatalf("ErrorMessage should be empty")
	}
	if !qr.Success {
		t.Fatalf("IsSuccess should be true")
	}

	results := qr.Results()
	if len(results) != 2 {
		t.Fatalf("Number of results should be 2")
	}
	if results[0] != qr.Result() {
		t.Fatalf("Result() should return the first result")
	}
	if !reflect.DeepEqual(targetAttrs, results[0].Bitmap().Attributes) {
		t.Fatalf("The bitmap result should contain the attributes")
	}
	if !reflect.DeepEqual(targetBits, results[0].Bitmap().Bits) {
		t.Fatalf("The bitmap result should contain the bits")
	}
	if !reflect.DeepEqual(targetCountItems, results[1].CountItems()) {
		t.Fatalf("The response should include count items")
	}
}

func TestNewQueryResponseWithErrorFromInternal(t *testing.T) {
	response := &pbuf.QueryResponse{
		Err: "some error",
	}
	qr, err := newQueryResponseFromInternal(response)
	if err != nil {
		t.Fatalf("Failed with error: %s", err)
	}
	if qr.ErrorMessage != "some error" {
		t.Fatalf("The response should include the error message")
	}
	if qr.Success {
		t.Fatalf("IsSuccess should be false")
	}
	if qr.Result() != nil {
		t.Fatalf("If there are no results, Result should return nil")
	}
}

func TestNewQueryResponseFromInternalFailure(t *testing.T) {
	attrs := []*pbuf.Attr{
		{Key: "name", StringValue: "some string", Type: 99},
	}
	bitmap := &pbuf.Bitmap{
		Attrs: attrs,
	}
	response := &pbuf.QueryResponse{
		Results: []*pbuf.QueryResult{{Type: QueryResultTypeBitmap, Bitmap: bitmap}},
	}
	qr, err := newQueryResponseFromInternal(response)
	if qr != nil && err == nil {
		t.Fatalf("Should have failed")
	}
	response = &pbuf.QueryResponse{
		ColumnAttrSets: []*pbuf.ColumnAttrSet{{ID: 1, Attrs: attrs}},
	}
	qr, err = newQueryResponseFromInternal(response)
	if qr != nil && err == nil {
		t.Fatalf("Should have failed")
	}
}

func TestCountResultItemToString(t *testing.T) {
	tests := []struct {
		item     *CountResultItem
		expected string
	}{
		{item: &CountResultItem{ID: 100, Count: 50}, expected: "100:50"},
		{item: &CountResultItem{Key: "blah", Count: 50}, expected: "blah:50"},
		{item: &CountResultItem{Key: "blah", ID: 22, Count: 50}, expected: "blah:50"},
		{item: &CountResultItem{Key: "blah", ID: 22}, expected: "blah:0"},
		{item: &CountResultItem{}, expected: "0:0"},
	}

	for i, tst := range tests {
		t.Run(fmt.Sprintf("%d: ", i), func(t *testing.T) {
			if tst.expected != tst.item.String() {
				t.Fatalf("%s != %s", tst.expected, tst.item.String())
			}
		})
	}
}

func TestMarshalResults(t *testing.T) {
	attrs := []*pbuf.Attr{
		{Key: "name", StringValue: "some string", Type: 1},
		{Key: "age", IntValue: 95, Type: 2},
		{Key: "registered", BoolValue: true, Type: 3},
		{Key: "height", FloatValue: 1.83, Type: 4},
	}
	bitmap := &pbuf.Bitmap{
		Attrs: attrs,
		Bits:  []uint64{5, 10},
	}
	pairs := []*pbuf.Pair{
		{ID: 10, Count: 100},
	}
	pbufResults := []*pbuf.QueryResult{
		{Type: QueryResultTypeBitmap, Bitmap: bitmap},
		{Type: QueryResultTypePairs, Pairs: pairs},
	}
	resultJSONStrings := make([]string, len(pbufResults))
	for i, pr := range pbufResults {
		r, err := newQueryResultFromInternal(pr)
		if err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		resultJSONStrings[i] = string(b)
	}
	targetJSON := []string{
		`{"attrs":{"age":95,"height":1.83,"name":"some string","registered":true},"bits":[5,10],"keys":[]}`,
		`[{"id":10,"count":100}]`,
	}
	for i := range targetJSON {
		if sortedString(targetJSON[i]) != sortedString(resultJSONStrings[i]) {
			t.Fatalf("%v != %v ", targetJSON[i], resultJSONStrings[i])
		}
	}

}

func TestUnknownQueryResultType(t *testing.T) {
	result := &pbuf.QueryResult{
		Type: 999,
	}
	_, err := newQueryResultFromInternal(result)
	if err != ErrUnknownType {
		t.Fatalf("Should have failed with ErrUnknownType")
	}
}

func TestTopNResult(t *testing.T) {
	result := TopNResult{
		CountResultItem{ID: 100, Count: 10},
	}
	expectResult(t, result, QueryResultTypePairs, BitmapResult{}, []CountResultItem{{100, "", 10}}, 0, 0, false)
}

func TestBitmapResult(t *testing.T) {
	result := BitmapResult{
		Bits: []uint64{1, 2, 3},
	}
	targetBmp := BitmapResult{
		Bits: []uint64{1, 2, 3},
	}
	expectResult(t, result, QueryResultTypeBitmap, targetBmp, nil, 0, 0, false)
}

func TestBitmapResultNilBits(t *testing.T) {
	result := BitmapResult{
		Bits: nil,
	}
	_, err := result.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSumCountResult(t *testing.T) {
	result := ValCountResult{
		Val: 100,
		Cnt: 50,
	}
	expectResult(t, result, QueryResultTypeValCount, BitmapResult{}, nil, 100, 50, false)
}

func TestIntResult(t *testing.T) {
	result := IntResult(11)
	expectResult(t, result, QueryResultTypeUint64, BitmapResult{}, nil, 0, 11, false)
}

func TestBoolResult(t *testing.T) {
	result := BoolResult(true)
	expectResult(t, result, QueryResultTypeBool, BitmapResult{}, nil, 0, 0, true)
}

func TestNilResult(t *testing.T) {
	result := NilResult{}
	expectResult(t, result, QueryResultTypeNil, BitmapResult{}, nil, 0, 0, false)
}

func expectResult(t *testing.T, r QueryResult, resultType uint32, bmp BitmapResult, countItems []CountResultItem, sum int64, count int64, changed bool) {
	if resultType != r.Type() {
		log.Fatalf("Result type: %d != %d", resultType, r.Type())
	}
	if !reflect.DeepEqual(bmp, r.Bitmap()) {
		log.Fatalf("Bitmap: %v != %v", bmp, r.Bitmap())
	}
	if !reflect.DeepEqual(countItems, r.CountItems()) {
		log.Fatalf("Count items: %v != %v", countItems, r.CountItems())
	}
	if count != r.Count() {
		log.Fatalf("Count: %d != %d", count, r.Count())
	}
	if sum != r.Value() {
		log.Fatalf("Sum: %d != %d", sum, r.Value())
	}
	if changed != r.Changed() {
		log.Fatalf("Changed: %v != %v", changed, r.Changed())
	}
}
