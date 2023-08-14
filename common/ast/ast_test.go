// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ast_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/cel-go/common"
	"github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/overloads"
	"github.com/google/cel-go/common/types"
)

func TestASTNilSafety(t *testing.T) {
	ex, err := ast.ProtoToExpr(nil)
	if err != nil {
		t.Fatalf("ast.ProtoToExpr() failed: %v", err)
	}
	info, err := ast.ProtoToSourceInfo(nil)
	if err != nil {
		t.Fatalf("ast.ProtoToSourceInfo() failed: %v", err)
	}
	tests := []*ast.AST{
		nil,
		ast.NewAST(nil, nil),
		ast.NewCheckedAST(nil, nil, nil),
		ast.NewCheckedAST(ast.NewAST(nil, nil), nil, nil),
		ast.NewAST(ex, info),
		ast.NewCheckedAST(ast.NewAST(ex, info), map[int64]*types.Type{}, map[int64]*ast.ReferenceInfo{}),
	}
	for i, tst := range tests {
		tc := tst
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			testAST := tc
			if testAST.Expr().ID() != 0 {
				t.Errorf("Expr().ID() got %v, wanted 0", testAST.Expr().ID())
			}
			if testAST.SourceInfo().SyntaxVersion() != "" {
				t.Errorf("SourceInfo().SyntaxVersion() got %s, wanted empty string", testAST.SourceInfo().SyntaxVersion())
			}
			if testAST.IsChecked() {
				t.Error("IsChecked() returned true, wanted false")
			}
			if testAST.GetType(testAST.Expr().ID()) != types.DynType {
				t.Errorf("GetType() got %v, wanted dyn", testAST.GetType(testAST.Expr().ID()))
			}
			if len(testAST.GetOverloadIDs(testAST.Expr().ID())) != 0 {
				t.Errorf("GetOverloadIDs() got %v, wanted empty set", testAST.GetOverloadIDs(testAST.Expr().ID()))
			}
		})
	}
}

func TestSourceInfo(t *testing.T) {
	src := common.NewStringSource("a\n? b\n: c", "custom description")
	info := ast.NewSourceInfo(src)
	if info.Description() != "custom description" {
		t.Errorf("Description() got %s, wanted 'custom description'", info.Description())
	}
	if len(info.LineOffsets()) != 3 {
		t.Errorf("LineOffsets() got %v, wanted 3 offsets", info.LineOffsets())
	}
	info.SetOffsetRange(1, ast.OffsetRange{Start: 0, Stop: 1}) // a
	info.SetOffsetRange(2, ast.OffsetRange{Start: 4, Stop: 5}) // b
	info.SetOffsetRange(3, ast.OffsetRange{Start: 8, Stop: 9}) // c
	if !reflect.DeepEqual(info.GetStartLocation(1), common.NewLocation(1, 0)) {
		t.Errorf("info.GetStartLocation(1) got %v, wanted line 1, col 0", info.GetStartLocation(1))
	}
	if !reflect.DeepEqual(info.GetStopLocation(1), common.NewLocation(1, 1)) {
		t.Errorf("info.GetStopLocation(1) got %v, wanted line 1, col 1", info.GetStopLocation(1))
	}
	if !reflect.DeepEqual(info.GetStartLocation(2), common.NewLocation(2, 2)) {
		t.Errorf("info.GetStartLocation(2) got %v, wanted line 2, col 2", info.GetStartLocation(2))
	}
	if !reflect.DeepEqual(info.GetStopLocation(2), common.NewLocation(2, 3)) {
		t.Errorf("info.GetStopLocation(2) got %v, wanted line 2, col 3", info.GetStopLocation(2))
	}
	if !reflect.DeepEqual(info.GetStartLocation(3), common.NewLocation(3, 2)) {
		t.Errorf("info.GetStartLocation(3) got %v, wanted line 2, col 2", info.GetStartLocation(3))
	}
	if !reflect.DeepEqual(info.GetStopLocation(3), common.NewLocation(3, 3)) {
		t.Errorf("info.GetStopLocation(3) got %v, wanted line 2, col 3", info.GetStopLocation(3))
	}
	if info.ComputeOffset(3, 2) != 8 {
		t.Errorf("info.ComputeOffset(3, 2) got %d, wanted 8", info.ComputeOffset(3, 2))
	}
}

