package engine

import (
	"bytes"
	"testing"
)

// These tests pin what a frame of palette colors writes. The expected bytes were
// captured from the renderer as it was before colors became a type, so they show that
// the redesign changed no output for frames that use only what existed then.

// goldenPalette maps the digits in the frame tables below to colors. 0 is the default.
var goldenPalette = [...]Color{{}, Palette(1), Palette(4), Palette(16), Palette(208), Palette(255)}

// goldenCanvasPalette maps the digits in the canvas tables to pixels. 0 is empty.
var goldenCanvasPalette = [...]Color{{}, Palette(16), Palette(196), Palette(21), Palette(46)}

type goldenCase struct {
	name         string
	chars        []string // 16 characters per row
	fg, bg       []string // a digit per cell, into goldenPalette
	clipW, clipH int      // 0: the whole frame
	want         string
}

var goldenFrames = []goldenCase{
	{"plain", []string{"plain text here.", "................", "                ", "abcdefghijklmnop"}, []string{"0000000000000000", "0000000000000000", "0000000000000000", "0000000000000000"}, []string{"0000000000000000", "0000000000000000", "0000000000000000", "0000000000000000"}, 0, 0, "\x1b[1;1Hplain text here.\x1b[2;1H................\x1b[3;1H                \x1b[4;1Habcdefghijklmnop"},
	{"plain clipped 7x3", []string{"plain text here.", "................", "                ", "abcdefghijklmnop"}, []string{"0000000000000000", "0000000000000000", "0000000000000000", "0000000000000000"}, []string{"0000000000000000", "0000000000000000", "0000000000000000", "0000000000000000"}, 7, 3, "\x1b[1;1Hplain t\x1b[2;1H.......\x1b[3;1H       "},
	{"plain clipped 16x2", []string{"plain text here.", "................", "                ", "abcdefghijklmnop"}, []string{"0000000000000000", "0000000000000000", "0000000000000000", "0000000000000000"}, []string{"0000000000000000", "0000000000000000", "0000000000000000", "0000000000000000"}, 16, 2, "\x1b[1;1Hplain text here.\x1b[2;1H................"},
	{"plain clipped 5x7", []string{"plain text here.", "................", "                ", "abcdefghijklmnop"}, []string{"0000000000000000", "0000000000000000", "0000000000000000", "0000000000000000"}, []string{"0000000000000000", "0000000000000000", "0000000000000000", "0000000000000000"}, 5, 7, "\x1b[1;1Hplain\x1b[2;1H.....\x1b[3;1H     \x1b[4;1Habcde"},
	{"plain clipped 20x20", []string{"plain text here.", "................", "                ", "abcdefghijklmnop"}, []string{"0000000000000000", "0000000000000000", "0000000000000000", "0000000000000000"}, []string{"0000000000000000", "0000000000000000", "0000000000000000", "0000000000000000"}, 20, 20, "\x1b[1;1Hplain text here.\x1b[2;1H................\x1b[3;1H                \x1b[4;1Habcdefghijklmnop"},
	{"foreground runs", []string{"RRRgggBBBxxxxyyy", "RRRgggBBBxxxxyyy", "aaaabbbbccccdddd"}, []string{"1112224441234500", "1111111111111111", "1100110011001100"}, []string{"0000000000000000", "0000000000000000", "0000000000000000"}, 0, 0, "\x1b[1;1H\x1b[38;5;1mRRR\x1b[38;5;4mggg\x1b[38;5;208mBBB\x1b[38;5;1mx\x1b[38;5;4mx\x1b[38;5;16mx\x1b[38;5;208mx\x1b[38;5;255my\x1b[0myy\x1b[2;1H\x1b[38;5;1mRRRgggBBBxxxxyyy\x1b[0m\x1b[3;1H\x1b[38;5;1maa\x1b[0maa\x1b[38;5;1mbb\x1b[0mbb\x1b[38;5;1mcc\x1b[0mcc\x1b[38;5;1mdd\x1b[0mdd"},
	{"foreground runs clipped 7x3", []string{"RRRgggBBBxxxxyyy", "RRRgggBBBxxxxyyy", "aaaabbbbccccdddd"}, []string{"1112224441234500", "1111111111111111", "1100110011001100"}, []string{"0000000000000000", "0000000000000000", "0000000000000000"}, 7, 3, "\x1b[1;1H\x1b[38;5;1mRRR\x1b[38;5;4mggg\x1b[38;5;208mB\x1b[0m\x1b[2;1H\x1b[38;5;1mRRRgggB\x1b[0m\x1b[3;1H\x1b[38;5;1maa\x1b[0maa\x1b[38;5;1mbb\x1b[0mb"},
	{"foreground runs clipped 16x2", []string{"RRRgggBBBxxxxyyy", "RRRgggBBBxxxxyyy", "aaaabbbbccccdddd"}, []string{"1112224441234500", "1111111111111111", "1100110011001100"}, []string{"0000000000000000", "0000000000000000", "0000000000000000"}, 16, 2, "\x1b[1;1H\x1b[38;5;1mRRR\x1b[38;5;4mggg\x1b[38;5;208mBBB\x1b[38;5;1mx\x1b[38;5;4mx\x1b[38;5;16mx\x1b[38;5;208mx\x1b[38;5;255my\x1b[0myy\x1b[2;1H\x1b[38;5;1mRRRgggBBBxxxxyyy\x1b[0m"},
	{"foreground runs clipped 5x7", []string{"RRRgggBBBxxxxyyy", "RRRgggBBBxxxxyyy", "aaaabbbbccccdddd"}, []string{"1112224441234500", "1111111111111111", "1100110011001100"}, []string{"0000000000000000", "0000000000000000", "0000000000000000"}, 5, 7, "\x1b[1;1H\x1b[38;5;1mRRR\x1b[38;5;4mgg\x1b[0m\x1b[2;1H\x1b[38;5;1mRRRgg\x1b[0m\x1b[3;1H\x1b[38;5;1maa\x1b[0maa\x1b[38;5;1mb\x1b[0m"},
	{"foreground runs clipped 20x20", []string{"RRRgggBBBxxxxyyy", "RRRgggBBBxxxxyyy", "aaaabbbbccccdddd"}, []string{"1112224441234500", "1111111111111111", "1100110011001100"}, []string{"0000000000000000", "0000000000000000", "0000000000000000"}, 20, 20, "\x1b[1;1H\x1b[38;5;1mRRR\x1b[38;5;4mggg\x1b[38;5;208mBBB\x1b[38;5;1mx\x1b[38;5;4mx\x1b[38;5;16mx\x1b[38;5;208mx\x1b[38;5;255my\x1b[0myy\x1b[2;1H\x1b[38;5;1mRRRgggBBBxxxxyyy\x1b[0m\x1b[3;1H\x1b[38;5;1maa\x1b[0maa\x1b[38;5;1mbb\x1b[0mbb\x1b[38;5;1mcc\x1b[0mcc\x1b[38;5;1mdd\x1b[0mdd"},
	{"background runs", []string{"background runs.", "                ", "    gap    gap  "}, []string{"0000000000000000", "0000000000000000", "0000000000000000"}, []string{"0001111444400550", "1111111111111111", "0000555500005555"}, 0, 0, "\x1b[1;1Hbac\x1b[48;5;1mkgro\x1b[48;5;208mund \x1b[0mru\x1b[48;5;255mns\x1b[0m.\x1b[2;1H\x1b[48;5;1m                \x1b[0m\x1b[3;1H    \x1b[48;5;255mgap \x1b[0m   g\x1b[48;5;255map  \x1b[0m"},
	{"background runs clipped 7x3", []string{"background runs.", "                ", "    gap    gap  "}, []string{"0000000000000000", "0000000000000000", "0000000000000000"}, []string{"0001111444400550", "1111111111111111", "0000555500005555"}, 7, 3, "\x1b[1;1Hbac\x1b[48;5;1mkgro\x1b[0m\x1b[2;1H\x1b[48;5;1m       \x1b[0m\x1b[3;1H    \x1b[48;5;255mgap\x1b[0m"},
	{"background runs clipped 16x2", []string{"background runs.", "                ", "    gap    gap  "}, []string{"0000000000000000", "0000000000000000", "0000000000000000"}, []string{"0001111444400550", "1111111111111111", "0000555500005555"}, 16, 2, "\x1b[1;1Hbac\x1b[48;5;1mkgro\x1b[48;5;208mund \x1b[0mru\x1b[48;5;255mns\x1b[0m.\x1b[2;1H\x1b[48;5;1m                \x1b[0m"},
	{"background runs clipped 5x7", []string{"background runs.", "                ", "    gap    gap  "}, []string{"0000000000000000", "0000000000000000", "0000000000000000"}, []string{"0001111444400550", "1111111111111111", "0000555500005555"}, 5, 7, "\x1b[1;1Hbac\x1b[48;5;1mkg\x1b[0m\x1b[2;1H\x1b[48;5;1m     \x1b[0m\x1b[3;1H    \x1b[48;5;255mg\x1b[0m"},
	{"background runs clipped 20x20", []string{"background runs.", "                ", "    gap    gap  "}, []string{"0000000000000000", "0000000000000000", "0000000000000000"}, []string{"0001111444400550", "1111111111111111", "0000555500005555"}, 20, 20, "\x1b[1;1Hbac\x1b[48;5;1mkgro\x1b[48;5;208mund \x1b[0mru\x1b[48;5;255mns\x1b[0m.\x1b[2;1H\x1b[48;5;1m                \x1b[0m\x1b[3;1H    \x1b[48;5;255mgap \x1b[0m   g\x1b[48;5;255map  \x1b[0m"},
	{"both colors", []string{"both colors here", "both colors here", "ABCDEFGHIJKLMNOP"}, []string{"1111222244440000", "1212121212121212", "4444444400000000"}, []string{"0000444411112222", "5555555500000000", "0011223344550011"}, 0, 0, "\x1b[1;1H\x1b[38;5;1mboth\x1b[38;5;4;48;5;208m col\x1b[38;5;208;48;5;1mors \x1b[39;48;5;4mhere\x1b[0m\x1b[2;1H\x1b[38;5;1;48;5;255mb\x1b[38;5;4mo\x1b[38;5;1mt\x1b[38;5;4mh\x1b[38;5;1m \x1b[38;5;4mc\x1b[38;5;1mo\x1b[38;5;4ml\x1b[38;5;1;49mo\x1b[38;5;4mr\x1b[38;5;1ms\x1b[38;5;4m \x1b[38;5;1mh\x1b[38;5;4me\x1b[38;5;1mr\x1b[38;5;4me\x1b[0m\x1b[3;1H\x1b[38;5;208mAB\x1b[48;5;1mCD\x1b[48;5;4mEF\x1b[48;5;16mGH\x1b[39;48;5;208mIJ\x1b[48;5;255mKL\x1b[0mMN\x1b[48;5;1mOP\x1b[0m"},
	{"both colors clipped 7x3", []string{"both colors here", "both colors here", "ABCDEFGHIJKLMNOP"}, []string{"1111222244440000", "1212121212121212", "4444444400000000"}, []string{"0000444411112222", "5555555500000000", "0011223344550011"}, 7, 3, "\x1b[1;1H\x1b[38;5;1mboth\x1b[38;5;4;48;5;208m co\x1b[0m\x1b[2;1H\x1b[38;5;1;48;5;255mb\x1b[38;5;4mo\x1b[38;5;1mt\x1b[38;5;4mh\x1b[38;5;1m \x1b[38;5;4mc\x1b[38;5;1mo\x1b[0m\x1b[3;1H\x1b[38;5;208mAB\x1b[48;5;1mCD\x1b[48;5;4mEF\x1b[48;5;16mG\x1b[0m"},
	{"both colors clipped 16x2", []string{"both colors here", "both colors here", "ABCDEFGHIJKLMNOP"}, []string{"1111222244440000", "1212121212121212", "4444444400000000"}, []string{"0000444411112222", "5555555500000000", "0011223344550011"}, 16, 2, "\x1b[1;1H\x1b[38;5;1mboth\x1b[38;5;4;48;5;208m col\x1b[38;5;208;48;5;1mors \x1b[39;48;5;4mhere\x1b[0m\x1b[2;1H\x1b[38;5;1;48;5;255mb\x1b[38;5;4mo\x1b[38;5;1mt\x1b[38;5;4mh\x1b[38;5;1m \x1b[38;5;4mc\x1b[38;5;1mo\x1b[38;5;4ml\x1b[38;5;1;49mo\x1b[38;5;4mr\x1b[38;5;1ms\x1b[38;5;4m \x1b[38;5;1mh\x1b[38;5;4me\x1b[38;5;1mr\x1b[38;5;4me\x1b[0m"},
	{"both colors clipped 5x7", []string{"both colors here", "both colors here", "ABCDEFGHIJKLMNOP"}, []string{"1111222244440000", "1212121212121212", "4444444400000000"}, []string{"0000444411112222", "5555555500000000", "0011223344550011"}, 5, 7, "\x1b[1;1H\x1b[38;5;1mboth\x1b[38;5;4;48;5;208m \x1b[0m\x1b[2;1H\x1b[38;5;1;48;5;255mb\x1b[38;5;4mo\x1b[38;5;1mt\x1b[38;5;4mh\x1b[38;5;1m \x1b[0m\x1b[3;1H\x1b[38;5;208mAB\x1b[48;5;1mCD\x1b[48;5;4mE\x1b[0m"},
	{"both colors clipped 20x20", []string{"both colors here", "both colors here", "ABCDEFGHIJKLMNOP"}, []string{"1111222244440000", "1212121212121212", "4444444400000000"}, []string{"0000444411112222", "5555555500000000", "0011223344550011"}, 20, 20, "\x1b[1;1H\x1b[38;5;1mboth\x1b[38;5;4;48;5;208m col\x1b[38;5;208;48;5;1mors \x1b[39;48;5;4mhere\x1b[0m\x1b[2;1H\x1b[38;5;1;48;5;255mb\x1b[38;5;4mo\x1b[38;5;1mt\x1b[38;5;4mh\x1b[38;5;1m \x1b[38;5;4mc\x1b[38;5;1mo\x1b[38;5;4ml\x1b[38;5;1;49mo\x1b[38;5;4mr\x1b[38;5;1ms\x1b[38;5;4m \x1b[38;5;1mh\x1b[38;5;4me\x1b[38;5;1mr\x1b[38;5;4me\x1b[0m\x1b[3;1H\x1b[38;5;208mAB\x1b[48;5;1mCD\x1b[48;5;4mEF\x1b[48;5;16mGH\x1b[39;48;5;208mIJ\x1b[48;5;255mKL\x1b[0mMN\x1b[48;5;1mOP\x1b[0m"},
	{"color at the end of a row", []string{"edge cell color!", "edge cell color!", "no color at end."}, []string{"0000000000000004", "0000000000000044", "4000000000000000"}, []string{"0000000000000001", "1000000000000000", "0000000000000000"}, 0, 0, "\x1b[1;1Hedge cell color\x1b[38;5;208;48;5;1m!\x1b[0m\x1b[2;1H\x1b[48;5;1me\x1b[0mdge cell colo\x1b[38;5;208mr!\x1b[0m\x1b[3;1H\x1b[38;5;208mn\x1b[0mo color at end."},
	{"color at the end of a row clipped 7x3", []string{"edge cell color!", "edge cell color!", "no color at end."}, []string{"0000000000000004", "0000000000000044", "4000000000000000"}, []string{"0000000000000001", "1000000000000000", "0000000000000000"}, 7, 3, "\x1b[1;1Hedge ce\x1b[2;1H\x1b[48;5;1me\x1b[0mdge ce\x1b[3;1H\x1b[38;5;208mn\x1b[0mo colo"},
	{"color at the end of a row clipped 16x2", []string{"edge cell color!", "edge cell color!", "no color at end."}, []string{"0000000000000004", "0000000000000044", "4000000000000000"}, []string{"0000000000000001", "1000000000000000", "0000000000000000"}, 16, 2, "\x1b[1;1Hedge cell color\x1b[38;5;208;48;5;1m!\x1b[0m\x1b[2;1H\x1b[48;5;1me\x1b[0mdge cell colo\x1b[38;5;208mr!\x1b[0m"},
	{"color at the end of a row clipped 5x7", []string{"edge cell color!", "edge cell color!", "no color at end."}, []string{"0000000000000004", "0000000000000044", "4000000000000000"}, []string{"0000000000000001", "1000000000000000", "0000000000000000"}, 5, 7, "\x1b[1;1Hedge \x1b[2;1H\x1b[48;5;1me\x1b[0mdge \x1b[3;1H\x1b[38;5;208mn\x1b[0mo co"},
	{"color at the end of a row clipped 20x20", []string{"edge cell color!", "edge cell color!", "no color at end."}, []string{"0000000000000004", "0000000000000044", "4000000000000000"}, []string{"0000000000000001", "1000000000000000", "0000000000000000"}, 20, 20, "\x1b[1;1Hedge cell color\x1b[38;5;208;48;5;1m!\x1b[0m\x1b[2;1H\x1b[48;5;1me\x1b[0mdge cell colo\x1b[38;5;208mr!\x1b[0m\x1b[3;1H\x1b[38;5;208mn\x1b[0mo color at end."},
	{"box drawing and blocks", []string{"┌──────────────┐", "│ ▀▄▀▄ ░▒▓█ ▄▀ │", "└──────────────┘"}, []string{"4444444444444444", "4033330222200004", "4444444444444444"}, []string{"0000000000000000", "0000004444000000", "0000000000000000"}, 0, 0, "\x1b[1;1H\x1b[38;5;208m┌──────────────┐\x1b[0m\x1b[2;1H\x1b[38;5;208m│\x1b[0m \x1b[38;5;16m▀▄▀▄\x1b[39;48;5;208m \x1b[38;5;4m░▒▓\x1b[49m█\x1b[0m ▄▀ \x1b[38;5;208m│\x1b[0m\x1b[3;1H\x1b[38;5;208m└──────────────┘\x1b[0m"},
	{"box drawing and blocks clipped 7x3", []string{"┌──────────────┐", "│ ▀▄▀▄ ░▒▓█ ▄▀ │", "└──────────────┘"}, []string{"4444444444444444", "4033330222200004", "4444444444444444"}, []string{"0000000000000000", "0000004444000000", "0000000000000000"}, 7, 3, "\x1b[1;1H\x1b[38;5;208m┌──────\x1b[0m\x1b[2;1H\x1b[38;5;208m│\x1b[0m \x1b[38;5;16m▀▄▀▄\x1b[39;48;5;208m \x1b[0m\x1b[3;1H\x1b[38;5;208m└──────\x1b[0m"},
	{"box drawing and blocks clipped 16x2", []string{"┌──────────────┐", "│ ▀▄▀▄ ░▒▓█ ▄▀ │", "└──────────────┘"}, []string{"4444444444444444", "4033330222200004", "4444444444444444"}, []string{"0000000000000000", "0000004444000000", "0000000000000000"}, 16, 2, "\x1b[1;1H\x1b[38;5;208m┌──────────────┐\x1b[0m\x1b[2;1H\x1b[38;5;208m│\x1b[0m \x1b[38;5;16m▀▄▀▄\x1b[39;48;5;208m \x1b[38;5;4m░▒▓\x1b[49m█\x1b[0m ▄▀ \x1b[38;5;208m│\x1b[0m"},
	{"box drawing and blocks clipped 5x7", []string{"┌──────────────┐", "│ ▀▄▀▄ ░▒▓█ ▄▀ │", "└──────────────┘"}, []string{"4444444444444444", "4033330222200004", "4444444444444444"}, []string{"0000000000000000", "0000004444000000", "0000000000000000"}, 5, 7, "\x1b[1;1H\x1b[38;5;208m┌────\x1b[0m\x1b[2;1H\x1b[38;5;208m│\x1b[0m \x1b[38;5;16m▀▄▀\x1b[0m\x1b[3;1H\x1b[38;5;208m└────\x1b[0m"},
	{"box drawing and blocks clipped 20x20", []string{"┌──────────────┐", "│ ▀▄▀▄ ░▒▓█ ▄▀ │", "└──────────────┘"}, []string{"4444444444444444", "4033330222200004", "4444444444444444"}, []string{"0000000000000000", "0000004444000000", "0000000000000000"}, 20, 20, "\x1b[1;1H\x1b[38;5;208m┌──────────────┐\x1b[0m\x1b[2;1H\x1b[38;5;208m│\x1b[0m \x1b[38;5;16m▀▄▀▄\x1b[39;48;5;208m \x1b[38;5;4m░▒▓\x1b[49m█\x1b[0m ▄▀ \x1b[38;5;208m│\x1b[0m\x1b[3;1H\x1b[38;5;208m└──────────────┘\x1b[0m"},
	{"black is not the default", []string{"black or default", "zero is default.", "16 is black.    "}, []string{"0000333300003333", "0000000000000000", "3333333333333333"}, []string{"0000000000000000", "3333333333333333", "0000000000000000"}, 0, 0, "\x1b[1;1Hblac\x1b[38;5;16mk or\x1b[0m def\x1b[38;5;16mault\x1b[0m\x1b[2;1H\x1b[48;5;16mzero is default.\x1b[0m\x1b[3;1H\x1b[38;5;16m16 is black.    \x1b[0m"},
	{"black is not the default clipped 7x3", []string{"black or default", "zero is default.", "16 is black.    "}, []string{"0000333300003333", "0000000000000000", "3333333333333333"}, []string{"0000000000000000", "3333333333333333", "0000000000000000"}, 7, 3, "\x1b[1;1Hblac\x1b[38;5;16mk o\x1b[0m\x1b[2;1H\x1b[48;5;16mzero is\x1b[0m\x1b[3;1H\x1b[38;5;16m16 is b\x1b[0m"},
	{"black is not the default clipped 16x2", []string{"black or default", "zero is default.", "16 is black.    "}, []string{"0000333300003333", "0000000000000000", "3333333333333333"}, []string{"0000000000000000", "3333333333333333", "0000000000000000"}, 16, 2, "\x1b[1;1Hblac\x1b[38;5;16mk or\x1b[0m def\x1b[38;5;16mault\x1b[0m\x1b[2;1H\x1b[48;5;16mzero is default.\x1b[0m"},
	{"black is not the default clipped 5x7", []string{"black or default", "zero is default.", "16 is black.    "}, []string{"0000333300003333", "0000000000000000", "3333333333333333"}, []string{"0000000000000000", "3333333333333333", "0000000000000000"}, 5, 7, "\x1b[1;1Hblac\x1b[38;5;16mk\x1b[0m\x1b[2;1H\x1b[48;5;16mzero \x1b[0m\x1b[3;1H\x1b[38;5;16m16 is\x1b[0m"},
	{"black is not the default clipped 20x20", []string{"black or default", "zero is default.", "16 is black.    "}, []string{"0000333300003333", "0000000000000000", "3333333333333333"}, []string{"0000000000000000", "3333333333333333", "0000000000000000"}, 20, 20, "\x1b[1;1Hblac\x1b[38;5;16mk or\x1b[0m def\x1b[38;5;16mault\x1b[0m\x1b[2;1H\x1b[48;5;16mzero is default.\x1b[0m\x1b[3;1H\x1b[38;5;16m16 is black.    \x1b[0m"},
}

