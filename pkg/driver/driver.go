package driver

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
)

const DefaultName = "csi-example-driver"

var volNameKeyFromContPub string = "csi-example-driver/volume-name"

type Driver struct {
	name     string
	region   string
	endpoint string

	srv *grpc.Server
	csi.UnimplementedNodeServer
	csi.UnimplementedControllerServer
	csi.UnimplementedIdentityServer

	ready bool

	// httpserver - Driver can have this if something is trying to healthcheck
	// storage client - Driver can have this to interact with storage provider
}

type InputParams struct {
	Name     string
	Endpoint string
	Token    string
	Region   string
}

func NewDriver(params InputParams) *Driver {
	return &Driver{
		name:     params.Endpoint,
		region:   params.Region,
		endpoint: params.Endpoint,
	}
}

// Start the gRPC server
func (d *Driver) Run() error {
	url, err := url.Parse(d.endpoint)
	if err != nil {
		return fmt.Errorf("parsing the endpoint %s\n", err.Error())
	}

	if url.Scheme != "unix" {
		return fmt.Errorf("only supported scheme is unix, but was given: %s\n", url.Scheme)
	}

	/*
		unix:///var/lib/csi/sockets/csi.sock is our endpoint
		.FromSlash is needed for different OSes
	*/
	grpcAddress := path.Join(url.Host, filepath.FromSlash(url.Path))
	if url.Host == "" {
		grpcAddress = filepath.FromSlash(url.Path)
	}

	// Remove the previous directory if it already exists to prevent address already in use error
	if err := os.Remove(grpcAddress); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing listen address %s\n", err.Error())
	}

	// Start listening on the Unix domain socket at the mounted volume in /var/lib/csi/sockets
	listener, err := net.Listen(url.Scheme, grpcAddress)
	if err != nil {
		return fmt.Errorf("Listen failed %s\n", err.Error())
	}

	fmt.Println(listener)

	// Create a gRPC server
	d.srv = grpc.NewServer()

	// Register the RPCs from `csi` package to the gRPC server
	csi.RegisterNodeServer(d.srv, d)
	csi.RegisterControllerServer(d.srv, d)
	csi.RegisterIdentityServer(d.srv, d)

	// Return this when probe is called
	d.ready = true

	// Start the gRPC server
	return d.srv.Serve(listener)
}
