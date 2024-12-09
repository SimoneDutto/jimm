CREATE TABLE public.identity_model_defaults (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    identity_name text NOT NULL,
    defaults bytea
);

ALTER TABLE public.identity_model_defaults OWNER TO jimm;

ALTER TABLE ONLY public.identity_model_defaults
    ADD CONSTRAINT user_model_defaults_identity_name_key UNIQUE (identity_name);

ALTER TABLE ONLY public.identity_model_defaults
    ADD CONSTRAINT user_model_defaults_pkey PRIMARY KEY (id);

CREATE INDEX idx_user_model_defaults_deleted_at ON public.identity_model_defaults USING btree (deleted_at);

ALTER TABLE ONLY public.identity_model_defaults
    ADD CONSTRAINT user_model_defaults_username_fkey FOREIGN KEY (identity_name) REFERENCES public.identities(name);

