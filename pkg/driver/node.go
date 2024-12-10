package driver

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (d *Driver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	// make sure all the req fields are present
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "VolumeID must be present in the NodeStageVolumeReq")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "StagingTargetPath must be present in the NodeStageVolReq")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "VolumeCaps must be present in the NodeStageVolReq")
	}

	switch req.VolumeCapability.AccessType.(type) {
	case *csi.VolumeCapability_Block:
		return &csi.NodeStageVolumeResponse{}, nil
	}

	volumeName := ""
	// Retrieve the volume name from the map
	if val, ok := req.PublishContext[volNameKeyFromContPub]; !ok {
		return nil, status.Error(codes.InvalidArgument, "Volumename is not present in the publish context of request")
	} else {
		volumeName = val
	}

	mnt := req.VolumeCapability.GetMount()
	fsType := "ext4"
	if mnt.FsType != "" {
		fsType = mnt.FsType
	}

	// Source is stored at /dev/disk/by-id/scsi-0DO_Volume
	source := getPathFromVolumeName(volumeName)
	// Target (staging area) is stored at /var/lib/kubelet
	target := req.StagingTargetPath

	// Format the volume and create a file system on it
	// For example: `mkfs.ext4 -F /dev/disk/by-id/scsi-0DO_Volume`
	err := formatAndMakeFS(source, fsType)
	if err != nil {
		fmt.Printf("unable to create fs error %s\n", err.Error())
		return nil, status.Error(codes.Internal, fmt.Sprintf("unable to create fs error %s\n", err.Error()))
	}

	// Mount the directory to the staging area
	err = mount(source, target, fsType, mnt.MountFlags)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error %s, mounting the source %s to taget %s\n", err.Error(), source, target))
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

// mount -t type device dir
func mount(source, target, fsType string, options []string) error {
	mountCmd := "mount"

	if fsType == "" {
		return fmt.Errorf("fstype is not provided")
	}

	mountArgs := []string{}

	// Create the staging directory with 0777 permissions
	err := os.MkdirAll(target, 0777)
	if err != nil {
		return fmt.Errorf("error: %s, creating the target dir\n", err.Error())
	}

	// Append arguments to the mount command
	mountArgs = append(mountArgs, "-t", fsType)

	// Append options to the mount comman
	if len(options) > 0 {
		mountArgs = append(mountArgs, "-o", strings.Join(options, ","))
	}

	// Append the source area (e.g. /dev/disk) and target directory to the mount command
	mountArgs = append(mountArgs, source)
	mountArgs = append(mountArgs, target)

	// Execute the mount command
	out, err := exec.Command(mountCmd, mountArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error %s, mounting the source %s to tar %s. Output: %s\n", err.Error(), source, target, out)
	}

	return nil
}

func getPathFromVolumeName(volName string) string {
	return fmt.Sprintf("/dev/disk/by-id/scsi-0DO_Volume_%s", volName)
}

func formatAndMakeFS(source, fsType string) error {
	// mkfsCmd is mkfs.ext4
	mkfsCmd := fmt.Sprintf("mkfs.%s", fsType)

	// Look for the mkfs binary
	_, err := exec.LookPath(mkfsCmd)
	if err != nil {
		return fmt.Errorf("unable to find the mkfs (%s) utiltiy errors is %s", mkfsCmd, err.Error())
	}

	// Form the 'mkfs.ext4 -F /dev/...' string
	mkfsArgs := []string{"-F", source}

	// Run the mkfs.ext4 -F /dev/... command
	out, err := exec.Command(mkfsCmd, mkfsArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("create fs command failed output: %s, and err: %s\n", out, err.Error())
	}

	return nil
}

func (d *Driver) NodeUnstageVolume(context.Context, *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	fmt.Printf("NodePublishVolume was called with source %s and target %s\n", req.StagingTargetPath, req.TargetPath)

	// Instructor had some usual validation code but this is skipped

	// Setup the 'bind' option
	options := []string{"bind"}
	if req.Readonly {
		options = append(options, "ro")
	}

	/*
		Here we are going to handle a request for filesystem. For block device mode the source is going to
		be the device directory where volume was attached from.
	*/
	fsType := "ext4"
	if req.VolumeCapability.GetMount().FsType != "" {
		// If the FsType is set on VolumeCapability then overwrite fsType
		fsType = req.VolumeCapability.GetMount().FsType
	}

	source := req.StagingTargetPath
	target := req.TargetPath

	// Run the bind mount
	// mount -t fstype source target -o bind,ro
	err := mount(source, target, fsType, options)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error %s, mounting the volume from staging dir to target dir", err.Error()))
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (d *Driver) NodeUnpublishVolume(context.Context, *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) NodeGetVolumeStats(context.Context, *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, nil
}

func (d *Driver) NodeExpandVolume(context.Context, *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, nil
}

func (d *Driver) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	fmt.Println("NodeGetCaps was called")
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
		},
	}, nil
}

func (d *Driver) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		// NodeID information comes from DigitalOcean API
		NodeId:            "",
		MaxVolumesPerNode: 5,
		AccessibleTopology: &csi.Topology{
			// DigitalOcean specific things
			Segments: map[string]string{
				"region": "am3",
			},
		},
	}, nil
}

func (d *Driver) mustEmbedUnimplementedNodeServer() {
}
