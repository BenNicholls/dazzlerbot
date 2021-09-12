package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const ENDOFSENTENCE string = " "

type prefix []string

func (p prefix) key() string {
	return strings.Join(p, " ")
}

func (p prefix) shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = strings.ToLower(strings.Trim(word, ",\"':()+-"))
}

type suffixes struct {
	sufs  map[string]int //key is the word, value is the weight of that word
	total int            //total suffixes recorded (weighted sum of suffixes)
}

func (s *suffixes) add(word string) {
	s.total += 1
	if _, ok := s.sufs[word]; !ok {
		s.sufs[word] = 1
	} else {
		s.sufs[word] = s.sufs[word] + 1
	}
}

func (s suffixes) rand() string {
	n := rand.Intn(s.total)
	for s, num := range s.sufs {
		n -= num
		if n <= 0 {
			return s
		}
	}

	//should never reach this.
	return "[ERROR: could not randomize good??]"
}

func (s suffixes) output() {
	for suf, num := range s.sufs {
		if suf == ENDOFSENTENCE {
			fmt.Print("<END>")
		} else {
			fmt.Print(suf)
		}
		if num > 1 {
			fmt.Print(" (x" + strconv.Itoa(num) + ")")
		}
		fmt.Print(" ")
	}
}

type Chain struct {
	chain      map[string]suffixes //map of prefix to multiple weighted suffixes
	singletons map[string]string   //map of prefix to suffix for prefixes with only 1 suffix
	prefixLen  int
}

func (c *Chain) init() {
	c.chain = make(map[string]suffixes)
	c.singletons = make(map[string]string)
	c.prefixLen = config.DefaultPrefixLen
}

func (c *Chain) Build(r io.Reader) {
	br := bufio.NewReader(r)
	for line, err := br.ReadString('\n'); err == nil; line, err = br.ReadString('\n') {
		lr := bufio.NewReader(strings.NewReader(line))
		p := make(prefix, c.prefixLen)
		for {
			var next string
			if _, err := fmt.Fscan(lr, &next); err != nil {
				next = ENDOFSENTENCE
			}

			if sufs, ok := c.chain[p.key()]; ok { //check chain
				sufs.add(next)
				c.chain[p.key()] = sufs
			} else if suf, ok := c.singletons[p.key()]; ok { //check singletons. if found, remove and add to chain
				delete(c.singletons, p.key())
				sufs := suffixes{
					sufs:  map[string]int{},
					total: 0,
				}
				sufs.sufs = make(map[string]int)
				sufs.add(suf)
				sufs.add(next)
				c.chain[p.key()] = sufs
			} else { //new prefix. add to singletons
				c.singletons[p.key()] = next
			}

			if next == ENDOFSENTENCE {
				break
			}
			p.shift(next)
		}
	}
}

func (c *Chain) GetNextWord(key string) string {
	//check chain
	if sufs, ok := c.chain[key]; ok {
		return sufs.rand()
	} else if suf, ok := c.singletons[key]; ok {
		return suf
	}

	//should never reach this
	return "[ERROR. INVALID PREFIX]"
}

func (c *Chain) Generate(n int) string {
	if len(c.chain) == 0 && len(c.singletons) == 0 {
		return "Error: could not generate nonsense, brain empty"
	}

	p := make(prefix, c.prefixLen)
	words := make([]string, 0)

	for i := 0; i < n; i++ {
		next := c.GetNextWord(p.key())
		if next == ENDOFSENTENCE {
			break
		}
		words = append(words, next)
		p.shift(next)
	}

	words = prettify(words)

	return strings.Join(words, " ")
}

//outputs the entire chain. WARNING: for large chains, this takes FOREVER.
func (c *Chain) output() {
	for pre, sufs := range c.chain {
		fmt.Print("(" + strconv.Itoa(sufs.total) + ") " + pre + ": ")
		sufs.output()
		fmt.Print("\n")
	}
	for pre, suf := range c.singletons {
		fmt.Println("(1) " + pre + ": " + suf)
	}
}

//computes and outputs stats for the chain
func (c *Chain) outputStats() {
	fmt.Println("Total prefixes: ", len(c.chain)+len(c.singletons))
	totalSufs := 0
	uniqueSufs := 0
	maxSufs := 0
	mostCommonPrefix := ""
	initSufs := 0
	for pre, sufs := range c.chain {
		totalSufs += sufs.total
		uniqueSufs += len(sufs.sufs)
		if sufs.total > maxSufs && pre != " " && !strings.HasPrefix(pre, " ") {
			maxSufs = sufs.total
			mostCommonPrefix = pre
		} else if pre == " " {
			initSufs = len(sufs.sufs)
		}
	}
	totalSufs += len(c.singletons)
	uniqueSufs += len(c.singletons)
	fmt.Println("Unique Suffixes: ", uniqueSufs)
	fmt.Println("Total Suffixes: ", totalSufs)
	fmt.Println("Complexity factor: ", float64(uniqueSufs)/float64(len(c.chain)+len(c.singletons)))
	fmt.Println("Single response prefixes: ", len(c.singletons))
	fmt.Println("Start of sentence suffixes: ", initSufs)
	fmt.Println("Most common prefix: ", mostCommonPrefix, "("+strconv.Itoa(maxSufs)+" suffixes total)")
}

func (c *Chain) inputTextFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return errors.New("Could not open archive file:" + err.Error())
	}
	defer file.Close()

	fmt.Println("loading", filePath)
	c.Build(file)

	return nil
}

func (c *Chain) AddString(s string) {
	if s == "" {
		return
	}

	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}

	sr := strings.NewReader(s)
	c.Build(sr)
}
