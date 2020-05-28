package oviewer

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/gdamore/tcell"
)

func Test_parseString(t *testing.T) {
	type args struct {
		line     string
		tabWidth int
	}
	tests := []struct {
		name string
		args args
		want lineContents
	}{
		{
			name: "test1",
			args: args{
				line: "test", tabWidth: 8,
			},
			want: lineContents{
				contents: []Content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('t'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('e'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('s'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('t'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 2, 3: 3, 4: 4},
			},
		},
		{
			name: "testASCII",
			args: args{line: "abc", tabWidth: 4},
			want: lineContents{
				contents: []Content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('b'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('c'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 2, 3: 3},
			},
		},
		{
			name: "testHiragana",
			args: args{line: "あ", tabWidth: 4},
			want: lineContents{
				contents: []Content{
					{width: 2, style: tcell.StyleDefault, mainc: rune('あ'), combc: nil},
					{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
				},
				byteMap: map[int]int{0: 0, 3: 2},
			},
		},
		{
			name: "testKANJI",
			args: args{line: "漢", tabWidth: 4},
			want: lineContents{
				contents: []Content{
					{width: 2, style: tcell.StyleDefault, mainc: rune('漢'), combc: nil},
					{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
				},
				byteMap: map[int]int{0: 0, 3: 2},
			},
		},
		{
			name: "testMIX",
			args: args{line: "abc漢", tabWidth: 4},
			want: lineContents{
				contents: []Content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('b'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('c'), combc: nil},
					{width: 2, style: tcell.StyleDefault, mainc: rune('漢'), combc: nil},
					{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 2, 3: 3, 6: 5},
			},
		},
		{
			name: "testTab",
			args: args{line: "a\tb", tabWidth: 4},
			want: lineContents{
				contents: []Content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('b'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 4, 3: 5},
			},
		},
		{
			name: "testTabMinus",
			args: args{line: "a\tb", tabWidth: -1},
			want: lineContents{
				contents: []Content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Reverse(true), mainc: rune('\\'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Reverse(true), mainc: rune('t'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('b'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 1, 3: 3, 4: 4},
			},
		},
		{
			name: "red",
			args: args{
				line: "\x1B[31mred\x1B[m", tabWidth: 8,
			},
			want: lineContents{
				contents: []Content{
					{width: 1, style: tcell.StyleDefault.Foreground(tcell.Color(1)), mainc: rune('r'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Foreground(tcell.Color(1)), mainc: rune('e'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Foreground(tcell.Color(1)), mainc: rune('d'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 2, 3: 3},
			},
		},
		{
			name: "bold",
			args: args{
				line: "\x1B[1mbold\x1B[m", tabWidth: 8,
			},
			want: lineContents{
				contents: []Content{
					{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('b'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('o'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('l'), combc: nil},
					{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('d'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 2, 3: 3, 4: 4},
			},
		},
		{
			name: "testOverstrike",
			args: args{line: "a\ba", tabWidth: 8},
			want: lineContents{
				contents: []Content{
					{width: 1, style: tcell.StyleDefault.Bold(true), mainc: rune('a'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 0, 3: 1},
			},
		},
		{
			name: "testOverstrike2",
			args: args{line: "\ba", tabWidth: 8},
			want: lineContents{
				contents: []Content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('a'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 0, 2: 1},
			},
		},
		{
			name: "testOverstrike3",
			args: args{line: "あ\bあ", tabWidth: 8},
			want: lineContents{
				contents: []Content{
					{width: 2, style: tcell.StyleDefault.Bold(true), mainc: rune('あ'), combc: nil},
					{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
				},
				byteMap: map[int]int{0: 0, 3: 2, 4: 0, 7: 2},
			},
		},
		{
			name: "testOverstrike4",
			args: args{line: "\a", tabWidth: 8},
			want: lineContents{
				contents: nil,
				byteMap:  map[int]int{0: 0, 1: 0},
			},
		},
		{
			name: "testOverstrikeUnderLine",
			args: args{line: "_\ba", tabWidth: 8},
			want: lineContents{
				contents: []Content{
					{width: 1, style: tcell.StyleDefault.Underline(true), mainc: rune('a'), combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 0, 3: 1},
			},
		},
		{
			name: "testOverstrikeUnderLine2",
			args: args{line: "_\bあ", tabWidth: 8},
			want: lineContents{
				contents: []Content{
					{width: 2, style: tcell.StyleDefault.Underline(true), mainc: rune('あ'), combc: nil},
					{width: 0, style: tcell.StyleDefault, mainc: 0, combc: nil},
				},
				byteMap: map[int]int{0: 0, 1: 1, 2: 0, 5: 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseString(tt.args.line, tt.args.tabWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lastContent(t *testing.T) {
	type args struct {
		contents []Content
	}
	tests := []struct {
		name string
		args args
		want Content
	}{
		{
			name: "tsetNil",
			args: args{
				contents: nil,
			},
			want: Content{},
		},
		{
			name: "tset1",
			args: args{
				contents: []Content{
					{width: 1, style: tcell.StyleDefault, mainc: rune('t'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('e'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('s'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune('t'), combc: nil},
				},
			},
			want: Content{width: 1, style: tcell.StyleDefault, mainc: rune('t'), combc: nil},
		},
		{
			name: "tsetWide",
			args: args{
				contents: []Content{
					{width: 2, style: tcell.StyleDefault, mainc: rune('あ'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: nil},
					{width: 2, style: tcell.StyleDefault, mainc: rune('い'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: nil},
					{width: 2, style: tcell.StyleDefault, mainc: rune('う'), combc: nil},
					{width: 1, style: tcell.StyleDefault, mainc: rune(' '), combc: nil},
				},
			},
			want: Content{width: 2, style: tcell.StyleDefault, mainc: rune('う'), combc: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lastContent(tt.args.contents); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lastContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_csToStyle(t *testing.T) {
	type args struct {
		style        tcell.Style
		csiParameter *bytes.Buffer
	}
	tests := []struct {
		name string
		args args
		want tcell.Style
	}{
		{
			name: "color8bit",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: bytes.NewBufferString("38;5;1"),
			},
			want: tcell.StyleDefault.Foreground(tcell.ColorMaroon),
		},
		{
			name: "color8bit2",
			args: args{
				style:        tcell.StyleDefault,
				csiParameter: bytes.NewBufferString("38;5;21"),
			},
			want: tcell.StyleDefault.Foreground(tcell.GetColor("#0000ff")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := csToStyle(tt.args.style, tt.args.csiParameter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("csToStyle() = %v, want %v", got, tt.want)
				gfg, gbg, gattr := got.Decompose()
				wfg, wbg, wattr := tt.want.Decompose()
				t.Errorf("csToStyle() = %x,%x,%v, want %x,%x,%v", gfg.Hex(), gbg.Hex(), gattr, wfg.Hex(), wbg.Hex(), wattr)
			}
		})
	}
}

func Test_strToContents(t *testing.T) {
	type args struct {
		line     string
		tabWidth int
	}
	tests := []struct {
		name string
		args args
		want []Content
	}{
		{
			name: "test1",
			args: args{line: "1", tabWidth: 4},
			want: []Content{
				{width: 1, style: tcell.StyleDefault, mainc: rune('1'), combc: nil},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strToContents(tt.args.line, tt.args.tabWidth); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("strToContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
