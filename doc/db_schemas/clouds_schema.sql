CREATE TABLE public.clouds (
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    name text NOT NULL,
    type text NOT NULL,
    host_cloud_region text NOT NULL,
    auth_types bytea,
    endpoint text NOT NULL,
    identity_endpoint text NOT NULL,
    storage_endpoint text NOT NULL,
    ca_certificates bytea,
    config bytea
);

ALTER TABLE public.clouds OWNER TO jimm;

ALTER TABLE ONLY public.clouds
    ADD CONSTRAINT clouds_name_key UNIQUE (name);

ALTER TABLE ONLY public.clouds
    ADD CONSTRAINT clouds_pkey PRIMARY KEY (id);

