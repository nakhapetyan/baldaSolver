package main

import (
	"testing"
)

func BenchmarkFind(b *testing.B) {

	tree := loadDict("test.dict")

	board := [][]string{
		{"","","","",""},
		{"","","","",""},
		{"k","i","t","t","y"},
		{"","","","",""},
		{"","","","",""},
	}


	var p FindParams
	p.exclude += " test "
	p.forecast = 0
	p.standart = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		find(board, tree, &p)
	}
}