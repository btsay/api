package main

import (
	"github.com/btlike/api/server"
	"github.com/btlike/api/utils"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	utils.Init()
	server.Run(utils.Config.Address)
}
