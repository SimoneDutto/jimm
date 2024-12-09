CREATE TABLE public.application_offers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    model_id bigint NOT NULL,
    name text NOT NULL,
    uuid text NOT NULL,
    url text NOT NULL
);

ALTER TABLE public.application_offers OWNER TO jimm;

ALTER TABLE ONLY public.application_offers
    ADD CONSTRAINT application_offers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.application_offers
    ADD CONSTRAINT application_offers_url_key UNIQUE (url);

ALTER TABLE ONLY public.application_offers
    ADD CONSTRAINT application_offers_uuid_key UNIQUE (uuid);

ALTER TABLE ONLY public.application_offers
    ADD CONSTRAINT application_offers_model_id_fkey FOREIGN KEY (model_id) REFERENCES public.models(id) ON DELETE CASCADE;

