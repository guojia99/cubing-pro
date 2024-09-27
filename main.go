package main

import root "github.com/guojia99/cubing-pro/cmd"

func main() {
	cmd := root.NewRootCmd()
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
