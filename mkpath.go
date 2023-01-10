package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"

	roundChan "github.com/trickest/mkpath/round"
)

const (
	fileRegex = "[^?*:;{}]+\\.[^/?*:;{}]+"

	bufferSizeMB      = 100
	maxWorkingThreads = 100000
	numberOfFiles     = 1
)

var (
	domain       string
	inputDomains []string
	domainFile   string

	wordlist    string
	toLowercase bool
	regex       string
	dirWordSet  map[string]bool
	fileWordSet map[string]bool

	depth          int
	onlyDirs       bool
	onlyFiles      bool
	outputFileName string
	silent         bool

	workers         int
	workerThreadMax = make(chan struct{}, maxWorkingThreads)
	done            = make(chan struct{})
	wg              sync.WaitGroup
	wgWrite         sync.WaitGroup
	robin           roundChan.RoundRobin
)

func readDomainFile() {
	inputFile, err := os.Open(domainFile)
	if err != nil {
		fmt.Println("Could not open file to read domains:", err)
		os.Exit(1)
	}
	defer inputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		inputDomains = append(inputDomains, strings.TrimSpace(scanner.Text()))
	}
}

func prepareDomains() {
	if domain == "" && domainFile == "" {
		fmt.Println("No domain input provided!")
		os.Exit(1)
	}

	inputDomains = make([]string, 0)
	if domain != "" {
		inputDomains = append(inputDomains, domain)
	} else {
		if domainFile != "" {
			readDomainFile()
		}
	}
}

func readWordlistFile() {
	var reg *regexp.Regexp
	var err error
	if regex != "" {
		reg, err = regexp.Compile(regex)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	wordlistFile, err := os.Open(wordlist)
	if err != nil {
		fmt.Println("Could not open file to read wordlist:", err)
		os.Exit(1)
	}
	defer wordlistFile.Close()

	fileReg, err := regexp.Compile(fileRegex)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dirWordSet = make(map[string]bool)
	fileWordSet = make(map[string]bool)
	scanner := bufio.NewScanner(wordlistFile)

	for scanner.Scan() {
		word := scanner.Text()
		if toLowercase {
			word = strings.ToLower(word)
		}
		word = strings.Trim(word, "/")
		if word != "" {
			if reg != nil {
				if !reg.Match([]byte(word)) {
					continue
				}
			}
			if fileReg.Match([]byte(word)) {
				fileWordSet[word] = true
			} else {
				dirWordSet[word] = true
			}
		}
	}
}

func closeWriters(number int) {
	for i := 0; i < number; i++ {
		done <- struct{}{}
	}
}

func spawnWriters(number int) {
	for i := 0; i < number; i++ {
		var bf bytes.Buffer
		ch := make(chan string, 100000)

		fileName := outputFileName
		fileSplit := strings.Split(fileName, ".")
		if len(fileSplit) == 1 {
			fileName += ".txt"
		}
		if number > 1 {
			fileSplit = strings.Split(fileName, ".")
			extension := "." + fileSplit[len(fileSplit)-1]
			fileName = strings.TrimSuffix(fileName, extension) + "-" + strconv.Itoa(i) + extension
		}
		file, err := os.Create(fileName)
		if err != nil {
			fmt.Println("Couldn't open file to write output:", err)
			os.Exit(1)
		}

		wgWrite.Add(1)
		go write(file, &bf, &ch)

		if robin == nil {
			robin = roundChan.New(&ch)
			continue
		}
		robin.Add(&ch)
	}
}

func write(file *os.File, buffer *bytes.Buffer, ch *chan string) {
mainLoop:
	for {
		select {
		case <-done:
			for {
				if !writeOut(file, buffer, ch) {
					break
				}
			}
			if buffer.Len() > 0 {
				if file != nil {
					_, _ = file.WriteString(buffer.String())
					buffer.Reset()
				}
			}
			break mainLoop
		default:
			writeOut(file, buffer, ch)
		}
	}
	wgWrite.Done()
}

func writeOut(file *os.File, buffer *bytes.Buffer, outputChannel *chan string) bool {
	select {
	case s := <-*outputChannel:
		buffer.WriteString(s)
		if buffer.Len() >= bufferSizeMB*1024*1024 {
			_, _ = file.WriteString(buffer.String())
			buffer.Reset()
		}
		return true
	default:
		return false
	}
}

func combo(_comb string, level int, wg *sync.WaitGroup, wt *chan struct{}) {
	defer wg.Done()
	workerThreadMax <- struct{}{}

	if strings.Count(_comb, "/") > 0 {
		processOutput(_comb, robin.Next())
	}

	var nextLevelWaitGroup sync.WaitGroup
	if level > 1 {
		nextLevelWt := make(chan struct{}, workers)
		for dw := range dirWordSet {
			nextLevelWaitGroup.Add(1)
			nextLevelWt <- struct{}{}
			go combo(_comb+"/"+dw, level-1, &nextLevelWaitGroup, &nextLevelWt)
		}
	} else {
		for dw := range dirWordSet {
			processOutput(_comb+"/"+dw, robin.Next())
		}
	}

	nextLevelWaitGroup.Wait()
	<-workerThreadMax
	<-*wt
}

func processOutput(out string, outChan *chan string) {
	if onlyDirs || onlyFiles == onlyDirs {
		if !silent {
			fmt.Print(out + "\n")
		}
		*outChan <- out + "\n"
	}

	if onlyFiles || onlyFiles == onlyDirs {
		for file := range fileWordSet {
			if !silent {
				fmt.Print(out + "/" + file + "\n")
			}
			*outChan <- out + "/" + file + "\n"
		}
	}
}

func main() {
	flag.StringVar(&domain, "d", "", "Input domain")
	flag.StringVar(&domainFile, "df", "", "Input domain file, one domain per line")
	flag.StringVar(&wordlist, "w", "", "Wordlist file")
	flag.BoolVar(&toLowercase, "lower", false, "Convert wordlist file content to lowercase (default false)")
	flag.StringVar(&regex, "r", "", "Regex to filter words from wordlist file")
	flag.IntVar(&depth, "l", 1, "URL path depth to generate")
	flag.StringVar(&outputFileName, "o", "", "Output file (optional)")
	flag.BoolVar(&onlyDirs, "only-dirs", false, "Generate directories only, files are filtered out (default false)")
	flag.BoolVar(&onlyFiles, "only-files", false, "Generate files only, file names are appended to given domains (default false)")
	flag.IntVar(&workers, "t", 100, "Number of threads for every path depth")
	flag.BoolVar(&silent, "silent", true, "Skip writing generated paths to stdout (faster)")
	flag.Parse()

	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
		<-signalChannel

		fmt.Println("Program interrupted, exiting...")
		os.Exit(0)
	}()

	if depth <= 0 || workers <= 0 {
		fmt.Println("Path depth and number of threads must be positive integers!")
		os.Exit(0)
	}

	prepareDomains()
	readWordlistFile()
	spawnWriters(numberOfFiles)

	if outputFileName == "" {
		silent = false
	}

	for _, d := range inputDomains {
		wg.Add(1)
		wt := make(chan struct{}, 1)
		wt <- struct{}{}
		go combo(d, depth, &wg, &wt)
	}

	wg.Wait()
	closeWriters(numberOfFiles)
	wgWrite.Wait()
}
