package main

import (
	"cmp"
	"os"
)

func getHeaderUsername() string {
	return cmp.Or(os.Getenv(EnvHeaderUsername), DefaultHeaderUsername)
}

func getAddress() string {
	return cmp.Or(os.Getenv(EnvAddress), DefaultAddr)
}
