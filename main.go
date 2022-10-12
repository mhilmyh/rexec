package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	prefix = flag.String("p", "tsh ssh", "prefix when running the command")
	user = flag.String("u", "root", "default user when no user provided")
	host  = flag.String("h", "", "comma separated host")
	edit  = flag.Bool("e", false, "edit config")
	group = flag.String("g", "", "specify group to run")
	regexIpAddress *regexp.Regexp
)

func init() {
	regexIpAddress = regexp.MustCompile(`^(\w+)@((\d{1,3})(\.?)){5,}$`)
}

func main() {
	flag.Parse()
	os.Exit(Main())
}

type ServiceResponse struct {
	Address     string
}

func Main() int {
	if len(os.Args) < 2 {
		fmt.Println("usage: rexec [-e | -h <hosts> | -g <group> | -u <user> | -p <prefix>] <command>")
		return 0
	}

	var hosts []string
	if *host != "" {
		hosts = strings.Split(*host, ",")
	}

	if *edit {
		err := editConfig()
		if err != nil {
			fmt.Println(errColor(err.Error()))
		}
		return 0
	}

	args := flag.Args()
	if len(args) == 0 {
		return 0
	}

	var groups []string
	if *group != "" {
		groups = strings.Split(*group, ",")
	}

	var addresses []string
	for _, grp := range groups {
		addrs, err := readHostConfig(grp)
		if err != nil {
			fmt.Println(errColor(err.Error()))
			return 0
		}
		addresses = append(addresses, addrs...)
	}

	var services []ServiceResponse
	for _, address := range addresses {
		s, err := getServices(address)
		if err != nil {
			fmt.Println(errColor(err.Error()))
		}
		services = append(services, s...)
	}

	for _, s := range services {
		if regexIpAddress.MatchString(s.Address) {
			hosts = append(hosts, s.Address)
		}
		hosts = append(hosts, *user+"@"+s.Address)
	}

	if len(hosts) == 0 {
		fmt.Println(errColor("no host found"))
		return 0
	}

	var grCount int
	errChan := make(chan error)
	for _, h := range hosts {
		go run(h, args, errChan)
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
	cmdStr := *prefix + " " + server + " " + strings.Join(command, " ")
	cmds := strings.Split(strings.TrimSpace(cmdStr), " ")
	fmt.Println("Executing : '", cmds, "'")

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
	if regexIpAddress.MatchString(address) {
		return []ServiceResponse{{Address: address}}, nil
	}
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
