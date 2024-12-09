CREATE TABLE public.cloud_defaults (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    identity_name text NOT NULL,
    cloud_id integer NOT NULL,
    region text NOT NULL,
    defaults bytea
);

ALTER TABLE public.cloud_defaults OWNER TO jimm;

ALTER TABLE ONLY public.cloud_defaults
    ADD CONSTRAINT cloud_defaults_identity_name_cloud_id_region_key UNIQUE (identity_name, cloud_id, region);

ALTER TABLE ONLY public.cloud_defaults
    ADD CONSTRAINT cloud_defaults_pkey PRIMARY KEY (id);

CREATE INDEX idx_cloud_defaults_deleted_at ON public.cloud_defaults USING btree (deleted_at);

ALTER TABLE ONLY public.cloud_defaults
    ADD CONSTRAINT cloud_defaults_cloud_id_fkey FOREIGN KEY (cloud_id) REFERENCES public.clouds(id);

ALTER TABLE ONLY public.cloud_defaults
    ADD CONSTRAINT cloud_defaults_username_fkey FOREIGN KEY (identity_name) REFERENCES public.identities(name);

