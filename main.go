package main

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

var (
	//go:embed template/config/config.go
	configFileTemplate string
)

func main() {
	// mkdir config
	if err := os.Mkdir("config", os.ModePerm); err != nil {
		panic(err)
	}
	// mkdir internal
	if err := os.Mkdir("internal", os.ModePerm); err != nil {
		panic(err)
	}
	// mkdir repository
	if err := os.Mkdir("repository", os.ModePerm); err != nil {
		panic(err)
	}
}

func generateProtoFile() {
	//protoc --proto_path=./protos_tmp --go_out=./gen/go --go-grpc_out=./gen/go ./protos_tmp/*/*.proto
	cmd := exec.Command("protoc", "--proto_path=./protos_tmp", "--go_out=./gen/go", "./protos_tmp/*/*.proto")
}

func RunExecCommand(cmd *exec.Cmd, stdout bool, timeout int) (string, error) {
	var (
		out io.ReadCloser
		err error
	)
	if stdout {
		out, _ = cmd.StdoutPipe() //指向cmd命令的stdout
	} else {
		out, _ = cmd.StderrPipe()
	}
	err = cmd.Start()
	if err != nil {
		log.Println("start command error: ", err)
		return "", err
	}
	killer := time.AfterFunc(time.Second*time.Duration(timeout), func() {
		cmd.Process.Kill()
		err = errors.New("run command timeout")
	})
	defer killer.Stop()
	outputBytes, _ := ioutil.ReadAll(out)
	if err := cmd.Wait(); err != nil {
		log.Println("read command error: ", err)
		return string(outputBytes), err
	}
	return string(outputBytes), err
}
