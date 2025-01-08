-- remove non essential fields from cloud.
ALTER TABLE clouds  DROP COLUMN auth_types, DROP COLUMN endpoint, DROP COLUMN identity_endpoint, 
 DROP COLUMN storage_endpoint, DROP COLUMN ca_certificates, DROP COLUMN config;
