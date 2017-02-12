/*
Package kgb provides functions for interacting with "kgb-bot".

See https://packages.debian.org/kgb-bot and/or https://kgb.alioth.debian.org/ for information about the upstream project.

The primary goal of this package is to faithfully represent the protocol described in https://kgb.alioth.debian.org/kgb-protocol.html without too many frills.

	package main

	import (
		"go.tianon.xyz/kgb"
	)

	func main() {
		project := kgb.NewClient("http://localhost:5391").Project("example-repo-id", "example-repo-password")

		err := project.RelayMessage("hi, y'all!")
		if err != nil {
			panic(err)
		}

		err = project.CommitV4(kgb.CommitV4Args{
			CommitId:  "beefcake",
			Author:    "The Real Slim Shady",
			Branch:    "master",
			CommitLog: "Updated the evil master plan",
			Changes: []string{
				"M master-plan.md",
			},
			Extra: map[string]interface{}{
				"web_link":        "https://godoc.org/go.tianon.xyz/kgb",
				"use_irc_notices": 1,
				"use_color":       1,
			},
		})
		if err != nil {
			panic(err)
		}
	}
*/
package kgb // import "go.tianon.xyz/kgb"
