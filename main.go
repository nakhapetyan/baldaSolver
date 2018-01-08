package main

import (
	"fmt"
	"net/http"
	"os"
	"bufio"
	"log"
	"flag"
	"strconv"
	"time"
	"sort"
	"io"
	"strings"
)


type Result struct {
	word string
	char string
	b byte
	points int
	x int
	y int
	forecast string
}

type FindParams struct {
	strBoard [][]string
	board [][]byte
	boardCounter []int
	initBoardCounter []int
	x1 int
	y1 int
	w int
	standart bool
	diagonally bool
	tree *Tree
	result []Result
	exclude string
	forecast int
}

type Tree struct {
	next map[byte]*Tree
	str string
	wrd string
	tree *Tree
}

func NewTree() *Tree {
	var o Tree
	o.next = make(map[byte]*Tree)
	return &o
}

// https://habrahabr.ru/post/207734/
func addWord(tree *Tree, treeInv *Tree, word string) {
	cur := tree
	last := len(word)-1
	letters := make(map[int]*Tree)
	// fill words tree
	// cat
	for i := 0; i <= last; i++ {
		if cur.next[word[i]] == nil {
			cur.next[word[i]] = NewTree()
		}
		letters[i] = cur
		cur = cur.next[word[i]]
		if i == last {
			cur.wrd = word
		}
	}

	// fill inverted prefix
	// cat -> tac, ac, c
	for last := len(word)-1; last >= 0; last-- { //!!!
		cur = treeInv
		for i := last; i >= 0; i-- {
			if cur.next[word[i]] == nil {
				cur.next[word[i]] = NewTree()
			}
			cur = cur.next[word[i]]
			if i == 0 {
				cur.tree = letters[last]
			}
		}
	}
}

func printTree(tree *Tree, indent string) string {
	s := ""
	//s += "{\n"
	//s += "next = [\n"
	s += fmt.Sprintf("(%p) next = [\n", tree)

	for k,v := range tree.next {
		s += "\t" + indent + fmt.Sprintf("%s = %s\n", string(k), printTree(v, indent + "\t"))
	}
	s += indent + "]\n"
	if tree.wrd != "" {
		s += indent + fmt.Sprintf("wrd = %s\n", tree.wrd)
	}
	if tree.tree != nil {
		s += indent + fmt.Sprintf("tree = %p\n", tree.tree)
	}
	//s += indent + "}\n"
	return s
}

func check(x int, y int, tree *Tree, p *FindParams, str string) bool {
	if x >= p.w || x < 0 || y >= p.w || y < 0 {
		return false
	}

	//str += fmt.Sprintf(" -> %s [%d,%d] %d", string(p.board[y][x]), x,y, p.boardCounter[y*p.w+x])

	next := tree.next[p.board[y][x]]
	if next == nil {
		//str += " NO next"
		//fmt.Println("Go", str)
		return false
	}

	index := y*p.w+x
	if p.boardCounter[index] <= 0 {
		//str += " NO count "
		//fmt.Println("Go", str)
		return false
	}
	p.boardCounter[index] -= 1
	////fmt.Println(fmt.Sprintf("-- [%d,%d] %d", x,y,p.boardCounter[index]))

	if next.wrd != "" && !strings.Contains(p.exclude, " "+next.wrd+" ") {
		// find word
		//str += " FIND " + next.wrd
		//fmt.Println("Go", str)
		//fmt.Sprintf("%s %d %d %s", next.wrd, p.x1, p.y1, string(p.board[p.y1][p.x1]))
		//fmt.Println(p)
		p.result = append(p.result, Result{ next.wrd, string(p.board[p.y1][p.x1]), p.board[p.y1][p.x1],  len(next.wrd), p.x1, p.y1, "" })
	}

	if next.tree != nil {
		// find prefix, go to word
		p.boardCounter[p.y1*p.w+p.x1] += 1
		check(p.x1,p.y1, next.tree, p, str + fmt.Sprintf(" Go to tree (%p) ", next.tree))
		p.boardCounter[p.y1*p.w+p.x1] -= 1
	}

	//fmt.Println("Go", str)

	if len(tree.next) == 0 {
		return false
	}

	if p.standart {
		check(x, y + 1, next, p, str)
		check(x, y - 1, next, p, str)
		check(x + 1, y, next, p, str)
		check(x - 1, y, next, p, str)
	}


	if p.diagonally {
		check(x+1, y-1, next, p, str)
		check(x+1, y+1, next, p, str)
		check(x-1, y-1, next, p, str)
		check(x-1, y+1, next, p, str)
	}


	p.boardCounter[index] += 1
	////fmt.Println(fmt.Sprintf("++ [%d,%d] %d", x,y,p.boardCounter[index]))


	return true
}

