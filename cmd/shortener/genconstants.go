package main

// import (
// 	"fmt"
// 	"go/format"
// 	"os"
// 	"strings"
// )

// func init() {

// 	if _, err := os.Stat("const.go"); len(os.Args) == 1 || err == nil {
// 		return
// 	}

// 	values := make(map[string]string)
// 	if (len(os.Args)-1)%2 != 0 {
// 		fmt.Println("bad args")
// 		return
// 	}
// 	for i := 1; i < len(os.Args); i += 2 {
// 		values[os.Args[i]] = os.Args[i+1]
// 	}

// 	var sb strings.Builder
// 	fmt.Fprintf(&sb, "%s", `
// // Code generated by go generate; DO NOT EDIT.
// // This file was generated by genconstants.go

// package main

//     func init(){
//     // Print build information using formatted strings and default values if the variables are empty.
//     `)
// 	for k, v := range values {
// 		fmt.Fprintf(&sb, "%s = %q\n", k, v)
// 	}

// 	fmt.Fprintf(&sb, "}")

// 	generated := []byte(sb.String())
// 	formatted, err := format.Source(generated)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	err = os.WriteFile("const.go", formatted, 0644)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// }