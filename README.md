# Wappalyzer CLI

CLI app based upon [projectdiscovery/wappalyzergo](https://github.com/projectdiscovery/wappalyzergo).


## Installation
```bash
go install -v github.com/boogah/wappalyzer-cli@main
```

## Usage
```bash
wappalyzer-cli --target https://littleroom.studio/ --disable-ssl --output output.txt --json
```
Example output:
```json
{
  "Google Web Server":{},
  "HSTS":{},
  "HTTP/3":{}
}
```

### Available arguments
```bash
Usage of wappalyzer:
  --target string  Target to analyze
  --output string  Output file
  --method string  Request method (default "GET")
  --header value   Set additional request headers
  --disable-ssl    Don't verify the site's SSL certificate
  --json           Json output format
  --no-color       Disable color output
  --silent         Don't display any output
  --version        Show version and exit
```


## Build
```bash
git clone https://github.com/boogah/wappalyzer-cli
cd wappalyzer-cli
go build -a -ldflags "-w -s -X main.buildNumber=1 -X main.buildVersion=custom" -o wappalyzer-cli
```


## Security
If you discover any security related issues, please email `boogah` on Gmail instead of using the issue tracker.


## Credits
- [boogah][link-author]
- [projectdiscovery/wappalyzergo](https://github.com/projectdiscovery/wappalyzergo)
- [All Contributors][link-contributors]


## License
The MIT License (MIT). Please see [License File](LICENSE.md) for more information.


[link-author]: https://github.com/boogah
[link-contributors]: https://github.com/boogah/wappalyzer-cli/graphs/contributors