func TestSourceInfoNilSafety(t *testing.T) {
	info, err := ast.ProtoToSourceInfo(nil)
	if err != nil {
		t.Fatalf("ast.ProtoToSourceInfo() failed: %v", err)
	}
	tests := []*ast.SourceInfo{
		nil,
		info,
		ast.NewSourceInfo(nil),
	}
	for i, tst := range tests {
		tc := tst
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			testInfo := tc
			if testInfo.SyntaxVersion() != "" {
				t.Errorf("SyntaxVersion() got %s, wanted empty string", testInfo.SyntaxVersion())
			}
			if testInfo.Description() != "" {
				t.Errorf("Description() got %s, wanted empty string", testInfo.Description())
			}
			if len(testInfo.LineOffsets()) != 0 {
				t.Errorf("LineOffsets() got %v, wanted empty list", testInfo.LineOffsets())
			}
			if len(testInfo.MacroCalls()) != 0 {
				t.Errorf("MacroCalls() got %v, wanted empty map", testInfo.MacroCalls())
			}
			if call, found := testInfo.GetMacroCall(0); found {
				t.Errorf("GetMacroCall(0) got %v, wanted not found", call)
			}
			if r, found := testInfo.GetOffsetRange(0); found {
				t.Errorf("GetOffsetRange(0) got %v, wanted not found", r)
			}
			if loc := testInfo.GetStartLocation(0); loc != common.NoLocation {
				t.Errorf("GetStartLocation(0) got %v, wanted no location", loc)
			}
			if loc := testInfo.GetStopLocation(0); loc != common.NoLocation {
				t.Errorf("GetStopLocation(0) got %v, wanted no location", loc)
			}
			if off := testInfo.ComputeOffset(1, 0); off != 0 {
				t.Errorf("ComputeOffset(1, 0) got %d, wanted 0", off)
			}
			if off := testInfo.ComputeOffset(-2, 0); off != -1 {
				t.Errorf("ComputeOffset(-2, 0) got %d, wanted -1", off)
			}
			if off := testInfo.ComputeOffset(2, 0); off != -1 {
				t.Errorf("ComputeOffset(2, 0) got %d, wanted -1", off)
			}
		})
	}
}

func TestReferenceInfoEquals(t *testing.T) {
	tests := []struct {
		name  string
		a     *ast.ReferenceInfo
		b     *ast.ReferenceInfo
		equal bool
	}{
		{
			name:  "single overload equal",
			a:     ast.NewFunctionReference(overloads.AddBytes),
			b:     ast.NewFunctionReference(overloads.AddBytes),
			equal: true,
		},
		{
			name:  "single overload not equal",
			a:     ast.NewFunctionReference(overloads.AddBytes),
			b:     ast.NewFunctionReference(overloads.AddDouble),
			equal: false,
		},
		{
			name:  "single and multiple overload not equal",
			a:     ast.NewFunctionReference(overloads.AddBytes),
			b:     ast.NewFunctionReference(overloads.AddBytes, overloads.AddDouble),
			equal: false,
		},
		{
			name:  "multiple overloads equal",
			a:     ast.NewFunctionReference(overloads.AddBytes, overloads.AddDouble),
			b:     ast.NewFunctionReference(overloads.AddDouble, overloads.AddBytes),
			equal: true,
		},
		{
			name:  "identifier reference equal",
			a:     ast.NewIdentReference("BYTES", nil),
			b:     ast.NewIdentReference("BYTES", nil),
			equal: true,
		},
		{
			name:  "identifier reference not equal",
			a:     ast.NewIdentReference("BYTES", nil),
			b:     ast.NewIdentReference("TRUE", nil),
			equal: false,
		},
		{
			name:  "identifier and constant reference not equal",
			a:     ast.NewIdentReference("BYTES", nil),
			b:     ast.NewIdentReference("BYTES", types.Bytes("bytes")),
			equal: false,
		},
		{
			name:  "constant references equal",
			a:     ast.NewIdentReference("BYTES", types.Bytes("bytes")),
			b:     ast.NewIdentReference("BYTES", types.Bytes("bytes")),
			equal: true,
		},
		{
			name:  "constant references not equal",
			a:     ast.NewIdentReference("BYTES", types.Bytes("bytes")),
			b:     ast.NewIdentReference("BYTES", types.Bytes("bytes-other")),
			equal: false,
		},
		{
			name:  "constant and overload reference not equal",
			a:     ast.NewIdentReference("BYTES", types.Bytes("bytes")),
			b:     ast.NewFunctionReference(overloads.AddDouble, overloads.AddBytes),
			equal: false,
		},
	}
	for _, tst := range tests {
		tc := tst
		t.Run(tc.name, func(t *testing.T) {
			out := tc.a.Equals(tc.b)
			if out != tc.equal {
				t.Errorf("%v.Equals(%v) got %v, wanted %v", tc.a, tc.b, out, tc.equal)
			}
		})
	}
}

func TestReferenceInfoAddOverload(t *testing.T) {
	add := ast.NewFunctionReference(overloads.AddBytes)
	add.AddOverload(overloads.AddDouble)
	if !add.Equals(ast.NewFunctionReference(overloads.AddBytes, overloads.AddDouble)) {
		t.Error("AddOverload() did not produce equal references")
	}
	add.AddOverload(overloads.AddDouble)
	if !add.Equals(ast.NewFunctionReference(overloads.AddBytes, overloads.AddDouble)) {
		t.Error("repeated AddOverload() did not produce equal references")
	}
}