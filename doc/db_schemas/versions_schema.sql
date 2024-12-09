CREATE TABLE public.versions (
    component text NOT NULL,
    major integer NOT NULL,
    minor integer NOT NULL
);

ALTER TABLE public.versions OWNER TO jimm;

ALTER TABLE ONLY public.versions
    ADD CONSTRAINT versions_pkey PRIMARY KEY (component);

