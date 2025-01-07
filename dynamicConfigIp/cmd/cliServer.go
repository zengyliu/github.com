package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	dynamicconfigipbetav1 "github.com/dynamicConfigIp/api/betav1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme       = runtime.NewScheme()
	k8sClient    client.Client
	cliServerLog = ctrl.Log.WithName("cli server")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dynamicconfigipbetav1.AddToScheme(scheme))
	// Initialize the scheme
	if err := dynamicconfigipbetav1.AddToScheme(scheme); err != nil {
		log.Fatalf("Unable to add dynamicconfigipbetav1 to scheme: %v", err)
	}

	// Initialize the Kubernetes client
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Unable to get Kubernetes config: %v", err)
	}

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		log.Fatalf("Unable to create Kubernetes client: %v", err)
	}
	log.Printf("Kubernetes client created successfully")
}

func addIpconfCR(
	ipConfigSpec dynamicconfigipbetav1.IpconfSpec,
) (error, string) {
	ipConfCR := dynamicconfigipbetav1.Ipconf{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ipConfigSpec.Owner,
			Namespace: "default",
		},
		Spec: ipConfigSpec,
	}

	if err := k8sClient.Create(context.Background(), &ipConfCR); err != nil {
		fmt.Println("Failed to create Ipconf", err)
		return err, "Failed to create Ipconf"
	}
	return nil, "Create Ipconf CR successfully"
}

func updateIpconfCR(
	ipConfCR dynamicconfigipbetav1.Ipconf,
	ipConfigSpec dynamicconfigipbetav1.IpconfSpec,
) (error, string) {
	fmt.Println("Ipconf to be updated: ", ipConfCR)

	for _, item := range ipConfigSpec.IpItems {
		for _, existingItem := range ipConfCR.Spec.IpItems {
			if existingItem.Ipaddress != item.Ipaddress || existingItem.Gateway != item.Gateway {
				ipConfCR.Spec.IpItems = append(ipConfCR.Spec.IpItems, item)
				break
			}
		}
	}

	fmt.Println("Ipconf to be updated: ", ipConfCR)
	if err := k8sClient.Update(context.Background(), &ipConfCR); err != nil {
		fmt.Println("Failed to update Ipconf", err)
		return err, "Failed to create Ipconf"
	}
	return nil, "Update Ipconf CR successfully"
}

func updateIpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		cliServerLog.Error(nil, fmt.Sprintf("Invalid request method: %s", r.Method))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read request body", http.StatusInternalServerError)
		cliServerLog.Error(err, "Could not read request body")
		return
	}
	defer r.Body.Close()

	var ipConfigSpec dynamicconfigipbetav1.IpconfSpec
	if err := json.Unmarshal(body, &ipConfigSpec); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		cliServerLog.Error(err, "Invalid JSON format")
		return
	}
	cliServerLog.Info(fmt.Sprintf("IpConfigSpec: ", ipConfigSpec))
	// Fetch Ipconf from API server with cr name owner in json file
	// This is a placeholder for the actual API call
	// Replace with actual implementation
	var ipConfCR dynamicconfigipbetav1.Ipconf
	if err := k8sClient.Get(
		context.Background(),
		client.ObjectKey{Namespace: "default", Name: ipConfigSpec.Owner},
		&ipConfCR,
	); err != nil {
		if errors.IsNotFound(err) {
			// Create a new Ipconf CR with the ipConfigSpec
			// Add the Ipconf to the API server
			err, errorString := addIpconfCR(ipConfigSpec)
			if err != nil {
				http.Error(w, errorString, http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Failed to get Ipconf", http.StatusInternalServerError)
			fmt.Println("Failed to get Ipconf", err)
		}
	} else {
		// Check and update ipConfCR with ipConfigSpec
		// Update the Ipconf to the API server
		err, errorString := updateIpconfCR(ipConfCR, ipConfigSpec)
		if err != nil {
			http.Error(w, errorString, http.StatusInternalServerError)
			return
		}
	}
	cliServerLog.Info("Ipconf updated successfully")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ipconf updated successfully"))

}

func main() {
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/networkupdate", updateIpHandler)

	cliServerLog.Info(fmt.Sprintf("Starting server on port:%s...", port))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		cliServerLog.Error(err, "Could not start server")
	}
}
