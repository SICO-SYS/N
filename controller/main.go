/*

LICENSE:  MIT
Author:   sine
Email:    sinerwr@gmail.com

*/

package controller

import (
	"github.com/getsentry/raven-go"
	"google.golang.org/grpc"
	"log"

	"github.com/SiCo-Ops/Pb"
	"github.com/SiCo-Ops/cfg"
	"github.com/SiCo-Ops/dao/mongo"
)

const (
	configPath string = "config.json"
)

var (
	config                              cfg.ConfigItems
	RPCServer                           = grpc.NewServer()
	orchestrationDB, orchestrationDBErr = mongo.NewDial()
	userDB, userDBErr                   = mongo.NewDial()
)

func ServePort() string {
	return config.RpcNPort
}

func init() {
	data, err := cfg.ReadFilePath(configPath)
	if err != nil {
		data = cfg.ReadConfigServer()
		if data == nil {
			log.Fatalln("config.json not exist and configserver was down")
		}
	}
	cfg.Unmarshal(data, &config)

	orchestrationDB, orchestrationDBErr = mongo.InitDial(config.MongoOrchestrationAddress, config.MongoOrchestrationUsername, config.MongoOrchestrationPassword)
	if orchestrationDBErr != nil {
		log.Fatalln(orchestrationDBErr)
	}
	userDB, userDBErr = mongo.InitDial(config.MongoUserAddress, config.MongoUserUsername, config.MongoUserPassword)
	if userDBErr != nil {
		log.Fatalln(userDBErr)
	}

	pb.RegisterOrchestrationServiceServer(RPCServer, &OrchestrationService{})

	if config.SentryNStatus == "active" && config.SentryNDSN != "" {
		raven.SetDSN(config.SentryNDSN)
	}
}
