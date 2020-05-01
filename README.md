pokehttp
========

Check a list of subdomains for active webservers.

```
Usage of pokehttp:
    -a string
        user-agent header to use (default "pokehttp: https://github.com/bruston/pokehttp")
    -c uint
        number of concurrent requests (default 6)
    -d string
        file containing list of domains or ip addresses seperated by newlines, reads from stdin if not supplied
    -k  ignore SSL errors
    -p string
        comma seperated list of ports to probe (default "80,433")
    -t uint
        timeout in seconds (default 5)
    -H string
        add a header, eg: -H "Header-One: foo"
```

# Example

```
pokehttp -d subdomains.txt -p="80,443,8080" -H "Header-One: foo" -H "Header-Two: bar" | tee -a subdomains.web
```
```
