// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proto

import (
	"fmt"
	"testing"

	"github.com/youtube/vitess/go/bson"
)

type reflectStreamEvent struct {
	Category   string
	TableName  string
	PKColNames []string
	PKValues   [][]interface{}
	Sql        string
	Timestamp  int64
	GroupId    string
}

type badStreamEvent struct {
	Extra      int
	Category   string
	TableName  string
	PKColNames []string
	PKValues   [][]interface{}
	Sql        string
	Timestamp  int64
	GroupId    string
}

func TestStreamEvent(t *testing.T) {
	reflected, err := bson.Marshal(&reflectStreamEvent{
		Category:   "str1",
		TableName:  "str2",
		PKColNames: []string{"str3", "str4"},
		PKValues: [][]interface{}{
			[]interface{}{
				"str5", 1, uint64(0xffffffffffffffff),
			},
			[]interface{}{
				"str6", 2, uint64(0xfffffffffffffffe),
			},
		},
		Sql:       "str7",
		Timestamp: 3,
		GroupId:   "str8",
	})
	if err != nil {
		t.Error(err)
	}
	want := string(reflected)

	custom := StreamEvent{
		Category:   "str1",
		TableName:  "str2",
		PKColNames: []string{"str3", "str4"},
		PKValues: [][]interface{}{
			[]interface{}{
				"str5", 1, uint64(0xffffffffffffffff),
			},
			[]interface{}{
				"str6", 2, uint64(0xfffffffffffffffe),
			},
		},
		Sql:       "str7",
		Timestamp: 3,
		GroupId:   "str8",
	}
	encoded, err := bson.Marshal(&custom)
	if err != nil {
		t.Error(err)
	}
	got := string(encoded)
	if want != got {
		t.Errorf("want\n%#v, got\n%#v", want, got)
	}

	var unmarshalled StreamEvent
	err = bson.Unmarshal(encoded, &unmarshalled)
	if err != nil {
		t.Error(err)
	}
	want = fmt.Sprintf("%#v", custom)
	got = fmt.Sprintf("%#v", unmarshalled)
	if want != got {
		t.Errorf("want\n%#v, got\n%#v", want, got)
	}

	unexpected, err := bson.Marshal(&badStreamEvent{})
	if err != nil {
		t.Error(err)
	}
	err = bson.Unmarshal(unexpected, &unmarshalled)
	want = "Unrecognized tag Extra"
	if err == nil || want != err.Error() {
		t.Errorf("want %v, got %v", want, err)
	}
}
