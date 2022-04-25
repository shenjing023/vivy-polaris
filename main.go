package main

import (
	_ "embed"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"text/template"
	"time"

	vtemplate "github.com/shenjing023/vivy-polaris/template"
)

var (
	//go:embed template/config/config.go
	configFileTemplate []byte
	//go:embed template/main_template.tpl
	serverFileTemplate string
	//go:embed template/repository/driver.tpl
	driverFileTemplate string
	//go:embed template/internal/handler.tpl
	handlerFileTemplate string
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
	if err := os.MkdirAll("repository/ent", os.ModePerm); err != nil {
		panic(err)
	}
	// mkdir grpc
	if err := os.MkdirAll(Cfg.Protobuf.DstDir, os.ModePerm); err != nil {
		panic(err)
	}
	os.WriteFile("config/config.go", configFileTemplate, os.ModePerm)

	// generate proto file
	generateProtoFile(Cfg.Protobuf.SourceDir, Cfg.Protobuf.DstDir, Cfg.Protobuf.CompileFiles...)

	// generate server file
	data := struct {
		PkgName    string
		ServerName string
		GRPCPath   string
	}{
		PkgName:    vtemplate.ImportPathForDir("."),
		ServerName: Cfg.ServerName,
		GRPCPath:   Cfg.Protobuf.DstDir,
	}
	f, err := os.Create("main.go")
	if err != nil {
		panic(err)
	}
	if err := template.Must(template.New("main.go").Parse(serverFileTemplate)).Execute(f, data); err != nil {
		panic(err)
	}

	// generate repository driver file
	f, err = os.Create("repository/driver.go")
	if err != nil {
		panic(err)
	}
	if err := template.Must(template.New("driver.go").Parse(driverFileTemplate)).Execute(f, data); err != nil {
		panic(err)
	}

	// generate handler file
	f, err = os.Create("internal/handler.go")
	if err != nil {
		panic(err)
	}
	if err := template.Must(template.New("handler.go").Parse(handlerFileTemplate)).Execute(f, data); err != nil {
		panic(err)
	}

	// go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	output, err := RunExecCommand(cmd, false, 30)
	if err != nil {
		panic(err)
	}
	log.Println(output)
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
