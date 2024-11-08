## Specification

The goal of this specification is to analyse every dbmodel we have and evaluate if each field is needed for something other than the tests, or it can be retrieved from the other sources without using the database.

To analyse that we need to understand for each `dbmodel` where the field is written/updated and where it is read.

### Model

- `ID, Created At/Updated At, Name, UUID`: generic fields
- `OwnerIdentityName, Owner `:  W in `jimm.AddModel()`, U in `jimm.ImportModel()`. Read all around to get owner's identity.
- `ControllerID, Controller`: W in `jimm.AddModel()`, U in `jimm.ImportModel()`/`jimm.UpdateMigratedModel`. Read all around to connect to controller's api.
- `MigrationControllerID`: not used.
- `CloudRegionID, CloudRegion`: W in `jimm.AddModel(),jimm.ImportModel()`, R in `jimm.ModelInfo()`
- `CloudCredentialID,CloudCredential`: W in `jimm.AddModel()`, U in `jimm.ImportModel()`, R in `jimm.ModelInfo()`
- `Type`: W in `jimm.AddModel(),jimm.ImportModel()`, R in `jimm.ModelInfo()`
- `IsController`: W in `jimm.AddModel(),jimm.ImportModel()`, R in `jimm.ModelInfo()`
- `DefaultSeries`: W in `jimm.AddModel(),jimm.ImportModel()`, R in `jimm.ModelInfo()`
- `Life`: W in `jimm.AddModel(),jimm.ImportModel()`, U in `jimm.DestroyModel(), watcher.watchController()`, R in `watcher.watchController()`
- `Status`: W in `jimm.AddModel(),jimm.ImportModel()`, U in `watcher.watchController()`, R in `jimm.ModelInfo()`
- `Machine`: W in `jimm.AddModel(),jimm.ImportModel()`, U in `watcher.watchController()`, R in `jimm.ModelInfo()`
- `Cores`: W in `jimm.AddModel(),jimm.ImportModel()`, U in `watcher.watchController()`, R in `jimm.ModelInfo()`
- `Units`: W in `jimm.AddModel(),jimm.ImportModel()`, U in `watcher.watchController()`, R in `jimm.ModelInfo()`

