package main
import "github.com/hashicorp/go-retryablehttp"
import "fmt"
import "os"
import "time"
import "io/ioutil"
import "regexp"
import "flag"
import "bufio"
import "net/url"
import "strings"
import "sync"
import "net/http"
import "crypto/tls"
import "syscall"
//import "errors"
//import "net"
//import "github.com/gijsbers/go-pcre"

type conf struct {
	Url string
	Urls string
	Mode string
	Outname string
	Thread int
	Xtension string
	Toxtract string
	Outfile *(os.File)
}

type error interface {
   Error() string
}

func storehere(d string, f *(os.File) ) { //store result (in string) to file
	if _, err := f.WriteString(d); err != nil {
		fmt.Printf("[-]storing function error :  ")
		panic(err)
	}	
}

func removeDuplicateValues(stringSlice []string) []string { 
    keys := make(map[string]bool) 
    list := []string{} 
  
    // If the key(values of the slice) is not equal 
    // to the already present value in new slice (list) 
    // then we append it. else we jump on another element. 
    for _, entry := range stringSlice { 
        if _, value := keys[entry]; !value { 
            keys[entry] = true
            list = append(list, entry) 
        } 
    } 
    return list 
} 

func errorCatch(err error, m string) {
	if err != nil {
			fmt.Println("[-]",m)
			panic(err)
		}	
}

func errorKill(m string) {
	fmt.Println("[-] ",m)
	os.Exit(3)
	
}

func urltoslice(url string, urls string) []string {
	var seeds = []string{}
	if url == "" && urls == "" {
		errorKill("-u or -ul is mandatory")
	} else if url != "" && urls != "" {
		errorKill("choose either -u or -ul")
	} else if url != "" {
		_, err := regexp.Match(`^https?://`, []byte(url))
		if err != nil {
	    	fmt.Printf("[-] regex.match failed  : ")
	    	panic(err)
	    }
		seeds = append(seeds,url)
		
	} else if urls != "" {
		//prepare to read file from -ul
		var g *os.File
    	var g2 *bufio.Scanner
    	g,_ = os.Open(urls) 
	    g2 = bufio.NewScanner(g)

    	for g2.Scan() {
    		var line = g2.Text()
    		_, err := regexp.Match(`^https?://`, []byte(url))
			if err != nil {
	    		fmt.Printf("[-] regex.match failed  : ")
	    		panic(err)
	    	}
			seeds = append(seeds,line)
			
    	}
	}
	return seeds
		
}


func myrequest(nc *retryablehttp.Client, seed string, wg *sync.WaitGroup)  (error,string) {
	body2 := ""

		req, err := retryablehttp.NewRequest("GET",seed, nil)
		if err != nil {
	    	wg.Done()
			return err,body2
	    }
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246")

		//send request
		resp, err := nc.Do(req)
		
		if err != nil {
			wg.Done()
			return err,body2
		}
		

		//reading request
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
	    	wg.Done()
			return err,body2
	    }
	   resp.Body.Close()
	body2 = string(body)
	wg.Done()
	return err,body2
}

func urlprocess(body2 string, seed string, c conf, wg *sync.WaitGroup) ([]string) {
	//normal URL regex
	clean := []string{}
	pat := regexp.MustCompile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)	
	raws := pat.FindAllString(body2,-1)
	
	//removing all param
	for i, s := range raws {
		temp := strings.Split(s,"?")
		temp = strings.Split(temp[0],")")
		temp = strings.Split(temp[0],"\"")
		raws[i] = temp[0]
		
	}
	
	//migrate to url format for easy URL segmentation
	unparse,err := url.QueryUnescape(seed)
	if err != nil {
		fmt.Printf("[-] QueryUnescape failed  : ")
		panic(err)
	}

	//get domain
	var parseed *(url.URL)
	if c.Toxtract != "" {
		parseed,err = url.Parse(c.Toxtract)
		if err != nil {
			fmt.Printf("[-] get domain url.Parse1 failed  : ")
			panic(err)
		}

	} else {
		parseed,err = url.Parse(unparse)
		if err != nil {
			fmt.Printf("[-] get domain url.Parse2 failed  : ")
			panic(err)
		}
	}
	
	//process what to xtract
	for _,raw := range raws {
		//if mode domain
		if c.Mode == "domain" {
			//domain regex
			unparse,err := url.QueryUnescape(raw)
			if err != nil {
    			fmt.Printf("[-] QueryUnescape failed  : ")
    			panic(err)
    		}
			
			u,err := url.Parse(unparse)
			if err != nil {
    			fmt.Printf("[-] process url.Parse failed  : ")
    			panic(err)
    		}

			//parse to obtain domain/subdomain specified in -u/-ul
			
			tem := strings.ReplaceAll(parseed.Host,".","\\.") 	
							
			pat2 := regexp.MustCompile(fmt.Sprintf(`(^|^[^:]+:\/\/|[^\.]+\.)`+tem))
			//pat2 := regexp.MustCompile(fmt.Sprintf(`(^|^[^:]+:\/\/|[^\.]+\.)w3\.com`))
			temp := pat2.FindAllString(u.Host,-1)
			if temp != nil {
				clean = append(clean,temp[0])
				
				
			}
			
		//if mode url
		} else if c.Mode == "url" {
			//mode xtennion enabled
			if c.Xtension != "" {

				formats := strings.Split(c.Xtension,",")
				
				formatx := ""
				for i := 0; i < len(formats); i++ {
				//for i, s := range formats {
					
					formatx += fmt.Sprintf("."+formats[i])
					if i != len(formats) -1 {
						formatx += "|"
					}

				}

				re:= regexp.MustCompile(`([a-zA-Z0-9\s_\\.\-\(\):])+(`+formatx+`)$`)
				if re.MatchString(raw) {
					clean = append(clean,raw)				
				}
			//mode xtension disabled
			} else {
				clean = append(clean,raw)
			}
			
		}	
	}
	wg.Done()
	return clean
		
}

