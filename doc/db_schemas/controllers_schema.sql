CREATE TABLE public.controllers (
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    name text NOT NULL,
    uuid text NOT NULL,
    admin_identity_name text NOT NULL,
    admin_password text NOT NULL,
    ca_certificate text NOT NULL,
    public_address text NOT NULL,
    cloud_name text NOT NULL,
    cloud_region text NOT NULL,
    deprecated boolean DEFAULT false NOT NULL,
    agent_version text NOT NULL,
    addresses bytea,
    unavailable_since timestamp with time zone,
    tls_hostname text
);

ALTER TABLE public.controllers OWNER TO jimm;

ALTER TABLE ONLY public.controllers
    ADD CONSTRAINT controllers_name_key UNIQUE (name);

ALTER TABLE ONLY public.controllers
    ADD CONSTRAINT controllers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.controllers
    ADD CONSTRAINT controllers_cloud_name_fkey FOREIGN KEY (cloud_name) REFERENCES public.clouds(name);

