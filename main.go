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

// var (
// 	//go:embed template/config/config.go
// 	configFileTemplate string
// )

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

/*
   protoDir: proto文件路径目录
   dstDir: 生成的go文件路径目录
   protoFiles: 指定需要的proto文件名称
*/
func generateProtoFile(protoDir, dstDir string, protoFiles ...string) {
	cmds := []string{
		"--proto_path=" + protoDir,
		"--go_out=" + dstDir,
		"--go_opt=paths=source_relative",
		"--go-grpc_out=" + dstDir,
		"--go-grpc_opt=paths=source_relative",
		// strings.Join(protoFiles, " "), // 注意：protoFiles不能放在这里，否则exec会把这个当成一个文件名
	}
	for _, f := range protoFiles {
		cmds = append(cmds, f)
	}
	cmd := exec.Command("protoc", cmds...)
	output, err := RunExecCommand(cmd, false, 3)
	if err != nil {
		log.Println(output)
		panic(err)
	}
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
