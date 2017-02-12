/*
Package kgb provides functions for interacting with "kgb-bot".

See https://packages.debian.org/kgb-bot and/or https://kgb.alioth.debian.org/ for information about the upstream project.

	package main

	import (
		"go.tianon.xyz/kgb"
	)

	func main() {
		project := kgb.NewClient("http://localhost:5391").Project("example-repo-id", "example-repo-password")

		err := project.RelayMessage("hi y'all!")
		if err != nil {
			panic(err)
		}
	}
*/
package kgb // import "go.tianon.xyz/kgb"
