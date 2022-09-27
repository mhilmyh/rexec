package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {
	flag.Parse()
	os.Exit(Main())
}

type ServiceResponse struct {
	Address     string
	ServiceName string
	ServiceTags []string
}

var (
	host  = flag.String("h", "", "comma separated host")
	edit  = flag.Bool("e", false, "edit config")
	group = flag.String("g", "", "specify group to run")
)

func Main() int {
	if len(os.Args) < 2 {
		fmt.Println("usage: rexec [-e | -h <hosts>|-g <group>] <command>")
		return 0
	}

	var hosts []string
	if *host != "" {
		hosts = strings.Split(*host, ",")
	}
	args := flag.Args()

	if *edit {
		err := editConfig()
		if err != nil {
			fmt.Println(errColor(err.Error()))
		}
		return 0
	}

	if len(args) == 0 {
		return 0
	}

	source := strings.Split(*group, ":")

	if len(source) == 0 {
		fmt.Println(errColor("please insert group"))
		return 0
	}

	if len(source) > 2 {
		fmt.Println(errColor("group not supported"))
		return 0
	}

	var tag string
	if len(source) > 1 {
		tag = source[1]
	}

	addresses, err := readHostConfig(source[0])
	if err != nil {
		fmt.Println(errColor(err.Error()))
		return 0
	}

	var services []ServiceResponse
	for _, address := range addresses {
		s, err := getServices(address)
		if err != nil {
			fmt.Println(errColor(err.Error()))
		}
		services = append(services, s...)
	}

	if len(services) == 0 {
		return 0
	}

	for _, s := range services {
		if tag == "" || tag == "all" {
			hosts = append(hosts, "root@"+s.Address)
		} else {
			for _, st := range s.ServiceTags {
				if tag == st {
					hosts = append(hosts, "root@"+s.Address)
					break
				}
			}
		}
	}

	if len(hosts) == 0 {
		fmt.Println(errColor("tag not found"))
		return 0
	}

	var grCount int
	errChan := make(chan error)
	for _, host := range hosts {
		go run(host, args, errChan)
		grCount++
	}
	for grCount != 0 {
		err := <-errChan
		if err != nil {
			println(err.Error())
		}
		grCount--
	}

	return 0
}

func run(server string, command []string, err chan error) {
	cmds := []string{"tsh", "ssh", server, strings.Join(command, " ")}
	fmt.Println("Executing : ", cmds)

	cmd := exec.Command(cmds[0], cmds[1:]...)

	cmd.Stdout = newWriter(randColor(fmt.Sprintf("[%s] ", server)))
	cmd.Stderr = newWriter(errColor(fmt.Sprintf("[%s] ERR : ", server)))

	if errno := cmd.Run(); errno != nil {
		err <- fmt.Errorf("[%s] %s", server, errno.Error())
		return
	}

	err <- fmt.Errorf("[%s] %s", server, "session closed")
}

func getServices(address string) ([]ServiceResponse, error) {
	resp, err := http.Get(address)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var services []ServiceResponse
	err = json.NewDecoder(resp.Body).Decode(&services)
	if err != nil {
		return nil, err
	}
	return services, nil
}
