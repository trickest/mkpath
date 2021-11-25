mkpath
------
Make paths using a wordlist

Read a wordlist file and generate paths for given domain or list of domains.
Input from wordlist file is lowercased and unique words are processed. Additionally, wordlist can be
filtered using regex. 

```
Usage of mkpath:
  -d string
    	Input domain
  -df string
    	Input domain file, one domain per line
  -l int
    	Path depth to generate (default 1)
  -o string
    	Output file (optional)
  -r string
    	Regex to filter words from wordlist file
  -w string
    	Wordlist file
  -only-dirs
        Flag for generating directories only, files are being filtered out (default false)
  -only-files
        Flag for generating files only, files are being concatenated to given domains (default false)
        
        *If both only-dirs and only-files are set to true or false, the output will be complete,
        meaning all possible paths will be generated, as well as those paths with all files appended*
```

### Example

##### wordlist.txt
```
dev
DEV
*
foo/bar
prod
```
```shell script
> go run mksub.go -d example.com -l 2 -w input.txt -r "^[a-zA-Z0-9\.-_]+$"
example.com/dev/
example.com/foo/bar/
example.com/prod/
example.com/foo/bar/dev/
example.com/prod/dev/
example.com/dev/dev/
example.com/dev/foo/bar/
example.com/foo/bar/foo/bar/
example.com/prod/foo/bar/
example.com/dev/prod/
example.com/foo/bar/prod/
example.com/prod/prod/

```
