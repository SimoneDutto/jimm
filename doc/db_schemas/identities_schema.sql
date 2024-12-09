CREATE TABLE public.identities (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    display_name text NOT NULL,
    last_login timestamp with time zone,
    disabled boolean,
    access_token text,
    refresh_token text,
    access_token_expiry timestamp with time zone,
    access_token_type text
);

ALTER TABLE public.identities OWNER TO jimm;

ALTER TABLE ONLY public.identities
    ADD CONSTRAINT identities_name_key UNIQUE (name);

ALTER TABLE ONLY public.identities
    ADD CONSTRAINT identities_pkey PRIMARY KEY (id);

CREATE INDEX idx_identities_deleted_at ON public.identities USING btree (deleted_at);

