// Copyright 2025 Canonical.

package dbmodel

import (
	"time"

	jujuparams "github.com/juju/juju/rpc/params"
	"github.com/juju/names/v5"
	"gorm.io/gorm"
)

// A Cloud represents a cloud service.
type Cloud struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Name is the name of the cloud.
	Name string `gorm:"not null;uniqueIndex"`

	// Type is the provider type of cloud.
	Type string `gorm:"not null"`

	// HostCloudRegion is the "cloud/region" that hosts this cloud, if the
	// cloud is hosted.
	HostCloudRegion string

	// Regions contains the regions associated with this cloud.
	Regions []CloudRegion `gorm:"foreignKey:CloudName;references:Name"`
}

// Tag returns a names.Tag for this cloud.
func (c Cloud) Tag() names.Tag {
	return c.ResourceTag()
}

// ResourceTag returns a tag for this cloud.  This method
// is intended to be used in places where we expect to see
// a concrete type names.CloudTag instead of the
// names.Tag interface.
func (c Cloud) ResourceTag() names.CloudTag {
	return names.NewCloudTag(c.Name)
}

// SetTag sets the name of the cloud to the value from the given cloud tag.
func (c *Cloud) SetTag(t names.CloudTag) {
	c.Name = t.Id()
}

// Region returns the cloud region with the given name. If there is no
// such region a zero valued region is returned.
func (c Cloud) Region(name string) CloudRegion {
	for _, r := range c.Regions {
		if r.Name == name {
			return r
		}
	}
	return CloudRegion{}
}

// FromJujuCloud updates a Cloud object with the details from the given
// jujuparams.Cloud.
func (c *Cloud) FromJujuCloud(cld jujuparams.Cloud) {
	c.Type = cld.Type
	c.HostCloudRegion = cld.HostCloudRegion
	regions := make([]CloudRegion, 0, len(c.Regions))
	for _, r := range cld.Regions {
		reg := c.Region(r.Name)
		reg.FromJujuCloudRegion(r)
		reg.Config = Map(cld.RegionConfig[r.Name])
		regions = append(regions, reg)
	}
	c.Regions = regions
}

// A CloudRegion is a region of a cloud.
type CloudRegion struct {
	gorm.Model

	// Cloud is the cloud this region belongs to.
	CloudName string `gorm:"uniqueIndex:idx_cloud_region_cloud_name_name"`
	Cloud     Cloud  `gorm:"foreignKey:CloudName;references:Name;constraint:OnDelete:CASCADE"`

	// Name is the name of the region.
	Name string `gorm:"not null;uniqueIndex:idx_cloud_region_cloud_name_name"`

	// Endpoint is the API endpoint URL for the region.
	Endpoint string

	// IdentityEndpoint is the API endpoint URL of the region identity
	// service.
	IdentityEndpoint string

	// StorageEndpoint is the API endpoint URL of the region storage
	// service.
	StorageEndpoint string

	// Config contains the configuration associated with this region.
	Config Map

	// Controllers contains any controllers that can provide service for
	// this cloud-region.
	Controllers []CloudRegionControllerPriority
}

// ToJujuCloudRegion converts a CloudRegion into a jujuparams.CloudRegion.
func (r CloudRegion) ToJujuCloudRegion() jujuparams.CloudRegion {
	var cr jujuparams.CloudRegion
	cr.Name = r.Name
	cr.Endpoint = r.Endpoint
	cr.IdentityEndpoint = r.IdentityEndpoint
	cr.StorageEndpoint = r.StorageEndpoint
	return cr
}

// FromJujuCloudRegion updates a CloudRegion object with the details from
// the given jujuparams.CloudRegion.
func (cr *CloudRegion) FromJujuCloudRegion(r jujuparams.CloudRegion) {
	cr.Name = r.Name
	cr.Endpoint = r.Endpoint
	cr.IdentityEndpoint = r.IdentityEndpoint
	cr.StorageEndpoint = r.StorageEndpoint
}
