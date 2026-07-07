-- controller_owner_name records the model owner as known by the backing Juju
-- controller. JAAS may alias a model's owner locally (e.g. imported models with
-- a new owner, or backing controller models), so this column preserves the owner
-- the backing controller actually knows, which is required when addressing the
-- model on the controller (e.g. building offer URLs).
--
-- It is nullable: when NULL the model's JAAS owner (owner_identity_name) is also
-- its controller-facing owner.
ALTER TABLE models ADD COLUMN IF NOT EXISTS controller_owner_name TEXT;

-- is_controller_model flags models that are a backing controller's own model,
-- tracked by JAAS.
ALTER TABLE models ADD COLUMN IF NOT EXISTS is_controller_model BOOLEAN NOT NULL DEFAULT FALSE;
