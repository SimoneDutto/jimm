CREATE TABLE public.cloud_regions (
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    cloud_name text NOT NULL,
    name text NOT NULL,
    endpoint text NOT NULL,
    identity_endpoint text NOT NULL,
    storage_endpoint text NOT NULL,
    config bytea
);

ALTER TABLE public.cloud_regions OWNER TO jimm;

ALTER TABLE ONLY public.cloud_regions
    ADD CONSTRAINT cloud_regions_cloud_name_name_key UNIQUE (cloud_name, name);

ALTER TABLE ONLY public.cloud_regions
    ADD CONSTRAINT cloud_regions_pkey PRIMARY KEY (id);

CREATE INDEX idx_cloud_regions_deleted_at ON public.cloud_regions USING btree (deleted_at);

ALTER TABLE ONLY public.cloud_regions
    ADD CONSTRAINT cloud_regions_cloud_name_fkey FOREIGN KEY (cloud_name) REFERENCES public.clouds(name) ON DELETE CASCADE;

