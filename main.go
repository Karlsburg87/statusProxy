package main

import (
	"log"

	sp "github.com/karlsburg87/statusProxy/internal/statusProxy"
)

func main() {
	log.Fatalln(sp.Proxy())
}
