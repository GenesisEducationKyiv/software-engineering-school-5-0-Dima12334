package main

import (
	"fmt"
	"net/url"
)

func main() {
	fmt.Println(url.QueryUnescape("%D0%9E%D0%B4%D0%B5%D1%81%D0%B0"))
	fmt.Println(url.QueryEscape("Одеса"))
}
