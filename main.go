package main

import (
	"csi_plugin/pkg/driver"
	"flag"
	"fmt"
)

func main() {

	/*
		Command line flags which return a pointer.
		Default values are the second argument.
		Third argument is the description.
		Running --help will see the values
	*/
	var (
		endpoint = flag.String("endpoint", "defaultValue", "Endpoint our gRPC server would run at")
		token    = flag.String("token", "defaultValue", "token of the storage provider")
		region   = flag.String("region", "ams3", "region where the volumes are going to be provisioned")
	)

	flag.Parse()

	fmt.Println(*endpoint, *token, *region)

	// Create a driver instance
	drv := driver.NewDriver(driver.InputParams{
		Name:     driver.DefaultName,
		Endpoint: *endpoint,
		Token:    *token,
		Region:   *region,
	})

	// Run on that driver instance and starts the gRPC server
	if err := drv.Run(); err != nil {
		fmt.Printf("Error %s running the driver", err.Error())
	}
}