func loadDict(dictFile string) *Tree {
	tree := NewTree()
	treeInv := NewTree()

	// read file by line
	file, err := os.Open(dictFile)
	if err != nil {
		log.Print(err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 9 {
			continue
		}
		line = line[3:len(line)-9]
		//fmt.Println(line)
		addWord(tree, treeInv, line)
	}
	//fmt.Println(printTree(treeInv, ""));
	//fmt.Println(printTree(tree, ""));

	return treeInv
}


func find(board [][]string, tree *Tree, p *FindParams) {
	p.w = len(board[0])
	p.strBoard = board
	p.tree = tree

	p.initBoardCounter = make([]int, len(board)*p.w)
	p.boardCounter = make([]int, len(board)*p.w)
	for i := range p.initBoardCounter {
		p.initBoardCounter[i] = 1
	}

	// convert to byte array
	if len(p.board) <= 0 {
		for y, row := range board {
			p.board = append(p.board, []byte{})
			for x, char := range row {
				p.board[y] = append(p.board[y], 0)
				if char != "" {
					p.board[y][x] = char[0]
				}
			}
		}
	}

	for y,row := range p.board {
		for x,boardChar := range row {
			if boardChar == 0 {
				for char, _ := range p.tree.next {
					copy(p.boardCounter, p.initBoardCounter);

					p.board[y][x] = char
					p.x1 = x
					p.y1 = y

					check(x, y, tree, p, "")
					//fmt.Println("Try", x,y)
				}
				p.board[y][x] = boardChar
			}
		}
	}


	sort.SliceStable(p.result, func(i, j int) bool {
		s1 := p.result[i].points
		s2 := p.result[j].points

		if s1 == s2 {
			return p.result[i].word < p.result[j].word
		}
		return s1 > s2
	})


	if p.forecast > 0 {
		p.forecast -= 1;
		maxLen := 0
		for k,v := range p.result {
			if maxLen == 0 {
				maxLen = v.points
			}

			if k > 3 && v.points < maxLen {
				break
			}

			// emulate opponent turn
			//subParams := *p
			//subParams.result = []Result{}
			//subParams.board = [][]byte{}

			var subParams FindParams
			subParams.forecast = p.forecast
			subParams.standart = p.standart
			subParams.diagonally = p.diagonally
			subParams.exclude = p.exclude+" "+v.word+" "
			subParams.board = p.board
			subParams.board[v.y][v.x] = v.b

			find(board, tree, &subParams)
			subParams.board[v.y][v.x] = 0

			for _,v2 := range subParams.result {
				p.result[k].forecast += v2.word + " "
				if p.forecast % 2 == 0 {
					p.result[k].points += (p.result[k].points - v2.points)
				} else {
					p.result[k].points -= (p.result[k].points - v2.points)
				}
				break;
			}



		}
	}



	//p.result = p.result[0 : int(math.Min(float64(len(p.result)), 50.0))] //

}

func resultsToJson(p *FindParams) string {
	//if len(p.result) == 0 {
	//	return "[]"
	//}
	json := ""
	for _,v := range p.result {
		json += fmt.Sprintf(", { \"word\":\"%s\", \"cell\":%d, \"char\":\"%d\", \"forecast\":\"%s\" }", v.word, v.y * p.w + v.x, v.b, v.forecast)
	}
	return fmt.Sprintf("[%s]\n", json[1:])
}


func main(){
	httpPort := flag.Int("http", 0, "the http port")
	dictFile := flag.String("dict", "test.dict", "")
	flag.Parse()

	// start web server
	if *httpPort > 0 {
		trees := make(map[string]*Tree)

		http.HandleFunc("/get", func (res http.ResponseWriter, req *http.Request) {
			res.Header().Set("Content-Type", "application/json")
			log.Println(req.RequestURI)
			keys := req.URL.Query()

			board := [][]string{}
			for i := 0; i < 20; i++ {
				key := "a["+strconv.Itoa(i)+"][]"
				if keys[key] != nil {
					board = append(board, keys[key])
				}
			}
			if len(board) <= 0 {
				return
			}

			dict := "/tmp/balda_words2_en.php"
			if keys["dict"] != nil {
				dict = keys["dict"][0]
			}
			if trees[dict] == nil {
				trees[dict] = loadDict(dict)
			}
			if trees[dict] == nil {
				return
			}
			tree := trees[dict]


			var p FindParams
			p.exclude = keys["exclude"][0]
			p.standart = keys["rules[standart]"] != nil
			p.diagonally = keys["rules[diagonally]"] != nil
			p.forecast = 0
			find(board, tree, &p)

			io.WriteString(res, resultsToJson(&p))
		})
		http.ListenAndServe(":"+strconv.Itoa(*httpPort), nil)
		os.Exit(3)
	}

	// command line run
	tree := loadDict(*dictFile)

	board := [][]string{
		{"","",""},
		{"c","",""},
		{"","t","y"},
	}
	board = [][]string{
		{"z","z","z","z","z"},
		{"a","z","z","z","z"},
		{"b","","z","z","z"},
		{"z","z","z","z","z"},
		{"z","z","z","z","z"},
	}
	board = [][]string{
		{"","","","",""},
		{"","","","",""},
		{"k","i","t","t","y"},
		{"","","","a",""},
		{"","","","",""},
	}
	start := time.Now().UnixNano()
	var p FindParams
	p.exclude += " test "
	p.forecast = 0
	p.standart = true
	find(board, tree, &p)
	fmt.Println(resultsToJson(&p))

	end := time.Now().UnixNano()
	fmt.Printf("duration: %f", float64(end-start) / float64(time.Millisecond))

}