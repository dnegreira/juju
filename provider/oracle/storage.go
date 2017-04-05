// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package oracle

import (
	"github.com/juju/errors"

	"github.com/juju/juju/provider/oracle/common"
	"github.com/juju/juju/storage"
)

const (
	oracleStorageProvideType = storage.ProviderType("oracle")

	// maxVolumeSizeInGB is the maximum size that a volume in the
	// orcle cloud environ can be
	maxVolumeSizeInGB = 2000
	// minVolumeSizeInGB is the minimum size that a volume in the
	// oracle cloud environ ca be
	minVolumeSizeInGB = 1
	// maxDevices is the maximum number of volumes that a instnace can have
	maxDevices = 10
	// blockDevicePrefix is the default block storage prefix for a oracle
	// cloud storage volume
	blockDevicePrefix = "xvd"
)

// storageProvider is the storage provider for the oracle cloud storage environment
//
// this type implements the storage.Provider for obtaining storage sources
// this type implements the storage.VolumeSource for creating,destroying,describing,
// attaching and detaching volumes in the environment. A storage.VolumeSource is
// configured in a particular way, and corresponds to a storage "pool".
//
// Each storage volume can be up to 2 TB in capacity and you can attach
// up to 10 storage volumes to each instance. So there is
// an upper limit on the capacity of block storage that you can add to an instance
//
// A storage volume can be attached in read-only mode to only one instance.
// So multiple instances can’t write to a volume.
//
// After creating an instance, you can easily scale up
// or scale down the block storage capacity for the instance by
// attaching or detaching storage volumes. However, you can’t
// detach a storage volume that was attached during instance creation
// Note that, when a storage volume is detached from an instance,
// data stored on the storage volume isn’t lost
//
// If you intend to use this storage volume as a boot disk,
// then the size must be at least 5% higher than the boot image disk size
//
// Note, you can increase the size of a storage volume after creating it,
// even if the storage volume is attached to an instance. However, you can’t
// reduce the size of a storage volume after you’ve created it.
// So ensure that you don’t overestimate your storage requirement
//
//
// Volumes that require low latency and high IOPS,
// such as for storing database files, select storage/latency.
// For all other storage volumes, select storage/default
type storageProvider struct {
	env *oracleEnviron
	api StorageAPI
}

var _ storage.Provider = (*storageProvider)(nil)

// StorageAPI defines the storage API in the oracle cloud provider
// this enables the provider to talk with the storage api and make
// storage specific operations
type StorageAPI interface {
	common.StorageVolumeAPI
	common.StorageAttachmentAPI
	common.Composer
}

func mibToGib(m uint64) uint64 {
	return (m + 1023) / 1024
}

// StorageProviderTypes returns the storage provider types
// contained within this registry.
//
// Determining the supported storage providers may be dynamic.
// Multiple calls for the same registry must return consistent
// results.
func (o *oracleEnviron) StorageProviderTypes() ([]storage.ProviderType, error) {
	return []storage.ProviderType{oracleStorageProvideType}, nil
}

// StorageProvider returns the storage provider with the given
// provider type. StorageProvider must return an errors satisfying
// errors.IsNotFound if the registry does not contain said the
// specified provider type.
func (o *oracleEnviron) StorageProvider(t storage.ProviderType) (storage.Provider, error) {
	if t == oracleStorageProvideType {
		return &storageProvider{
			env: o,
			api: o.client,
		}, nil
	}

	return nil, errors.NotFoundf("storage provider %q", t)
}
