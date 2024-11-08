## Specification

The goal of this specification is to analyse every dbmodel we have and evaluate if each field is needed for something other than the tests, or it can be retrieved from the other sources without using the database.

To analyse that we need to understand for each `dbmodel` where the field is written/updated and where it is read.

### Model

- `ID, Created At/Updated At, Name, UUID`: generic fields
- `OwnerIdentityName, Owner `:  owner's identity used by JIMM.
- `ControllerID, Controller`: Controller info to route request to controllers
- `MigrationControllerID`: not used.
- `CloudRegionID, CloudRegion`: Used to display info.
- `CloudCredentialID,CloudCredential`: Used when model is created.
- `Type`: Used to display info.
- `IsController`: Used to display info.
- `DefaultSeries`: Used to display info.
- `Life`: Used to handle model destroy, updated from the watch controller and jimm methods.
- `Status`: Used to display info, updated from the watch controller.
- `Machine`: Used to display info, updated from the watch controller.
- `Cores`: Used to display info, updated from the watch controller.
- `Units`: Used to display info, updated from the watch controller.

### Controller
- `ID, Created At/Updated At, Name, UUID`: generic fields
- `AdminIdentityName,AdminPassword`: Used to authtenticate requests going to controllers. Potentially removable because we use `CredentialsStore`
- `CACertificate,PublicAddress,TLSHostname,Addresses`: Used by JIMM to route requests.
- `CloudName,CloudRegion`: used by JIMM to set cloud's priority
- `Deprecated`: used by JIMM to deprecate controllers.
- `AgentVersion`: used to retrieve earliest version of controller registered for JIMM
- `UnavailableSince`: updated by the watcher when the controller is not available. Used by JIMM to make sure a controller is not available before deleting the models, and the controller from the db.
- `CloudRegions`: used by JIMM to set cloud's priority
- `Models`: not used

### Cloud
- `ID, Created At/Updated At, Name, UUID`: generic fields
- `Type`: used by JIMM to decide to redact credentials. 
- `HostCloudRegion`: ??
- `AuthTypes`: used to display info
- `Endpoint`: used to display info
- `IdentityEndpoint`: used to display info
- `StorageEndpoint`: used to display info
- `Regions`: ??
- `CACertificates`: used to display info
- `Config`: used to display info
