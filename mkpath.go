package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	fileRegex = "[^?*:;{}]+\\.[^/?*:;{}]+"
)

func generatePaths(wordSet map[string]bool, depth int) []string {
	results := make([]string, 0)
	for i := 0; i < depth; i += 1 {
		toMerge := results[0:]
		if len(toMerge) == 0 {
			for word := range wordSet {
				results = append(results, word)
			}
		} else {
			for _, sd := range toMerge {
				for word := range wordSet {
					results = append(results, fmt.Sprintf("%s/%s", word, sd))
				}
			}
		}
	}
	return results
}

func generateFiles(paths []string, files map[string]bool) []string {
	results := make([]string, 0)
	for _, path := range paths {
		for file := range files {
			results = append(results, path+"/"+file)
		}
	}
	return results
}

func generateAll(paths map[string]bool, files map[string]bool, depth int) []string {
	var results []string
	results = generatePaths(paths, depth)
	results = append(results, generateFiles(generatePaths(paths, depth), files)...)
	return results
}

func main() {
	domain := flag.String("d", "", "Input domain")
	domainFile := flag.String("df", "", "Input domain file, one domain per line")
	wordlist := flag.String("w", "", "Wordlist file")
	r := flag.String("r", "", "Regex to filter words from wordlist file")
	depth := flag.Int("l", 1, "URL path depth to generate (default 1)")
	output := flag.String("o", "", "Output file (optional)")
	onlyDirs := flag.Bool("only-dirs", false, "Flag for generating directories only, files are being filtered out (default false)")
	onlyFiles := flag.Bool("only-files", false, "Flag for generating files only, files are being concatenated to given domains (default false)")
	flag.Parse()

	inputDomains := make([]string, 0)
	if *domain != "" {
		inputDomains = append(inputDomains, *domain)
	}
	if *domainFile != "" {
		inputFile, err := os.Open(*domainFile)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		defer inputFile.Close()
		scanner := bufio.NewScanner(inputFile)
		for scanner.Scan() {
			inputDomains = append(inputDomains, scanner.Text())
		}
	}
	if len(inputDomains) == 0 {
		fmt.Println("No input provided")
		os.Exit(1)
	}

	wordlistFile, err := os.Open(*wordlist)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer wordlistFile.Close()

	var reg *regexp.Regexp
	if *r != "" {
		reg, err = regexp.Compile(*r)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	var fileReg *regexp.Regexp
	fileReg, err = regexp.Compile(fileRegex)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var outputFile *os.File
	if *output != "" {
		outputFile, err = os.Create(*output)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		defer outputFile.Close()
	}

	dirWordSet := make(map[string]bool)
	fileWordSet := make(map[string]bool)
	scanner := bufio.NewScanner(wordlistFile)

	for scanner.Scan() {
		word := strings.ToLower(scanner.Text())
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

	var results []string
	if *onlyDirs != *onlyFiles {
		if *onlyDirs {
			results = generatePaths(dirWordSet, *depth)
		} else {
			results = generateFiles(results, fileWordSet)
		}
	} else {
		results = generateAll(dirWordSet, fileWordSet, *depth)
	}

	for _, domain := range inputDomains {
		for _, subpath := range results {
			fmt.Println(domain + "/" + subpath)
			if outputFile != nil {
				_, _ = outputFile.WriteString(domain + "/" + subpath + "\n")
			}
		}
	}
}
