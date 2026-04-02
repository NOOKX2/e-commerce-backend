package main

import (
  "fmt"
  "github.com/gosimple/slug"
)

func main(){
  fmt.Println(slug.Make("mobile phone"))
  fmt.Println(slug.Make("Mobile Phone"))
  fmt.Println(slug.Make("mobile  phone"))
}