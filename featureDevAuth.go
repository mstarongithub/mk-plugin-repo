//go:build authDev

package main

import "fmt"

func init() {
	authMode = "dev"
	fmt.Println("Dev authentication")
}
