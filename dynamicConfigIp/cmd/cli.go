package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/bits"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	// +kubebuilder:scaffold:imports

	dynamicconfigipbetav1 "github.com/dynamicConfigIp/api/betav1"
)

const (
	cliServerAddress = "controller-manager-cli-service.system.svc.cluster.local"
	cliServerPort    = "8080"
)

func ValidateParameter(owner string, ipList string, gateway string, iface string, masklen string) error {
	if owner == "" {
		return errors.New("Owner is mandatory")
	}

	if ipList == "" && gateway == "" {
		return errors.New("Must provide at least one of ipList or gateway")
	}

	if len(ipList) != 0 && iface == "" {
		return errors.New("Must give iface when configure IP!")
	}
	mask, err := strconv.Atoi(masklen)
	if err != nil {
		return errors.New("Invalid mask length")
	}
	if mask > 32 {
		return errors.New("only support ipv4, Mask length must be less than or equal to 32")
	}
	return nil
}

func translateNetmask(netmask string) string {
	if strings.Contains(netmask, ".") {
		parts := strings.Split(netmask, ".")
		cidr := 0
		for _, part := range parts {
			num, err := strconv.Atoi(part)
			if err != nil {
				fmt.Printf("Invalid netmask format: %v\n", err)
				os.Exit(1)
			}
			cidr += bits.OnesCount(uint(num))
		}
		return fmt.Sprintf("%d", cidr)
	} else {
		return netmask
	}
}

func main() {
	// Define input parameters
	owner := flag.String("owner", "", "Name of the IP CRD")
	ipType := flag.String("type", "physical", "IP address")
	ipList := flag.String("ips", "", "Comma-separated list of IP addresses")
	netmask := flag.String("netmask", "24", "Netmask of the IP address")
	iface := flag.String("interface", "", "Network interface")
	gateway := flag.String("gateway", "", "gateway of the route")
	destination := flag.String(
		"destination",
		"0",
		"destination of the route, if it is default route, cannot ignore it",
	)
	// Parse input parameters
	flag.Parse()
	masklen := translateNetmask(*netmask)
	// Validate input parameters
	if err := ValidateParameter(*owner, *ipList, *gateway, *iface, masklen); err != nil {
		fmt.Println("Must give iface when configure IP!", err)
		flag.Usage()
		os.Exit(1)
	}
	// Convert netmask to CIDR notation if necessary
	var nesConf dynamicconfigipbetav1.IpconfSpec
	nesConf.Owner = *owner
	// Split IP list
	for _, ip := range strings.Split(*ipList, ",") {
		nesConf.IpItems = append(nesConf.IpItems, dynamicconfigipbetav1.IpAddressConfig{
			Iface:     *iface,
			Ipaddress: ip,
			Netmask:   *netmask,
			Type:      *ipType,
		})
	}
	if *gateway != "" {
		nesConf.IpItems = append(nesConf.IpItems, dynamicconfigipbetav1.IpAddressConfig{
			Type:        "gateway",
			Destination: *destination,
		})
	}
	neworkConfInJson, err := json.Marshal(nesConf)
	if err != nil {
		fmt.Printf("Error marshalling nesConf: %v\n", err)
		os.Exit(1)
	}

	// Send CRD to Kubernetes API server
	k8sApiServerURL := fmt.Sprintf("http://%s:%s/networkupdate", cliServerAddress, cliServerPort)
	client := &http.Client{}
	req, err := http.NewRequest("POST", k8sApiServerURL, bytes.NewBuffer(neworkConfInJson))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")

	client.Timeout = time.Second * 10
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error response from server: %v\n", resp.Status)
		os.Exit(1)
	}

	fmt.Println("CRD created successfully")
}
