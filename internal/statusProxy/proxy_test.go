package statusproxy

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestProxy(t *testing.T) {
	//set necessary environment vars
	if err := os.Setenv("PORT", "8080"); err != nil {
		t.Fatalf("%v", err)
	}
	if err := os.Setenv("PROXY_TO", "http://localhost:8081"); err != nil {
		t.Fatalf("%v", err)
	}

	//start fake server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]any{"message": "Hello World!"}); err != nil {
			t.Fatalf("json encode error: %v", err)
		}
	})
	server := http.Server{
		Addr:              ":8081",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Println("test server running on 8081")
	go server.ListenAndServe()

	//start the proxy
	go func(t *testing.T) {
		if err := Proxy(); err != nil {
			t.Logf("proxy error: %+v", err)
			t.Fail()
		}
	}(t)

	//Send message via proxy
	res, err := http.Post("http://localhost:8080", "application/json", nil)
	if err != nil {
		t.Logf("http get error:Path %v\n", err)
		t.Fail()
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fail()
	}
	t.Logf("payload to client : %s with status: %s \n", string(b), res.Status)

	//check message is as expected
	message := make(map[string]any)
	if err := json.Unmarshal(b, &message); err != nil {
		t.Fatalf("json decode error: %v", err)
	}

	if message["message"] != "Hello World!" {
		t.Fatalf("Expected Hello World!, got %s", message["message"])
	}
}
