CREATE TABLE public.roles (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    name text NOT NULL,
    uuid text NOT NULL
);

ALTER TABLE public.roles OWNER TO jimm;

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_name_key UNIQUE (name);

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_uuid_key UNIQUE (uuid);

