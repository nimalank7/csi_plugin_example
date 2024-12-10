package driver

import (
	"context"
	"fmt"
	"strconv"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/digitalocean/godo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (d *Driver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	fmt.Println("CreateVolume of the controller service was called")

	// name is present
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume must be called with a req name")
	}

	// extract required memory (e.g. 8Gi)
	// make sure the value is a positive integer and less than what the SP can support
	sizeBytes := req.CapacityRange.GetRequiredBytes()

	// volume capabilities is a required field so check if its been specified
	if req.VolumeCapabilities == nil || len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "")
	}

	const gb = 1024 * 2014 * 1024

	// ensure that accessMode specified in PVC (e.g. 'ReadWriteOnce') is actually supported by the SP

	// check if volumeContentSource is specified - optional field

	// handle accessibility requirements

	// DigitalOcean code to create a VolumeCreateRequest
	volReq := godo.VolumeCreateRequest{
		Name:          req.Name,
		Region:        d.region,
		SizeGigaBytes: sizeBytes / gb,
	}

	// Call the DigitalOcean API and provision a volume
	vol, _, err := d.storage.CreateVolume(ctx, &volReq)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed provisoing the volume error %s\n", err.Error()))
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			CapacityBytes: sizeBytes,
			VolumeId:      vol.ID,
			// specify content source, but only in cases where its specified in the PVC
		},
	}, nil
}

func (d *Driver) DeleteVolume(context.Context, *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	fmt.Println("ControllerPublishVolume of controller plugin was called")

	// check if volumeID is present and volume is available on SP
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "VolumeID is mandatory in ControllerPublishVolume request")
	}

	// if nodeID is set, and node is actually present on SP
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeID is mandatory in CPVolume request")
	}

	// Retrieve volume from Digital Ocean
	vol, _, err := d.storage.GetVolume(ctx, req.VolumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, "Volume is not available anymore")
	}

	// Instructor skips if the volume is already attached and some extra validation (e.g. readOnly etc...)

	// Convert string nodeID to integer
	nodeID, err := strconv.Atoi(req.NodeId)
	if err != nil {
		return nil, status.Error(codes.Internal, "was not able to convert nodeID to int value")
	}

	// Calls DigitalOcean API to attach the volume to the node
	action, _, err := d.storageAction.Attach(ctx, req.VolumeId, nodeID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed attaching volume to the node, error %s", err.Error()))
	}

	// Helper method 'waitForCompletetion' that polls DigitalOcean. I've omitted it for brevity

	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			volNameKeyFromContPub: vol.Name,
		},
	}, nil
}

func (d *Driver) ControllerUnpublishVolume(context.Context, *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) ValidateVolumeCapabilities(context.Context, *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, nil
}

func (d *Driver) ListVolumes(context.Context, *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, nil
}

func (d *Driver) GetCapacity(context.Context, *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, nil
}

func (d *Driver) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	caps := []*csi.ControllerServiceCapability{}

	for _, c := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	} {
		caps = append(caps, &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: c,
				},
			},
		})
	}

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: caps,
	}, nil

}

func (d *Driver) CreateSnapshot(context.Context, *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, nil
}

func (d *Driver) DeleteSnapshot(context.Context, *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, nil
}

func (d *Driver) ListSnapshots(context.Context, *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, nil
}

func (d *Driver) ControllerExpandVolume(context.Context, *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) ControllerGetVolume(context.Context, *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) ControllerModifyVolume(context.Context, *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) mustEmbedUnimplementedControllerServer() {
}
