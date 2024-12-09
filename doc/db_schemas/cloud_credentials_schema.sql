CREATE TABLE public.cloud_credentials (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    cloud_name text NOT NULL,
    owner_identity_name text NOT NULL,
    name text NOT NULL,
    auth_type text NOT NULL,
    label text NOT NULL,
    attributes_in_vault boolean NOT NULL,
    attributes bytea,
    valid boolean
);

ALTER TABLE public.cloud_credentials OWNER TO jimm;

ALTER TABLE ONLY public.cloud_credentials
    ADD CONSTRAINT cloud_credentials_cloud_name_owner_identity_name_name_key UNIQUE (cloud_name, owner_identity_name, name);

ALTER TABLE ONLY public.cloud_credentials
    ADD CONSTRAINT cloud_credentials_pkey PRIMARY KEY (id);

CREATE INDEX idx_cloud_credentials_deleted_at ON public.cloud_credentials USING btree (deleted_at);

ALTER TABLE ONLY public.cloud_credentials
    ADD CONSTRAINT cloud_credentials_cloud_name_fkey FOREIGN KEY (cloud_name) REFERENCES public.clouds(name) ON DELETE CASCADE;

ALTER TABLE ONLY public.cloud_credentials
    ADD CONSTRAINT cloud_credentials_owner_username_fkey FOREIGN KEY (owner_identity_name) REFERENCES public.identities(name);

