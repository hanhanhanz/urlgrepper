# urlgrepper


Tools for grepping URL or subdomain with Golang Concurrency utilized


### Requirement

urlgrepper uses a [go-retryablehttp](github.com/hashicorp/go-retryablehttp") to work properly

### Installation

```sh
$ go build urlgrepper.go
$ ./urlgrepper
```



### Usage

```sh
Usage of urlgrepper:
  -m string
        choose what to extract (domain,url) (default "domain")
  -o string
        output result to a file
  -t int
        goroutine number to be utilized (kinda like thread) (in development) (default 10)
  -tx string
        specify extact domain if it is different with source URL requested (in development)
  -u string
        single source where url will be taken
  -ul string
        source where url will be taken in a file
  -x string
        xtension to extract (only work for URL mode)

```



### Example
```sh
$ go run urlgrepper.go -u https://tokopedia.com
2020/12/22 05:32:09 [DEBUG] GET https://tokopedia.com
gql.tokopedia.com
hub.tokopedia.com
www.tokopedia.com
m.tokopedia.com
accounts.tokopedia.com
chat.tokopedia.com
seller.tokopedia.com
ta.tokopedia.com
ace.tokopedia.com
pulsa.tokopedia.com
tiket.tokopedia.com
pay.tokopedia.com
gw.tokopedia.com
goldmerchant.tokopedia.com

```sh

