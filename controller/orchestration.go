/*

LICENSE:  MIT
Author:   sine
Email:    sinerwr@gmail.com

*/

package controller

import (
	"encoding/json"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
	"regexp"

	"github.com/SiCo-Ops/Pb"
	"github.com/SiCo-Ops/dao/mongo"
)

type OrchestrationService struct {
}

type orchestration struct {
	HookID  string   `bson:"hookid"`
	Project string   `bson:"project"`
	Key     string   `bson:"key"`
	Value   string   `bson:"value"`
	Belong  string   `bson:"belong"`
	Task    []string `bson:"task"`
}

type HookResponse struct {
	Project  string `json:"project"`
	Branch   string `json:"branch"`
	CommitID string `json:"commitid"`
	Tag      string `json:"tag"`
	Time     string `json:"time"`
}

func SSHconn(user, ip, port, cmd string, b []byte) error {
	signer, _ := ssh.ParsePrivateKey(b)
	sshconfig := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", ip+":"+port, sshconfig)
	if err != nil {
		return err
	}
	sess, err := client.NewSession()
	if err != nil {
		return err
	}
	_, err = sess.Output(cmd)
	if err != nil {
		return err
	}
	return nil
}

func getPem(id, name string) ([]byte, error) {
	q := map[string]string{"id": id, "name": name}
	m, err := mongo.FindOne(userDB, mongo.CollectionUserRSAName("ssh"), q)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}
	key, _ := m["key"].(string)
	return []byte(key), nil
}

func (o *OrchestrationService) CreateRPC(ctx context.Context, in *pb.OrchestrationCreateCall) (*pb.OrchestrationCreateBack, error) {
	v := &orchestration{HookID: in.Hookid, Project: in.Project, Key: in.Key, Value: in.Value, Belong: in.Belong, Task: in.Task}
	err := mongo.Insert(orchestrationDB, mongo.CollectionOrchestrationName(), v)
	if err != nil {
		return &pb.OrchestrationCreateBack{Code: 206}, nil
	}
	return &pb.OrchestrationCreateBack{Code: 0}, nil
}

func (o *OrchestrationService) CheckRPC(ctx context.Context, in *pb.OrchestrationCheckCall) (*pb.OrchestrationCheckBack, error) {
	switch in.Type {
	case "dockerhub":
		v := &HookResponse{}
		json.Unmarshal(in.Params, v)
		q := map[string]string{"hookid": in.Hookid, "project": v.Project}
		m, err := mongo.FindAll(orchestrationDB, mongo.CollectionOrchestrationName(), q)
		if err != nil {
			return &pb.OrchestrationCheckBack{Code: 206}, nil
		}
		for _, result := range m {
			key, _ := result["key"].(string)
			if key != "tag" {
				continue
			}
			regexpstr, _ := result["value"].(string)
			ok, _ := regexp.MatchString(regexpstr, v.Tag)
			if !ok {
				continue
			}
			task, _ := result["task"].([]interface{})
			cmd, _ := task[0].(string)
			subcmd, _ := task[1].(string)
			switch cmd {
			case "swarm":
				if subcmd == "update" {
					user, _ := task[2].(string)
					ip, _ := task[3].(string)
					ns, _ := task[4].(string)
					pemName, _ := task[5].(string)
					pem, _ := getPem(in.Id, pemName)
					cmd := "docker service update --image " + ns + "/" + v.Project + ":" + v.Tag + " " + v.Project
					executor := []string{"ssh", user, ip, pemName, cmd}
					err = SSHconn(user, ip, "22", cmd, pem)
					if err != nil {
						return &pb.OrchestrationCheckBack{Code: 0}, nil
					}
					return &pb.OrchestrationCheckBack{Code: 0, Task: executor}, nil
				}
			}
		}
	}
	return &pb.OrchestrationCheckBack{Code: 0}, nil
}
