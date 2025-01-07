package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/dynamicConfigIp/internal/controller"
)

func networkUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	reqs := controller.ParseNetworkUpdateRequest(&r.Body)
	if reqs == nil {
		http.Error(w, "Failed to parse request method", http.StatusBadRequest)
		return
	}
	var result error
	for _, req := range reqs {
		if err := configureNetwork(&req); err != nil {
			result = err
		}
	}
	if result != nil {
		http.Error(w, fmt.Sprintf("Failed to configure network: %v", result), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Network configuration updated successfully"))
}

func configureNetwork(req *controller.NetworkUpdateRequest) error {
	// Configure the IP address
	fmt.Println("Configuring network with request:", req)
	if req.IPAddress != "" && req.Netmask != "" && req.Interface != "" {
		if err := exec.Command("ip", "addr", "add", fmt.Sprintf("%s/%s", req.IPAddress, req.Netmask), "dev", req.Interface).Run(); err != nil {
			return fmt.Errorf("failed to add IP address: %v", err)
		}
		// Bring the interface up
		if err := exec.Command("ip", "link", "set", req.Interface, "up").Run(); err != nil {
			return fmt.Errorf("failed to bring interface up: %v", err)
		}
	}

	// Configure the gateway
	if req.Gateway != "" {
		if req.Destination == "" {
			if err := exec.Command("ip", "route", "add", "default", "via", req.Gateway).Run(); err != nil {
				return fmt.Errorf("failed to add default route: %v", err)
			}
		} else {
			if err := exec.Command("ip", "route", "add", req.Destination, "via", req.Gateway).Run(); err != nil {
				return fmt.Errorf("failed to add route: %v", err)
			}
		}
	}

	return nil
}

func main() {
	var httpPort string
	flag.StringVar(&httpPort, "httpPort", "8080", "http server port")
	flag.Parse()
	http.HandleFunc("/networkupdate", networkUpdateHandler)
	log.Println("Starting HTTP server on port ", httpPort, "...")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", httpPort), nil))
}
