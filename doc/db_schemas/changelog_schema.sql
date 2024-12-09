CREATE TABLE public.changelog (
    store text NOT NULL,
    object_type text NOT NULL,
    object_id text NOT NULL,
    relation text NOT NULL,
    _user text NOT NULL,
    operation integer NOT NULL,
    ulid text NOT NULL,
    inserted_at timestamp with time zone NOT NULL
);

ALTER TABLE public.changelog OWNER TO jimm;

ALTER TABLE ONLY public.changelog
    ADD CONSTRAINT changelog_pkey PRIMARY KEY (store, ulid, object_type);

