package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	// +kubebuilder:scaffold:imports
)

func main() {
	// Define input parameters
	name := flag.String("name", "", "Name of the IP CRD")
	ipType := flag.String("type", "physical", "IP address")
	cniVersion := flag.String("cniVersion", "0.3.1", "CNI version")
	ipList := flag.String("ips", "", "Comma-separated list of IP addresses")
	iface := flag.String("interface", "", "Network interface")
	owner := flag.String("owner", "", "Owner of the IP configuration")
	trust := flag.String("trust", "true", "Trust level of the IP configuration (optional)")

	// Parse input parameters
	flag.Parse()
	//if name is not provided, use the current timestamp
	if *name == "" {
		*name = time.Now().Format("20060102150405")
	}
	// Validate input parameters
	if len(*ipList) == 0 || *iface == "" || *owner == "" {
		fmt.Println("All parameters (ip, interface, owner) are required")
		flag.Usage()
		os.Exit(1)
	}

	// Create IpConfSpec
	// Create Kubernetes CRD
	crd := map[string]interface{}{
		"apiVersion": "dynamicconfigip.github.com/betav1",
		"kind":       "Ipconf",
		"metadata": map[string]interface{}{
			"name":      *name,
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"type":       *ipType,
			"owner":      *owner,
			"cniVersion": *cniVersion,
			"trust":      *trust,
			"ipItems": []map[string]interface{}{
				{
					"iface":     *iface,
					"ipaddress": strings.Split(*ipList, ",")[0],
					"netmask":   "24",
					"type":      "static",
				},
				{
					"iface":     *iface,
					"ipaddress": strings.Split(*ipList, ",")[1],
					"netmask":   "24",
					"type":      "static",
				},
			},
		},
	}

	// Convert CRD to JSON
	crdJSON, err := json.Marshal(crd)
	if err != nil {
		fmt.Printf("Error marshalling CRD: %v\n", err)
		os.Exit(1)
	}

	// Send CRD to Kubernetes API server
	k8sApiServerURL := "http://your-k8s-apiserver-url/apis/dynamicconfigip.github.com/v1/namespaces/default/ipconfs"
	client := &http.Client{}
	req, err := http.NewRequest("POST", k8sApiServerURL, bytes.NewBuffer(crdJSON))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")

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