func TestPaletteFramesWriteTheSameBytesAsBeforeTheRedesign(t *testing.T) {
	for _, tt := range goldenFrames {
		t.Run(tt.name, func(t *testing.T) {
			fb, err := NewFrameBuffer(16, len(tt.chars))
			if err != nil {
				t.Fatal(err)
			}
			for y := range tt.chars {
				for x, r := range []rune(tt.chars[y]) {
					fb.Set(x, y, Cell{Character: r, Foreground: goldenPalette[tt.fg[y][x]-'0'], Background: goldenPalette[tt.bg[y][x]-'0']})
				}
			}

			var buf bytes.Buffer
			if tt.clipW == 0 {
				err = fb.present(&buf)
			} else {
				err = fb.presentClipped(&buf, tt.clipW, tt.clipH)
			}
			if err != nil {
				t.Fatal(err)
			}
			if buf.String() != tt.want {
				t.Errorf("wanted %q, got %q", tt.want, buf.String())
			}
		})
	}
}

type goldenCanvas struct {
	name     string
	w, h     int
	seamless bool
	pixels   []string // a digit per pixel, into goldenCanvasPalette
	want     string
}

var goldenCanvases = []goldenCanvas{
	{"canvas half blocks", 8, 3, false, []string{"01000230", "02000310", "10300004", "00200430", "44000011", "00330000"}, "\x1b[1;1H \x1b[38;5;16;48;5;196m▀\x1b[0m   \x1b[38;5;196;48;5;21m▀\x1b[38;5;21;48;5;16m▀\x1b[0m \x1b[2;1H\x1b[38;5;16m▀\x1b[0m \x1b[38;5;21;48;5;196m▀\x1b[0m  \x1b[38;5;46m▄\x1b[38;5;21m▄\x1b[38;5;46m▀\x1b[0m\x1b[3;1H\x1b[38;5;46m▀▀\x1b[38;5;21m▄▄\x1b[0m  \x1b[38;5;16m▀▀\x1b[0m"},
	{"canvas solid cells", 8, 3, true, []string{"01000230", "02000310", "10300004", "00200430", "44000011", "00330000"}, "\x1b[1;1H \x1b[48;5;16m \x1b[0m   \x1b[48;5;196m \x1b[48;5;21m \x1b[0m \x1b[2;1H\x1b[48;5;16m \x1b[0m \x1b[48;5;21m \x1b[0m  \x1b[48;5;46m \x1b[48;5;21m \x1b[48;5;46m \x1b[0m\x1b[3;1H\x1b[48;5;46m  \x1b[48;5;21m  \x1b[0m  \x1b[48;5;16m  \x1b[0m"},
	{"canvas same color pairs", 6, 2, false, []string{"111222", "111222", "330044", "330044"}, "\x1b[1;1H\x1b[48;5;16m   \x1b[48;5;196m   \x1b[0m\x1b[2;1H\x1b[48;5;21m  \x1b[0m  \x1b[48;5;46m  \x1b[0m"},
}

func TestPaletteCanvasesWriteTheSameBytesAsBeforeTheRedesign(t *testing.T) {
	for _, tt := range goldenCanvases {
		t.Run(tt.name, func(t *testing.T) {
			fb, err := NewFrameBuffer(tt.w, tt.h)
			if err != nil {
				t.Fatal(err)
			}
			c := fb.Canvas()
			c.Seamless = tt.seamless
			for y, row := range tt.pixels {
				for x := range row {
					c.Set(x, y, goldenCanvasPalette[row[x]-'0'])
				}
			}
			c.Compose()

			var buf bytes.Buffer
			if err := fb.present(&buf); err != nil {
				t.Fatal(err)
			}
			if buf.String() != tt.want {
				t.Errorf("wanted %q, got %q", tt.want, buf.String())
			}
		})
	}
}
