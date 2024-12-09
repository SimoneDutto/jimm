CREATE TABLE public.root_keys (
    id bytea NOT NULL,
    created_at timestamp with time zone NOT NULL,
    expires timestamp with time zone NOT NULL,
    root_key bytea NOT NULL
);

ALTER TABLE public.root_keys OWNER TO jimm;

ALTER TABLE ONLY public.root_keys
    ADD CONSTRAINT root_keys_pkey PRIMARY KEY (id);

CREATE INDEX idx_root_keys_created_at ON public.root_keys USING btree (created_at);

CREATE INDEX idx_root_keys_expires ON public.root_keys USING btree (expires);

