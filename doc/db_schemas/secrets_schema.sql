CREATE TABLE public.secrets (
    id bigint NOT NULL,
    "time" timestamp with time zone,
    type text NOT NULL,
    tag text NOT NULL,
    data jsonb
);

ALTER TABLE public.secrets OWNER TO jimm;

ALTER TABLE ONLY public.secrets
    ADD CONSTRAINT secrets_pkey PRIMARY KEY (id);

CREATE UNIQUE INDEX idx_secret_name ON public.secrets USING btree (type, tag);

