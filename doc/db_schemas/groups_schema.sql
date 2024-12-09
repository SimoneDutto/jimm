CREATE TABLE public.groups (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    name text NOT NULL,
    uuid text NOT NULL
);

ALTER TABLE public.groups OWNER TO jimm;

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_name_key UNIQUE (name);

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_uuid_key UNIQUE (uuid);

CREATE INDEX idx_group_name ON public.groups USING btree (name);

