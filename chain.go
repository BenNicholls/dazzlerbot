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

func (s suffixes) output() string {
	var sb strings.Builder
	for suf, num := range s.sufs {
		if suf == ENDOFSENTENCE {
			sb.WriteString("<END>")
		} else {
			sb.WriteString(suf)
		}
		if num > 1 {
			sb.WriteString("(x")
			sb.WriteString(strconv.Itoa(num))
			sb.WriteString(")")
		}
		sb.WriteString(" ")
	}

	return sb.String()
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
	for line, err := br.ReadString('\n'); true; line, err = br.ReadString('\n') {
		if err != nil && err != io.EOF {
			break
		}
		
		if line = strings.TrimSpace(line); line != "" {
			lr := bufio.NewReader(strings.NewReader(line))
			p := make(prefix, c.prefixLen)
			for {
				var next string
				if _, f_err := fmt.Fscan(lr, &next); f_err != nil {
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

		if err == io.EOF {
			break
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

	//should never reach this. abort!
	return ENDOFSENTENCE
}

func (c *Chain) Generate(n int) string {
	return c.GenerateWithPrefix(n, make([]string, 0))
}

func (c *Chain) GenerateWithPrefix(n int, prefixWords []string) string {
	if len(c.chain) == 0 && len(c.singletons) == 0 {
		return "Error: could not generate nonsense, brain empty"
	}

	p := make(prefix, c.prefixLen)
	words := make([]string, 0)
	for _, word := range prefixWords {
		p.shift(word)
		words = append(words, word)
	}

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

// outputs the entire chain. WARNING: for large chains, this takes FOREVER unless you're outputting
// to a file
func (c *Chain) output(to_file bool) {
	var file *os.File
	if to_file {
		var ferr error
		file, ferr = os.Create("brain.txt")
		if ferr != nil {
			fmt.Println("Could not output brain: ", ferr)
			return
		}
		defer file.Close()

		fmt.Println("outputting brain to file 'brain.txt'")
	}

	var sb strings.Builder

	for pre, sufs := range c.chain {
		sb.WriteString("(" + strconv.Itoa(sufs.total) + ") ")
		if pre == " " {
			sb.WriteString("<START>")
		} else if strings.HasPrefix(pre, " ") {
			sb.WriteString("<START>" + pre)
		} else {
			sb.WriteString(pre)
		}

		sb.WriteString(": " + sufs.output() + "\n")

		if !to_file {
			fmt.Print(sb.String())
			sb.Reset()
		}
	}

	for pre, suf := range c.singletons {
		sb.WriteString("(1) ")
		if pre == " " {
			sb.WriteString("<START>")
		} else if strings.HasPrefix(pre, " ") {
			sb.WriteString("<START>" + pre)
		} else {
			sb.WriteString(pre)
		}

		sb.WriteString(": ")
		if suf == ENDOFSENTENCE {
			sb.WriteString("<END>")
		} else {
			sb.WriteString(suf)
		}
		sb.WriteString("\n")

		if !to_file {
			fmt.Print(sb.String())
			sb.Reset()
		}
	}

	if to_file {
		file.WriteString(sb.String())
		fmt.Println("output complete!")
	}
}

// computes and outputs stats for the chain
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
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}

	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}

	sr := strings.NewReader(s)
	c.Build(sr)
}
