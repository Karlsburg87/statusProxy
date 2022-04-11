package statusproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

func Proxy() error {
	//httputil.NewSingleHostReverseProxy(nil)
	proxyTo := os.Getenv("PROXY_TO")
	if proxyTo == "" {
		log.Fatalln("must set PROXY_TO environment variable")
	}

	proxyDestination, err := url.Parse(proxyTo)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("proxying to: %+v\n", proxyDestination)

	//setup proxy
	proxy := httputil.ReverseProxy{
		Director: func(r *http.Request) { //TODO: authentication of messages and analytics logging happen here
			//change request destination from proxy address to backend
			r.URL.Scheme = proxyDestination.Scheme
			r.URL.Host = proxyDestination.Host
			r.URL.RawPath = proxyDestination.RawPath
			r.URL.RawQuery = proxyDestination.RawQuery

			//check request here to protect backend. Cancel req is unsatisfied
			if r.Method != http.MethodPost { //TODO: more checks here

				// create a cancellable context, and re-set the request to be able to cancel request and not bother backend
				//https://stackoverflow.com/questions/71031456/cancel-a-web-request-and-handle-errors-inside-the-reverseproxy-director-function
				ctx, cancel := context.WithCancel(r.Context())
				*r = *r.WithContext(ctx)
				log.Printf("only POST requests permitted to backend - received %s", r.Method)
				cancel()
			}
		},
		ModifyResponse: func(r *http.Response) error { //TODO: acknowledgements of receipts and exit logging happen here
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			//get body content to log
			var body []byte
			if r.Body != nil {
				var errT error
				body, errT = io.ReadAll(r.Body)
				if errT != nil {
					log.Println("request body readall error")
				}
				defer r.Body.Close()
			}
			//log error and return error json to client
			e := map[string]any{"error": err.Error(), "request_header": fmt.Sprintf("%+v", r.Header), "request_body": string(body)}
			log.Printf("%+v", e)
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(e)
		},
		Transport: &http.Transport{
			Proxy: func(r *http.Request) (*url.URL, error) {
				//allow aborted requests from Director
				select {
				case <-r.Context().Done():
					r.Body.Close()
					return nil, r.Context().Err()
				default:
				}

				//return backend url
				return proxyDestination, nil //TODO: in future there can be distribution of traffic amongst several instances of statusSentry VMs done here

			},
			TLSHandshakeTimeout:   20 * time.Second,
			ResponseHeaderTimeout: 20 * time.Second,
			ExpectContinueTimeout: 30 * time.Second,
			IdleConnTimeout:       1 * time.Minute,
			MaxIdleConns:          20,
			MaxConnsPerHost:       20,
		},
		FlushInterval: -1,
	}

	//setup server
	port := os.Getenv("PORT")
	if port != "" {
		port = "8080" //default
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]any{"message": "This is the proxy service for statusSentry. Please use the '/proxy' endpoint"}); err != nil {
			w.WriteHeader(http.StatusBadGateway)
		}
	})
	mux.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux, // could also use &proxy to use the proxy as global handler
	}

	//launch proxy server
	log.Printf("running on port %s", port)
	return server.ListenAndServe()
}
