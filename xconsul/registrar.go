package xconsul

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/pborman/uuid"
)

type Registrar struct {
	broker       string
	client       *api.Client
	registration *api.AgentServiceRegistration
	healthPrefix string
	exitch       chan int
}

func NewRegistrar(broker string) *Registrar {
	exitch := make(chan int, 1)
	return &Registrar{
		broker: broker,
		exitch: exitch,
	}
}

func (r *Registrar) Open(name string, subname string, ver string, portal string, healthPrefix string) error {
	log.Println("new consul registar to", r.broker, "for", name, subname, ver, portal)
	r.healthPrefix = healthPrefix

	ppart := strings.Split(portal, ":")
	if len(ppart) < 2 {
		return fmt.Errorf("invalid portal", portal)
	}

	portalHost := ppart[0]
	portalPort, _ := strconv.Atoi(ppart[1])
	log.Println("portal", portalHost, portalPort)
	id := name + "_" + subname + "_" + uuid.New()
	r.registration = &api.AgentServiceRegistration{
		ID:      id,
		Name:    name + "_" + subname + "_" + ver,
		Address: portalHost,
		Port:    portalPort,
		Tags:    []string{name, subname},
		Check: &api.AgentServiceCheck{
			Timeout:                        "5s",
			Interval:                       "10s",
			HTTP:                           "http://" + portal + r.healthPrefix + "?id=" + id,
			DeregisterCriticalServiceAfter: "30s",
			Notes:                          "Consul check service health status",
		},
	}

	go r.loop()
	return nil
}

func (r *Registrar) Close() {
	r.exitch <- 1 // 退出循环
	if r.client != nil {
		log.Println("dergister from consul", r.broker)
		r.client.Agent().ServiceDeregister(r.registration.ID)
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

func register(config *api.Config, reg *api.AgentServiceRegistration) (*api.Client, error) {
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	err = client.Agent().ServiceRegister(reg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (r *Registrar) MakeHealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		params := req.URL.Query()
		// log.Println("params", params)
		idParam := params["id"]
		if len(idParam) < 1 {
			http.Error(w, "invalid id", 400)
			return
		}

		id := idParam[0]
		if id != r.registration.ID {
			http.Error(w, "invalid id", 400)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"status\":true}"))
	})
}
