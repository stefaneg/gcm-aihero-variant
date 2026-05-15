package main

import (
	"os"

	appcmd "git-clone-manager/internal/cmd"
)

func main() {
	os.Exit(appcmd.Execute(os.Args[1:], os.Stdout, os.Stderr))
}
