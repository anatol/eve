/*
 * Copyright (c) 2022. Zededa, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package hardware

import (
	"fmt"
	"time"

	"github.com/anatol/smart.go"
	"github.com/jaypipes/ghw"
	"github.com/lf-edge/eve/pkg/pillar/types"
)

func ReadSMARTinfoForDisk(diskName string) (*types.DiskSmartInfo, error) {
	dev, err := smart.Open(diskName)
	if err != nil {
		return nil, err
	}
	defer dev.Close()

	switch sm := dev.(type) {
	case *smart.SataDevice:
		return GetInfoFromSATAdisk(diskName, sm)
	case *smart.ScsiDevice:
		return GetInfoFromSCSIDisk(diskName, sm)
	case *smart.NVMeDevice:
		return GetInfoFromNVMeDisk(diskName, sm)
	default:
		// When cannot open the disk, it means that it will not be
		// possible to get SMART information from it. It's ok
		return getInfoFromUnknownDisk(diskName, "unknown"), nil
	}
}

// ReadSMARTinfoForDisks - сollects disks information via API,
// returns types.DisksInformation
func ReadSMARTinfoForDisks() (*types.DisksInformation, error) {
	disksInfo := new(types.DisksInformation)
	// Get information about disks
	block, err := ghw.Block()
	if err != nil {
		return nil, fmt.Errorf("error getting block storage info: %v", err)
	}

	for _, disk := range block.Disks {
		diskName := fmt.Sprintf("/dev/%v", disk.Name)

		diskSmartInfo, _ := ReadSMARTinfoForDisk(diskName)
		disksInfo.Disks = append(disksInfo.Disks, diskSmartInfo)
	}
	return disksInfo, nil
}

// getInfoFromUnknownDisk - takes a disk name (/dev/sda or /dev/nvme0n1)
// and disk type as input and returns *types.DiskSmartInfo
// indicating an unknown disk type
func getInfoFromUnknownDisk(diskName, diskType string) *types.DiskSmartInfo {
	diskInfo := new(types.DiskSmartInfo)
	diskInfo.DiskName = diskName
	diskInfo.DiskType = types.SmartDiskTypeUnknown
	diskInfo.Errors = fmt.Errorf("disk with name: %s have %s type", diskName, diskType)
	diskInfo.CollectingStatus = types.SmartCollectingStatusError
	diskInfo.TimeUpdate = uint64(time.Now().Unix())
	return diskInfo
}

// GetInfoFromSATAdisk - takes a disk name (/dev/sda or /dev/nvme0n1)
// as input and returns information on it
func GetInfoFromSATAdisk(diskName string, dev *smart.SataDevice) (*types.DiskSmartInfo, error) {
	diskInfo := new(types.DiskSmartInfo)

	diskInfo.DiskName = diskName
	diskInfo.TimeUpdate = uint64(time.Now().Unix())
	diskInfo.DiskName = diskName
	diskInfo.DiskType = types.SmartDiskTypeSata

	devIdentify, err := dev.Identify()
	if err != nil {
		diskInfo.Errors = fmt.Errorf("failed identify SATA device with name: %s; error:%v", diskName, err)
		diskInfo.CollectingStatus = types.SmartCollectingStatusError
		return diskInfo, diskInfo.Errors
	}

	smartAttrList, err := dev.ReadSMARTData()
	if err != nil {
		diskInfo.Errors = fmt.Errorf("failed read S.M.A.R.T. attr info from SATA device with name: %s; error:%v", diskName, err)
		diskInfo.CollectingStatus = types.SmartCollectingStatusError
		return diskInfo, diskInfo.Errors
	}

	for _, attr := range smartAttrList.Attrs {
		smartAttr := new(types.DAttrTable)
		smartAttr.ID = int(attr.Id)
		smartAttr.Flags = int(attr.Flags)
		smartAttr.RawValue = int(attr.VendorBytes[0])
		smartAttr.Value = int(attr.ValueRaw)
		smartAttr.Worst = int(attr.Worst)
		diskInfo.SmartAttrs = append(diskInfo.SmartAttrs, smartAttr)
	}

	diskInfo.ModelNumber = devIdentify.ModelNumber()
	diskInfo.SerialNumber = devIdentify.SerialNumber()
	diskInfo.Wwn = devIdentify.WWN()
	diskInfo.CollectingStatus = types.SmartCollectingStatusSuccess
	return diskInfo, nil
}

// GetInfoFromNVMeDisk - takes a disk name (/dev/sda or /dev/nvme0n1)
// as input and returns information on it
func GetInfoFromNVMeDisk(diskName string, dev *smart.NVMeDevice) (*types.DiskSmartInfo, error) {
	diskInfo := new(types.DiskSmartInfo)

	diskInfo.DiskName = diskName
	diskInfo.DiskType = types.SmartDiskTypeNvme
	diskInfo.TimeUpdate = uint64(time.Now().Unix())

	identController, _, err := dev.Identify()
	if err != nil {
		diskInfo.DiskName = diskName
		diskInfo.Errors = fmt.Errorf("failed  NVMe identifye error:%v", err)
		diskInfo.CollectingStatus = types.SmartCollectingStatusError
		return diskInfo, diskInfo.Errors
	}

	diskInfo.ModelNumber = identController.ModelNumber()
	diskInfo.SerialNumber = identController.SerialNumber()

	smartAttr, err := dev.ReadSMART()
	if err != nil {
		diskInfo.Errors = fmt.Errorf("failed read S.M.A.R.T. attr info from NVMe device with name: %s; error:%v", diskName, err)
		diskInfo.CollectingStatus = types.SmartCollectingStatusError
		return diskInfo, diskInfo.Errors
	}

	smartTemperature := new(types.DAttrTable)
	smartTemperature.ID = types.SmartAttrIDTemperatureCelsius
	smartTemperature.RawValue = int(smartAttr.Temperature)
	diskInfo.SmartAttrs = append(diskInfo.SmartAttrs, smartTemperature)

	smartPowerOnHours := new(types.DAttrTable)
	smartPowerOnHours.ID = types.SmartAttrIDPowerOnHours
	smartPowerOnHours.RawValue = int(smartAttr.PowerOnHours.Val[0])
	diskInfo.SmartAttrs = append(diskInfo.SmartAttrs, smartPowerOnHours)

	smartPowerCycles := new(types.DAttrTable)
	smartPowerCycles.ID = types.SmartAttrIDPowerCycleCount
	smartPowerCycles.RawValue = int(smartAttr.PowerCycles.Val[0])
	diskInfo.SmartAttrs = append(diskInfo.SmartAttrs, smartPowerCycles)
	diskInfo.CollectingStatus = types.SmartCollectingStatusSuccess

	return diskInfo, nil
}

// GetInfoFromSCSIDisk - takes a disk name (/dev/sda or /dev/nvme0n1)
// as input and returns information on it
func GetInfoFromSCSIDisk(diskName string, dev *smart.ScsiDevice) (*types.DiskSmartInfo, error) {
	diskInfo := new(types.DiskSmartInfo)

	diskInfo.DiskName = diskName
	diskInfo.TimeUpdate = uint64(time.Now().Unix())
	diskInfo.DiskType = types.SmartDiskTypeScsi
	var err error
	diskInfo.SerialNumber, err = dev.SerialNumber()
	if err != nil {
		diskInfo.Errors = fmt.Errorf("failed get SCSI device with name: %s; error:%v", diskName, err)
		diskInfo.CollectingStatus = types.SmartCollectingStatusError
		return diskInfo, diskInfo.Errors
	}
	diskInfo.CollectingStatus = types.SmartCollectingStatusSuccess

	return diskInfo, nil
}

// GetSerialNumberForDisk takes a disk name (from dev directory,
// for example /dev/sda or /dev/sda1) as input and return serial number
func GetSerialNumberForDisk(diskName string) (string, error) {
	dev, err := smart.Open(diskName)
	if err != nil {
		return "", fmt.Errorf("disk with name: %s have unknown type", diskName)
	}
	defer dev.Close()

	var diskSmartInfo *types.DiskSmartInfo
	switch sm := dev.(type) {
	case *smart.SataDevice:
		diskSmartInfo, err = GetInfoFromSATAdisk(diskName, sm)
	case *smart.ScsiDevice:
		diskSmartInfo, err = GetInfoFromSCSIDisk(diskName, sm)
	case *smart.NVMeDevice:
		diskSmartInfo, err = GetInfoFromNVMeDisk(diskName, sm)
	default:
		return "", fmt.Errorf("failed to get serial number for %s disk with type %s", diskName, dev.Type())
	}

	if err != nil {
		return "", err
	}
	return diskSmartInfo.SerialNumber, nil
}
