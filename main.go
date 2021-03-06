package main

import (
	"flag"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"os"
	"strings"
	"text/template"
)

type Inventory struct {
	Name        string
	IP          string
	PublicPort  int64
	PrivatePort int64
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var color, image string

func main() {

	colorPtr := flag.String("color", "green", "deploy color")
	imagePtr := flag.String("image", "ss/dream", "deploy image")
	verbPtr := flag.Bool("v", false, "verbose output info")
	outFilePtr := flag.String("o", "haproxy.cfg", "output file")
	socketPtr := flag.String("socket", "unix:///var/run/docker.sock", " socket to docker server")
	envPtr := flag.String("env", "", "Environment string in name docker")
	flag.Parse()

	color = *colorPtr
	image = *imagePtr
	environment := *envPtr
	servers := []*Inventory{}
	outFile := *outFilePtr
	endpoint := *socketPtr
	client, _ := docker.NewClient(endpoint)
	verb := *verbPtr

	if verb {
		fmt.Println("color: ", color)
		fmt.Println("image: ", image)
		fmt.Println("output: ", outFile)
		fmt.Println("docker socket: ", endpoint)
	}
	containers, err := client.ListContainers(docker.ListContainersOptions{All: false})
	check(err)
	for _, ss := range containers {

		if ss.Image == image {
			if strings.Contains(ss.Names[0], color) {
				if strings.Contains(ss.Names[0], environment) {
					serv := new(Inventory)
					cc, err := client.InspectContainer(ss.ID)
					check(err)
					if verb {
						fmt.Println("--------------------------------")
						fmt.Println("Names: ", ss.Names[0])
						fmt.Println("NN: ", cc.NetworkSettings.IPAddress)
						fmt.Println("ID: ", ss.ID)
						fmt.Println("Image: ", ss.Image)
						fmt.Println("Ports: ", ss.Ports)
						fmt.Println("Status: ", ss.Status)
					}
					for _, port := range ss.Ports {
						serv.PrivatePort = port.PrivatePort
						serv.PublicPort = port.PublicPort
						if verb {
							fmt.Println("ip + port: ", port.IP, port.PrivatePort, " public: ", port.PublicPort)
						}
					}
					serv.IP = cc.NetworkSettings.IPAddress
					serv.Name = strings.Replace(ss.Names[0], "/", "", -1)
					servers = append(servers, serv)
				}
			}
		}

	}
	if len(servers) > 0 {
		f, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		check(err)
		defer f.Close()
		var tmpl *template.Template
		if environment == "ss" {
			tmpl, err = template.ParseFiles("conf.d/" + environment + "/" + "haproxy.tmpl")
		} else {
			tmpl, err = template.ParseFiles("haproxy.tmpl")
		}
		check(err)
		err = tmpl.ExecuteTemplate(f, "haproxy.tmpl", servers)
		check(err)

	} else {
		fmt.Println("Error: server not  found")
		os.Exit(1)
	}
	os.Exit(0)
}