func cleanandstore(c conf, clean []string, wg *sync.WaitGroup) {
	clean = removeDuplicateValues(clean)
	for _, s := range clean {
		fmt.Println(s)
		if c.Outname != "" {
			storehere(s+"\n",c.Outfile)
		}
	}
	wg.Done()
}

func cleanandstore2(c conf, clean []string) {
	clean = removeDuplicateValues(clean)
	for _, s := range clean {
		fmt.Println(s)
		if c.Outname != "" {
			storehere(s+"\n",c.Outfile)
		}
	}
}




func main() {
	
	//flag declaration
	var c = conf{}
	flag.StringVar(&(c.Url),"u","","single source where url will be taken")
	flag.StringVar(&(c.Urls),"ul","","source where url will be taken in a file")
	flag.StringVar(&(c.Mode),"m","domain","choose what to extract (domain,url)")
	flag.StringVar(&(c.Outname),"o","","output result to a file")
	flag.IntVar(&(c.Thread),"t",0,"goroutine number to be utilized (kinda like thread)")
	flag.StringVar(&(c.Xtension),"x","","xtension to extract (only work for URL mode)")
	flag.StringVar(&(c.Toxtract),"tx","","specify extract domain if it is different with source URL requested")
	flag.Parse()

	///building client
	//disable cert check
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	customTransport.IdleConnTimeout = time.Second * 5
	//default retry conf
	var nc = retryablehttp.NewClient()
	nc.RetryWaitMin = 2 * time.Second
	nc.RetryWaitMax = 2 * time.Second
	nc.RetryMax = 2
	nc.Logger = nil
	nc.HTTPClient = &http.Client{Transport: customTransport}
	
	//url to slice
	var seeds = []string{}
	seeds = urltoslice(c.Url,c.Urls)

	//preparing thread
    if c.Thread == 0 {
		c.Thread = len(seeds) 
	}

	//increase file descriptors
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
        fmt.Println("Error Getting Rlimit ", err)
    }
    if rLimit.Cur < uint64(c.Thread*3) {
    	rLimit.Cur = uint64(c.Thread*3)
    	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
    	if err != nil {
        	fmt.Println("Error Setting Rlimit ", err)
    	}
    }
    err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
    if err != nil {
       	fmt.Println("Error Getting Rlimit ", err)
    }
    //fmt.Println(rLimit)

	
	//prepare file for -o
	if c.Outname != "" {
	    var _ = os.Remove(c.Outname)
	    f,err := os.OpenFile(c.Outname,os.O_APPEND|os.O_WRONLY|os.O_CREATE,0644)
	    if err != nil {
	    	fmt.Printf("[-]create file error : ")
	    	panic(err)
	    }
	    c.Outfile = f
    }
	
	
    
    //final result slice declaration
    clean := []string{}

	//channel declaration
	body2 := make(chan string, len(seeds))
	body3 := make(chan string, len(seeds))
	body4 := make(chan []string, len(seeds))
	
	//goroutine guard number
    guard1 := make(chan struct{}, c.Thread)
    guard2 := make(chan struct{}, c.Thread)
    guard3 := make(chan struct{}, c.Thread)

    var wg sync.WaitGroup
 	wg.Add(len(seeds))
//goroutine1=======================================================================================================================================================================================================================
	for _,seed := range seeds {
		//building request
		guard1 <- struct{}{}
		go func(seed string)  {
			
			err,data := myrequest(nc,seed,&wg)
			if err != nil {
				if strings.Contains(err.Error(), "i/o timeout") {
	    			data = ""
	    		} else {
	    			fmt.Printf("[-] myrequest func error : ")
	    			fmt.Println(err)
	    			data = ""
	    		} 	
			}

			body2 <- data
			body3 <- seed
		
			<-guard1
			
		}(seed)
		
	}
//goroutine2=======================================================================================================================================================================================================================

	wg.Wait()
	wg.Add(len(seeds))
	for i := 0; i < len(seeds); i++ {
		guard2 <- struct{}{}
		go func() {

			message1 := <-body2
			message2 := <-body3
			body4 <- urlprocess(message1,message2,c,&wg)
			<-guard2
		}()

	}
	
	wg.Wait()
	close(body2)
	close(body3)
	

//goroutine3(optional)=======================================================================================================================================================================================================================				
	if c.Toxtract != "" {
		for i := 0; i < len(seeds); i++ {
	
			message3 := <- body4
			for _,mes := range message3 {
				clean = append(clean,mes)
			}
		}
		cleanandstore2(c,clean)
	} else {
		wg.Add(len(seeds))

		for i := 0; i < len(seeds); i++ {
			guard3 <- struct{}{}
			go func() {
				message3 := <- body4
				cleanandstore(c,message3,&wg)
				<-guard3
			}()
			
		}
		wg.Wait()
		close(body4)
		
	}

}
