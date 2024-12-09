CREATE TABLE public.models (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    name text NOT NULL,
    uuid text,
    owner_identity_name text NOT NULL,
    controller_id integer,
    cloud_region_id integer,
    cloud_credential_id bigint,
    type text NOT NULL,
    is_controller boolean NOT NULL,
    default_series text NOT NULL,
    life text NOT NULL,
    status_status text NOT NULL,
    status_info text NOT NULL,
    status_data bytea,
    status_since timestamp with time zone,
    status_version text NOT NULL,
    sla_level text NOT NULL,
    sla_owner text NOT NULL,
    cores bigint DEFAULT 0 NOT NULL,
    machines bigint DEFAULT 0 NOT NULL,
    units bigint DEFAULT 0 NOT NULL,
    migration_controller_id integer
);

ALTER TABLE public.models OWNER TO jimm;

ALTER TABLE ONLY public.models
    ADD CONSTRAINT models_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.models
    ADD CONSTRAINT models_uuid_key UNIQUE (uuid);

ALTER TABLE ONLY public.models
    ADD CONSTRAINT unique_model_names UNIQUE (owner_identity_name, name);

ALTER TABLE ONLY public.models
    ADD CONSTRAINT models_cloud_credential_id_fkey FOREIGN KEY (cloud_credential_id) REFERENCES public.cloud_credentials(id);

ALTER TABLE ONLY public.models
    ADD CONSTRAINT models_cloud_region_id_fkey FOREIGN KEY (cloud_region_id) REFERENCES public.cloud_regions(id);

ALTER TABLE ONLY public.models
    ADD CONSTRAINT models_controller_id_fkey FOREIGN KEY (controller_id) REFERENCES public.controllers(id);

ALTER TABLE ONLY public.models
    ADD CONSTRAINT models_migration_controller_id_fkey FOREIGN KEY (migration_controller_id) REFERENCES public.controllers(id);

ALTER TABLE ONLY public.models
    ADD CONSTRAINT models_owner_username_fkey FOREIGN KEY (owner_identity_name) REFERENCES public.identities(name);

