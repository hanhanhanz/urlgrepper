# urlgrepper


Tools for grepping URL or subdomain from webpage with Golang Concurrency utilized. 


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
  -r int
        max times of redirect (minimum 0) (default 2)
  -t int
        goroutine number to be utilized (kinda like thread)
  -tx string
        specify extract domain if it is different with source URL requested
  -u string
        single source where url will be taken
  -ul string
        source where url will be taken in a file
  -v    enable verbose mode
  -x string
        xtension to extract (only work for URL mode)


```



### Example
```sh
$ go run urlgrepper.go -u https://tokopedia.com
iteration 0, total domain : 14
iteration 1, total domain : 14
iteration 2, total domain : 18
iteration 3, total domain : 18
final result :
gql.tokopedia.com
hub.tokopedia.com
www.tokopedia.com
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
mojito.tokopedia.com
m.tokopedia.com
api.tokopedia.com
js.tokopedia.com
kamus.tokopedia.com

```

