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

### CloudDefaults
- `IdentityName,Identity`: used by JIMM to retrieve cloud default.
- `CloudId,Cloud`: used by JIMM to retrieve cloud default.
- `Region`: used by JIMM to retrieve cloud default.
- `Map`: map that holds the default values.

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

# Controller
Analyse model manager facade methods to determine:
- juju input
- jimm's representation
- juju output
- potential juju api to get info instead of storing.
- need for jimm db

## `ConfigSet, ConfigSet`
we manipulate the Controller Config for JIMM. 
We don't use the config anywhere else, we don't use this field in JIMM. But maybe the client needs it for its thing.

# `AllModels`
- input: 
- jimm's representation: no
- output: `jujuparams.UserModelList`
- need for jimm db: no, we are not overriding the owner identity.

# `ControllerVersion`
we get the minimun value for all controllers.
We potentially can get this from the controllers api, instead of using jimm's db.

# `GetControllerAccess`
JIMM is needed because we are the one answering the access query.

# `IdentityProviderURL`
return always an empty string.

# `ModelConfig`
not implemented.

# `ModelStatus`
- input: 
- jimm's representation: none
- output: `jujuparams.ModelStatusResults`
- need for jimm db: no, but we are not overriding the owner identity.

# `MongoVersion`
not supported

# `WatchModelSummaries`
create a watcher for all models an user has access to.

# `WatchAllModelSummaries`
same as before but you need to be a jimm admin.

# `InitiateMigration`
used.


# Model manager

Analyse model manager facade methods to determine:
- juju input
- jimm's representation
- juju output
- potential juju api to get info instead of storing.
- need for jimm db

The goal is to understand what can be extracted from the juju controller's API and removed from JIMM.

## `ChangeModelCredential`
- input: `jujuparams.ChangeModelCredentialsParams`
- jimm's representation: `dbmodel.Model`
- output: `jujuparams.UserModelList`
- need for jimm db: yes
- potential 

All info we have stored in JIMM's db. No need to use api.

## `ListModels`
- input: -
- jimm's representation: `dbmodel.Model`
- output: `jujuparams.UserModelList`
- need for jimm db: yes

All info we have stored in JIMM's db. No need to use api.

## `ListModelSummaries`
- input: -
- jimm's representation: `dbmodel.Model`
- output: `ujuparams.ModelSummaryResults`
- potential juju api: `api.ModelInfo`
- need for jimm db: yes

## `CreateModel`
- input: `jujuparams.ModelCreateArgs`
- jimm's representation: `dbmodel.Model`
- output: `jujuparams.ModelInfo`
- need for jimm db: yes

## `DestroyModels`
- input: `jujuparams.DestroyModelsParams`
- jimm's representation: `dbmodel.Model`
- output: `jujuparams.ErrorResults`
- need for jimm db: yes

_Note: we don't need a watcher to finally release the model from the database, we can lazily get rid of models when contacting the controller's api to list models, and use newly Ales cleanup routine to remove dangling openfga permission._

## `ModelInfo`
- input: `jujuparams.Entities`
- jimm's representation: `dbmodel.Model`
- output: `jujuparams.ModelInfoResults`
- potential juju api: `api.ModelInfo`
- need for jimm db: yes

## `ModelStatus`
- input: `jujuparams.Entities`
- jimm's representation: `dbmodel.Model`
- output: `jujuparams.ModelStatusResults`
- potential juju api: `api.ModelInfo`
- need for jimm db: yes

## `DumpModelsDB`
- input: `jujuparams.Entities`
- jimm's representation: `dbmodel.Model`
- output: `jujuparams.MapResults`
- potential juju api: `api.DumpModelDB`
- need for jimm db: no

## `DumpModelsDB`
- input: `jujuparams.Entities`
- jimm's representation: `dbmodel.Model`
- output: `jujuparams.StringResults`
- potential juju api: `api.DumpModel`
- need for jimm db: no

## `SetModelDefaults`
- input: `jujuparams.SetModelDefaults`
- jimm's representation: `dbmodel.CloudDefault`
- output: `jujuparams.ErrorResults`
- need for jimm db: yes

## `UnsetModelDefaults`
- input: `jujuparams.UnsetModelDefaults`
- jimm's representation: `dbmodel.CloudDefault`
- output: `jujuparams.ErrorResults`
- need for jimm db: yes

## `ModelDefaultsForClouds`
- input: `jujuparams.Entities`
- jimm's representation: `dbmodel.CloudDefault`
- output: `jujuparams.ErrorResults`
- need for jimm db: yes

> question: why do we need to store cloud defaults on our side? Why don't we leave credentials storing to controller, and we just ask controllers "deploy this" and it verifies it has the right credentials.

## `ModifyModelAccess`
- input: `jujuparams.ModifyModelAccessRequest`
- jimm's representation: `openfga.Relation`
- output: `jujuparams.ErrorResults`
- need for jimm db: yes

## `ValidateModelUpgrades`
- input: `jujuparams.ValidateModelUpgradeParams`
- output: `jujuparams.ErrorResults`
- need for jimm db: no

# Application Offer

## `Offer`
- input: `jujuparams.AddApplicationOffers`
- jimm's representation: `jimm.AddApplicationOfferParams`
- output: `jujuparams.ErrorResults`
- need for jimm db: yes

## `GetConsumeDetails`
- input: `jujuparams.ConsumeOfferDetailsArg`
- jimm's representation: `dbmodel.ApplicationOffer`
- output: `jujuparams.ConsumeOfferDetailsResults`
- need for jimm db: yes

## `ListApplicationOffers`
- input: `jujuparams.OfferFilters`
- jimm's representation: `dbmodel.ApplicationOffer`
- output: `jujuparams.QueryApplicationOffersResultsV5`
- need for jimm db: yes

## `ModifyOfferAccess`
- input: `jujuparams.ModifyOfferAccessRequest`
- jimm's representation: `dbmodel.ApplicationOffer`
- output: `jujuparams.ErrorResults`
- need for jimm db: yes

## `DestroyOffers`
- input: `jujuparams.DestroyApplicationOffers`
- jimm's representation: `dbmodel.ApplicationOffer`
- output: `jujuparams.ErrorResults`
- need for jimm db: yes

## `FindApplicationOffers`
- input: `jujuparams.OfferFilters`
- jimm's representation: `dbmodel.ApplicationOffer`
- output: `jujuparams.QueryApplicationOffersResultsV5`
- need for jimm db: yes

## `ApplicationOffers`
- input: `jujuparams.OfferURLs`
- jimm's representation: `dbmodel.ApplicationOffer`
- output: `jujuparams.ApplicationOffersResults`
- need for jimm db: yes

