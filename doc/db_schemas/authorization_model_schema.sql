CREATE TABLE public.authorization_model (
    store text NOT NULL,
    authorization_model_id text NOT NULL,
    type text NOT NULL,
    type_definition bytea,
    schema_version text DEFAULT '1.0'::text NOT NULL
);

ALTER TABLE public.authorization_model OWNER TO jimm;

ALTER TABLE ONLY public.authorization_model
    ADD CONSTRAINT authorization_model_pkey PRIMARY KEY (store, authorization_model_id, type);

