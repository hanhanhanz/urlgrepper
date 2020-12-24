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


func main() {
	
	//flag declaration
	var c = conf{}
	flag.StringVar(&(c.Url),"u","","single source where url will be taken")
	flag.StringVar(&(c.Urls),"ul","","source where url will be taken in a file")
	flag.StringVar(&(c.Mode),"m","domain","choose what to extract (domain,url)")
	flag.StringVar(&(c.Outname),"o","","output result to a file")
	flag.IntVar(&(c.Thread),"t",10,"goroutine number to be utilized (kinda like thread) (in development)")
	flag.StringVar(&(c.Xtension),"x","","xtension to extract (only work for URL mode)")
	flag.StringVar(&(c.Toxtract),"tx","","specify extact domain if it is different with source URL requested (in development)")
	flag.Parse()

	//building client
	var nc = retryablehttp.NewClient()
	nc.RetryWaitMin = 1 * time.Second
	nc.RetryWaitMax = 2 * time.Second
	nc.RetryMax = 3
	nc.Logger = nil
	

	//body := bytes.NewReader([]byte(""))
	var seeds = []string{}
	if c.Url == "" && c.Urls == "" {
		errorKill("-u or -ul is mandatory")
	} else if c.Url != "" && c.Urls != "" {
		errorKill("choose either -u or -ul")
	} else if c.Url != "" {
		_, err := regexp.Match(`^https?://`, []byte(c.Url))
		if err != nil {
	    	fmt.Printf("[-] regex.match failed  : ")
	    	panic(err)
	    }
		seeds = append(seeds,c.Url)
	} else if c.Urls != "" {
		//prepare to read file from -ul
		var g *os.File
    	var g2 *bufio.Scanner
    	g,_ = os.Open(c.Urls) 
	    g2 = bufio.NewScanner(g)

    	for g2.Scan() {
    		var line = g2.Text()
    		_, err := regexp.Match(`^https?://`, []byte(c.Url))
			if err != nil {
	    		fmt.Printf("[-] regex.match failed  : ")
	    		panic(err)
	    	}
			seeds = append(seeds,line)
    	}
	}
	
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

	clean := []string{}
	
	for _,seed := range seeds {

		//building request
		req, err := retryablehttp.NewRequest("GET",seed, nil)
		if err != nil {
	    	fmt.Printf("[-] new request failed : ")
	    	panic(err)
	    }
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246")

		//send request
		resp, err := nc.Do(req)
		defer resp.Body.Close()
		if err != nil {
				fmt.Println("error")
				os.Exit(3)
			}
		
		//reading request
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
	    	fmt.Printf("[-] Reading request failed : ")
	    	panic(err)
	    }
		body2 := string(body)

		//normal URL regex
		pat := regexp.MustCompile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
		
		raws := pat.FindAllString(body2,-1)
		
		//removing all param
		for i, s := range raws {
			temp := strings.Split(s,"?")
			temp = strings.Split(temp[0],")")
			temp = strings.Split(temp[0],"\"")
			raws[i] = temp[0]
			
		}
		
		
		unparse,err := url.QueryUnescape(seed)
		
		
		if err != nil {
			fmt.Printf("[-] QueryUnescape failed  : ")
			panic(err)
		}

		var parseed *(url.URL)
		if c.Toxtract != "" {
			parseed,err = url.Parse(c.Toxtract)
		} else {
			parseed,err = url.Parse(unparse)
		}
		

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

				//parse to obtain domain/subdomain specified in -u/-ul
				
				tem := strings.ReplaceAll(parseed.Host,".","\\.") 	
								
				//pat2 := regexp.MustCompile(fmt.Sprintf(`(^|^[^:]+:\/\/|[^\.]+\.)`+tem))
				pat2 := regexp.MustCompile(fmt.Sprintf(`(^|^[^:]+:\/\/|[^\.]+\.)w3\.com`))
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

					//re:= regexp.MustCompile(`([a-zA-Z0-9\s_\\.\-\(\):])+(.js|.docx|.pdf)$`)
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
		
		
	}
	
	clean = removeDuplicateValues(clean)
	
	for _, s := range clean {
		fmt.Println(s)
		if c.Outname != "" {
			storehere(s+"\n",c.Outfile)
		}
	}
}
