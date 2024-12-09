CREATE TABLE public.tuple (
    store text NOT NULL,
    object_type text NOT NULL,
    object_id text NOT NULL,
    relation text NOT NULL,
    _user text NOT NULL,
    user_type text NOT NULL,
    ulid text NOT NULL,
    inserted_at timestamp with time zone NOT NULL
);

ALTER TABLE public.tuple OWNER TO jimm;

ALTER TABLE ONLY public.tuple
    ADD CONSTRAINT tuple_pkey PRIMARY KEY (store, object_type, object_id, relation, _user);

CREATE INDEX idx_reverse_lookup_user ON public.tuple USING btree (store, object_type, relation, _user);

CREATE INDEX idx_tuple_partial_user ON public.tuple USING btree (store, object_type, object_id, relation, _user) WHERE (user_type = 'user'::text);

CREATE INDEX idx_tuple_partial_userset ON public.tuple USING btree (store, object_type, object_id, relation, _user) WHERE (user_type = 'userset'::text);

CREATE UNIQUE INDEX idx_tuple_ulid ON public.tuple USING btree (ulid);

