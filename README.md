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
prod/
admin.py
app/login.html

```
```shell script
> go run mkpath.go -d example.com -l 2 -w wordlist.txt
example.com/dev
example.com/prod
example.com/dev/dev
example.com/prod/dev
example.com/dev/prod
example.com/prod/prod
example.com/dev/admin.py
example.com/dev/app/login.html
example.com/prod/admin.py
example.com/prod/app/login.html
example.com/dev/dev/admin.py
example.com/dev/dev/app/login.html
example.com/prod/dev/admin.py
example.com/prod/dev/app/login.html
example.com/dev/prod/admin.py
example.com/dev/prod/app/login.html
example.com/prod/prod/admin.py
example.com/prod/prod/app/login.html

```
