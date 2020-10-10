package xconsul

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/pborman/uuid"
)

type Registrar struct {
	broker       string
	client       consul.Client
	registration *api.AgentServiceRegistration
	exitch       chan int
}

func NewRegistrar(broker string) *Registrar {
	exitch := make(chan int, 1)
	return &Registrar{
		broker: broker,
		exitch: exitch,
	}
}

func (r *Registrar) Open(name string, subname string, portal string) error {
	log.Println("new consul registar", r.broker)

	ppart := strings.Split(portal, ":")
	if len(portal) < 2 {
		return fmt.Errorf("invalid portal", portal)
	}

	check := api.AgentServiceCheck{
		HTTP:     "http://" + portal + "/health",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Consul check service health status",
	}

	portalHost := ppart[0]
	portalPort, _ := strconv.Atoi(ppart[1])
	log.Println("portal", portalHost, portalPort)
	r.registration = &api.AgentServiceRegistration{
		ID:      name + "_" + subname + "_" + uuid.New(),
		Name:    name + "_" + subname,
		Address: portalHost,
		Port:    portalPort,
		Tags:    []string{name, subname},
		Check:   &check,
	}

	go r.loop()
	return nil
}

func (r *Registrar) Close() {
	r.exitch <- 1 // 退出循环
	if r.client != nil {
		log.Println("dergister from consul", r.broker)
		r.client.Deregister(r.registration)
	}
}

func (r *Registrar) loop() {
	config := api.DefaultConfig()
	config.Address = r.broker

	log.Println("start register to consul ...")
	sleep := time.Duration(0)
	for {
		select {
		case <-time.After(sleep * time.Second):
			client, err := register(config, r.registration)
			if err == nil {
				r.client = client
				log.Println("register to consul sucess!", r.broker)
				return
			}

			log.Println("register to consul failed!", err)
			sleep = 5

		case <-r.exitch:
			log.Println("process be canceled!")
			return
		}
	}
}

func register(config *api.Config, reg *api.AgentServiceRegistration) (consul.Client, error) {
	stdclient, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	client := consul.NewClient(stdclient)
	err = client.Register(reg)
	if err != nil {
		return nil, err
	}

	return client, nil
}
