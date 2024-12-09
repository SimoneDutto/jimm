CREATE TABLE public.assertion (
    store text NOT NULL,
    authorization_model_id text NOT NULL,
    assertions bytea
);

ALTER TABLE public.assertion OWNER TO jimm;

ALTER TABLE ONLY public.assertion
    ADD CONSTRAINT assertion_pkey PRIMARY KEY (store, authorization_model_id);

