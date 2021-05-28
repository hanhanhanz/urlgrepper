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
import "net"
import "crypto/tls"
import "syscall"
import "log"

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
	Verbose bool
	Outfile *(os.File)
	Rdirect int
}

type error interface {
   Error() string
}

func storehere(d string, f *(os.File) ) { //store result (in string) to file
	if _, err := f.WriteString(d); err != nil {
		log.Fatal("[-] storing function error :  ", err)
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
	    	log.Fatal("[-] regex.match failed  : ",err)
	    	
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
	    		log.Fatal("[-] regex.match failed  : ",err)
	    		
	    	}
			seeds = append(seeds,line)
			
    	}
	}
	return seeds
		
}


func myrequest(c conf, nc *retryablehttp.Client, seed string)  (error,string) {
	//fmt.Println("doMyRequest")
	body2 := ""

		//fmt.Println(seed)
		unparse,err := url.QueryUnescape(seed)
		u,err := url.Parse(unparse)
		//fmt.Println(u.Host)
		var domain = ""

		if err != nil {
			log.Println("[-] error, something wrong when parsing the url in directory: %s",err)
		}
		
		if u.Scheme == "" { //parsing when no http schema
			u.Scheme = "https" 

			domain = u.Scheme + "://" + seed
		
		} else { //parsing when there's http schema
			//domain = u.Scheme + "://" + u.Host + "/"
			domain = u.Scheme + "://" + u.Host 
			//temp = strings.Replace(u.Path,"/","",1)
		}

		var resp *(http.Response)
		cnt := 0
		for true {// check for redirect
			req, err := retryablehttp.NewRequest("GET",domain, nil)

			if err != nil {
		    	//wg.Done()
		    	log.Println("[-] Generate NewRequest failed : ",err)
				return err,body2
		    }
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246")
			
			//send request
			//fmt.Printf(domain + " ")
			resp, err = nc.Do(req)
			if err != nil {
				//wg.Done()
				log.Println("[-] request failure1 : ", err)
				return err,body2
			}
			//resp, err = http.Get(domain)
			//fmt.Println(resp.StatusCode)
			u2 := resp.Request.URL
			//test
			//unparse,err = url.QueryUnescape(u2)
			//u,err = url.Parse(u2)
			domain2 := u2.Scheme + "://" + u2.Host + u2.Path
			//fmt.Println(domain)
			//fmt.Println(u2)
			if domain != domain2 {
				req, err = retryablehttp.NewRequest("GET",domain2, nil)
				if err != nil {
			    	//wg.Done()
			    	log.Println("[-] Generate NewRequest failed : ",err)
					return err,body2
			    }

			    resp, err = nc.Do(req)
				if err != nil {
					//wg.Done()
					log.Println("[-] request failure2 : ", err)
					return err,body2
				}
				domain = domain2	
			} else { //check if there is redirect
				break
			}
			cnt += 1	
			if cnt == c.Rdirect {
				break //break if max redirect achive
			}
		}
		
		

		
		//reading request
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
	    	//wg.Done()
	    	log.Println("[-] response reading failure : ", err)
			return err,body2
	    }

	  resp.Body.Close()
	body2 = string(body)
	//wg.Done()
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
		log.Fatal("[-] QueryUnescape failed  : ",err)
		
	}

	//get domain
	var parseed *(url.URL)
	if c.Toxtract != "" {
		parseed,err = url.Parse(c.Toxtract)
		if err != nil {
			log.Fatal("[-] get domain url.Parse1 failed  : ",err)
			
		}

	} else {
		parseed,err = url.Parse(unparse)
		if err != nil {
			log.Fatal("[-] get domain url.Parse2 failed  : ",err)
			
		}
	}
	
	//process what to xtract
	for _,raw := range raws {
		//if mode domain
		if c.Mode == "domain" {
			//domain regex
			unparse,err := url.QueryUnescape(raw)
			if err != nil {
    			log.Fatal("[-] QueryUnescape failed  : ",err)
    			
    		}
			
			u,err := url.Parse(unparse)
			if err != nil {
    			log.Println("[-] process url.Parse failed  : ",err)
    			break
    			//panic(err)
    		}

			//parse to obtain domain/subdomain specified in -u/-ul
			
			tem := strings.ReplaceAll(parseed.Host,".","\\.") 	
			
			//pat2 := regexp.MustCompile(fmt.Sprintf(`(^|^[^:]+:\/\/|[^\.]+\.)`+tem))
			pat2 := regexp.MustCompile(fmt.Sprintf(`(\*\.)?([\w\d]+\.)+[\w\d]?`+tem))
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
	//wg.Done()
	return clean
		
}

func cleanandstore(c conf, clean []string, wg *sync.WaitGroup) (out []string) {
	out = []string{}
	clean = removeDuplicateValues(clean)
	for _, s := range clean {
		//fmt.Println(s)
		out = append(out,s)
		if c.Outname != "" {
			storehere(s+"\n",c.Outfile)
		}
	}
	wg.Done()
	return out
}

func cleanandstore2(c conf, clean []string) (out []string) {
	out = []string{}
	clean = removeDuplicateValues(clean)
	for _, s := range clean {
		//fmt.Println(s)
		out = append(out,s)
		if c.Outname != "" {
			storehere(s+"\n",c.Outfile)
		}
	}
	return out
}



func compareslice(master []string, slave []string) (out bool) {
	p := 0
	res :=  false
	//fmt.Println(len(temp))

			for _,j := range master {
				for _,i := range slave {
					
					if j == i {
						p += 1
						//fmt.Println(p)
						if p == len(slave) {
							res = true
						}
					}
				}
			}
		return res
}

func compareslice2(master []string, slave []string) (out int) {
	p := 0
	res :=  0
	//fmt.Println(len(temp))

			for _,j := range master {
				for _,i := range slave {
					
					if j == i {
						p += 1
						//fmt.Println(p)
						if p == len(slave) {
							res = p
						}
					}
				}
			}
		return res
}	


func play(c conf, seeds []string ,nc *retryablehttp.Client) (out []string) {
	    //final result slice declaration
	    clean := []string{}

		//channel declaration
		body2 := make(chan string, len(seeds))
		body3 := make(chan string, len(seeds))
		body4 := make(chan []string, len(seeds))
		
		//goroutine guard number
	    guard1 := make(chan struct{}, c.Thread)
	    guard2 := make(chan struct{}, c.Thread)
	    //guard3 := make(chan struct{}, c.Thread)

	    var wg sync.WaitGroup
	 	//wg.Add(len(seeds))
	//goroutine1=======================================================================================================================================================================================================================
		//go func() {
			for _,seed := range seeds {
				wg.Add(1)		
				//building request
				guard1 <- struct{}{}
				//go func(seed string)  {
				go func(seed string)  {
					defer wg.Done()						
					//fmt.Println("doplaylv1")
					
					//fmt.Println()
					err,data := myrequest(c,nc,seed)
					if err != nil {
						if strings.Contains(err.Error(), "i/o timeout") {
			    			data = ""
			    		} else {
			    			log.Println("[-] myrequest func error : ",err)
			    			
			    			data = ""
			    		} 	
					}

					body2 <- data
					body3 <- seed

					
					<-guard1
				}(seed)
				
			}
			
		//}()

		//go func() {
			//time.Sleep(4 * time.Second)
			wg.Wait()
			close(body2)
			close(body3)

		//}()
	//goroutine2=======================================================================================================================================================================================================================
		
		
		//wg.Add(len(seeds))
		//go func() {
			for i := 0; i < len(seeds); i++ {
				wg.Add(1)
				guard2 <- struct{}{}
				go func(body2 chan string, body3 chan string) {
					defer wg.Done()	
					message1 := <-body2
					message2 := <-body3

					body4 <- urlprocess(message1,message2,c,&wg)
					

					
					<-guard2
				}(body2,body3)

			}
			
			wg.Wait()
			close(body4)
		//}()
		
		

	//goroutine3(optional)=======================================================================================================================================================================================================================				
		var master = []string{}
		//if c.Toxtract != "" {
			
			for i := 0; i < len(seeds); i++ {
		
				message3 := <- body4
				for _,mes := range message3 {
					clean = append(clean,mes)
				}
			}
			master = cleanandstore2(c,clean)
	
		return master//HERE

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
	flag.BoolVar(&(c.Verbose),"v",false,"enable verbose mode")
	flag.IntVar(&(c.Rdirect),"r",2,"max times of redirect (minimum 0)")
	flag.Parse()
	

	//switch for verbose mode
	if c.Verbose == false {
		log.SetOutput(ioutil.Discard)	
	}
	

	///building client
	//disable cert check
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	customTransport.IdleConnTimeout = time.Second * 10
	customTransport.DialContext = (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 10 * time.Second}).DialContext


	//default retry conf
	var nc = retryablehttp.NewClient()
		nc.RetryWaitMin = 7 * time.Second
		nc.RetryWaitMax = 10 * time.Second
		nc.RetryMax = 1
		nc.Logger = nil
		nc.HTTPClient = &http.Client{
    		//CheckRedirect: func(req *http.Request, via []*http.Request) error {
      		//	fmt.Println("LOL")
      		//	return http.ErrUseLastResponse
      		Transport: customTransport,
  		} 


	//url to slice
	var seeds = []string{}
	seeds = urltoslice(c.Url,c.Urls)

	//preparing thread
    if c.Thread == 0 {
		c.Thread = len(seeds)
		//fmt.Println(seed)
		//fmt.Println(c.Thread)
	}

	//set redirect to 2 if stdin less than 2
	if c.Rdirect < 0 {
		c.Rdirect = 2
	}

	//increase file descriptors
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
        log.Fatal("[-] Error Getting Rlimit ", err)
    }
    if rLimit.Cur < uint64(c.Thread*3) {
    	rLimit.Cur = uint64(c.Thread*3)
    	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
    	if err != nil {
        	log.Fatal("[-] Error Setting Rlimit ", err)
    	}
    }
    err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
    if err != nil {
       	log.Fatal("[-] Error Getting Rlimit ", err)
    }
    //fmt.Println(rLimit)

	
	//prepare file for -o
	if c.Outname != "" {
	    var _ = os.Remove(c.Outname)
	    f,err := os.OpenFile(c.Outname,os.O_APPEND|os.O_WRONLY|os.O_CREATE,0644)
	    if err != nil {
	    	log.Fatal("[-] create file error : ",err)
	    	
	    }
	    c.Outfile = f
    }
	



	//indev============================================================================================================================================================
	//fmt.Println(seeds)

	//resp, err := nc.Do("https://google.com")
	//len(resp)
	
	var master = []string{}
	for _,sd := range seeds { 
		
		var seed = []string{}
		seed = append(seed,sd)
		
		dodol:=play(c,seed,nc)//HERE

		master = append(master,dodol...)
		master = removeDuplicateValues(master)
		fmt.Println(fmt.Sprintf("iteration 0, total domain : %d",len(master)))
		
		
		var new = []string{}
		
		c.Toxtract = sd
		
		
		if len(master) >= 1  {
			c.Thread = len(master) 
		} else {
			c.Thread = 1
		}		


		new = append(new,play(c,master,nc)...)	
		//fmt.Println(fmt.Sprintf("iteration 1, total domain : %d",len(removeDuplicateValues(master))))

		t := 0
		for true {
			
			master = removeDuplicateValues(append(master,new...))
			//if t != 0 {
				fmt.Println(fmt.Sprintf("iteration %d, total domain : %d",t+1,len(master)))	
			//}
			
			
			if (compareslice(master,new) == false || t < 2 ) && (len(new) != 0) { //new value found
				
				//sp := fmt.Sprintf("iteration %d with %d domain",t + 2,len(master))
				//fmt.Println(sp)
				var new2 = []string{}

				if len(master) > 1  {
					c.Thread = len(master) - 1
				} else {
					c.Thread = 1
				}

				new2 = append(new2,play(c,new,nc)...) //playnya dibikin goroutine biar gercep?
				//fmt.Println(new2)
				//fmt.Println(len(new2))
				new = new2
				
				t += 1
			
			} else {
				//fmt.Println("true")
				//fmt.Println(t)
				break
			}

		}

		master = removeDuplicateValues(master)
		//sp := fmt.Sprintf("iteration " + string(t) + " with " + string(len(master)) + " domain")
		//fmt.Println(sp)
		fmt.Println(fmt.Sprintf("final result : %d domain",len(master)))
		for _,i := range master { 
			fmt.Println(i)
		}

		//mesti make https*/
	}



}
