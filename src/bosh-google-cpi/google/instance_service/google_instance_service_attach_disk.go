package instance

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"bosh-google-cpi/api"
	"bosh-google-cpi/util"
	"google.golang.org/api/compute/v1"
)

func (i GoogleInstanceService) AttachDisk(id string, diskLink string) (string, error) {
	var deviceName string

	// Find the instance
	instance, found, err := i.Find(id, "")
	if err != nil {
		return deviceName, err
	}
	if !found {
		return deviceName, api.NewVMNotFoundError(id)
	}

	deviceName = util.ResourceSplitter(diskLink)
	disk := &compute.AttachedDisk{
		DeviceName: deviceName,
		Mode:       "READ_WRITE",
		Source:     diskLink,
		Type:       "PERSISTENT",
	}

	// Attach the disk
	i.logger.Debug(googleInstanceServiceLogTag, "Attaching Google Disk '%s' to Google Instance '%s'", util.ResourceSplitter(diskLink), id)
	operation, err := i.computeService.Instances.AttachDisk(i.project, util.ResourceSplitter(instance.Zone), id, disk).Do()
	if err != nil {
		return deviceName, bosherr.WrapErrorf(err, "Failed to attach Google Disk '%s' to Google Instance '%s'", util.ResourceSplitter(diskLink), id)
	}

	if _, err = i.operationService.Waiter(operation, instance.Zone, ""); err != nil {
		return deviceName, bosherr.WrapErrorf(err, "Failed to attach Google Disk '%s' to Google Instance '%s'", util.ResourceSplitter(diskLink), id)
	}

	// Find the instance again, as we need to get the new attached disks info
	instance, found, err = i.Find(id, "")
	if err != nil {
		return deviceName, err
	}
	if !found {
		return deviceName, api.NewVMNotFoundError(id)
	}

	return deviceName, nil
}
