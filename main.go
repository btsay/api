package main

import (
	"os"

	"github.com/btsay/api/server"
	"github.com/btsay/api/utils"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "demo" {
			utils.Demo = true
		}
	}
	utils.Init()
	server.Run(utils.Config.Address)
}
