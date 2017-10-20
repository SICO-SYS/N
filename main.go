/*

LICENSE:  MIT
Author:   sine
Email:    sinerwr@gmail.com

*/

package main

import (
	"net"

	"github.com/SiCo-Ops/N/controller"
)

func Run() {
	lis, _ := net.Listen("tcp", "0.0.0.0:"+controller.ServePort())
	controller.RPCServer.Serve(lis)
}

func main() {
	Run()
}
