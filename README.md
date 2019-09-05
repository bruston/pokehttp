pokehttp
========

Check a list of subdomains for active webservers.

```
Usage of pokehttp:
    -a string
        user-agent header to use (default "pokehttp: https://github.com/bruston/pokehttp")
    -c uint
        number of concurrent quests (default 6)
    -ck string
        add a custom header to all requests (key)
    -cv string
        add a custom header to all requests (value)
    -d string
        file containing list of domains or ip addresses seperated by newlines
    -h string
        host header
    -k  ignore SSL errors
    -p string
        comma seperated list of ports to probe (default "80,433")
    -t uint
        timeout in seconds (default 5)
    -x string
        X-Forwarded-For header
```

# Example

```
pokehttp -d subdomains.txt -p="80,443,8080" | tee -a subdomains.web
```
