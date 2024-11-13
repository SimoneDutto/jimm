## Specification

The goal of this specification is to analyse every dbmodel we have and evaluate if each field is needed for something other than the tests, or it can be retrieved from the other sources without using the database.

To analyse that we need to understand for each `dbmodel` where the field is written/updated and where it is read.

### Model

- `ID, Created At/Updated At, `: generic fields.
- `Name`: coming from users. 
- `UUID`: coming from juju controllers.
- `OwnerIdentityName, Owner`:  owner's identity used by JIMM.
- `ControllerID, Controller`: Controller info to route request to controllers
- `MigrationControllerID`: not used.
- `CloudRegionID, CloudRegion`: Used to display info.
- `CloudCredentialID, CloudCredential`: Used to update model's cloud credentials. Potentially removable but not as easy. (potentially rethink cloud credentials associations)
- `Type`: Used to display info.
- `IsController`: Used to display info.
- `DefaultSeries`: Used to display info.
- `Life`: Used to handle model destroy, updated from the controller's watcher and jimm methods.
- `Status`: Used to display info, updated from the controller's watcher.
- `Machine`: Used to display info, updated from the controller's watcher.
- `Cores`: Used to display info, updated from the controller's watcher.
- `Units`: Used to display info, updated from the controller's watcher.

> Notes
> All things used to display info we can get from API calls w/o persisting anything.
> The watcher for applicationoffer can be removed in favor of retrieving Machine, Cores, Units from the api.

### Controller
- `ID, Created At/Updated At, Name, UUID`: generic fields
- `AdminIdentityName,AdminPassword`: Used to authtenticate requests going to controllers. Potentially removable because we could use `CredentialsStore`
- `CACertificate,PublicAddress,TLSHostname,Addresses`: Used by JIMM to route requests.
- `CloudName,CloudRegion`: used by JIMM to set cloud's priority
- `Deprecated`: used by JIMM to deprecate controllers.
- `AgentVersion`: used to retrieve earliest version of controller registered for JIMM
- `UnavailableSince`: updated by the controller's watcher when it is not available. Used by JIMM to make sure a controller is not available before deleting its models, and the controller from the db.
- `CloudRegions`: used by JIMM to set cloud's priority
- `Models`: not used (gorm reverse link maybe?).


### Cloud
- `ID, Created At/Updated At, Name, UUID`: generic fields
- `Type`: used by JIMM to decide to redact credentials.
- `HostCloudRegion`: ??
- `AuthTypes`: used to display info
- `Endpoint`: used to display info
- `IdentityEndpoint`: used to display info
- `StorageEndpoint`: used to display info
- `Regions`: CloudRegions
- `CACertificates`: used to display info
- `Config`: used to display info

> Notes
> we should ONLY allow removing hosted clouds (k8s).. atm you can remove any cloud

### CloudRegions
- `ID, Created At/Updated At, Name, UUID`: generic fields
- `CloudName, Cloud`: used to display info, used by JIMM to handle access (JWT, openfga)
- `Endpoint`: used to display info
- `IdentityEndpoint`: used to display info
- `StorageEndpoint`: used to display info
- `Config`: used to display info
- `Controllers(Priorities)`: used by JIMM to revoke cloud credentials, update Cloud definition, remove Cloud from Controller.

> Notes
> We store information to manage graceful destroy of controllers, clouds, regions.
> We could avoid taking decisions during facade methods and delete dangling permission, dbmodel. What if we do so in a separate cleanup routine.

### CloudRegionControllerPriority
- `ID, Created At/Updated At, Name, UUID`: generic fields
- `CloudRegionID,CloudRegion`: used by JIMM to handle access (JWT, openfga)
- `ControllerID,Controller`: referenced controller in the cloud priority.
- `Priority`: used by JIMM in AddHostedCloud

### CloudCredentials
- `ID, Created At/Updated At, Name, UUID`: generic fields
- `CloudName`: used by JIMM to retrieve CloudCredentials
- `Cloud`: used by JIMM to decide to redact credentials. 
- `OwnerIdentityName,Owner`: used by JIMM to set/check permissions when retrieving credentials.
- `AuthType`: used by JIMM to decide to redact credentials. 
- `Label`: not used
- `AttributesInVault`: used by JIMM to decide where to retrieve attributes from.
- `Attributes`: attributes saved in the database.
- `Valid`: used by JIMM to revoke credentials.
- `Models`: used by JIMM to retrieve models belonging to credentials. But I can't find any place were we keep this updated (useless?)

### ApplicationOffer
- `ID, Created At/Updated At, Name, UUID`: generic fields
- `ModelId, Model`: used by JIMM to extract the controller to offer the right url to consume by using JIMM database and controller's api.
- `ApplicationDescription`: used to display info
- `URL`: used to consume the offer
- `Endpoints`: used to display info
- `Spaces`: not used
- `Bindings`: not used
- `Connections`: used to display info
- `CharmUrl`: used to display info, updated by the controller's watcher

> The watcher for applicationoffer can be removed in favor of retrieving CharmUrl from the api.

### Identity
- `ID, Created At/Updated At, Name, UUID, DisplayName`: generic fields
- `DisplayName, LastLogin, Disabled, AccessToken, RefreshToken, AccessTokenExpiry, AccessTokenType`: use by JIMM for authentication purposes.
- `CloudCredentials`: not used

> Notes
> Identity is the core dbmodel for JIMM's permission model.

### IdentityModelDefaults
- `ID, Created At/Updated At, Name, UUID`: generic fields
- `IdentityName, Identity`: we never `SetIdentityModelDefaults`, so it's just a wrapper around identity. Potentially useless.
- `IdentityModelDefaults`: same as above.


> Notes
> `SetIdentityModelDefaults` is not used.

### Group
- `ID, Created At/Updated At, Name, UUID`: generic fields

> Notes
> Group is the core dbmodel for JIMM's permission model.

