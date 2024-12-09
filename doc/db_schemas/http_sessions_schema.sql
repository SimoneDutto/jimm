CREATE TABLE public.http_sessions (
    id bigint NOT NULL,
    key bytea,
    data bytea,
    created_on timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    modified_on timestamp with time zone,
    expires_on timestamp with time zone
);

ALTER TABLE public.http_sessions OWNER TO jimm;

ALTER TABLE ONLY public.http_sessions
    ADD CONSTRAINT http_sessions_pkey PRIMARY KEY (id);

CREATE INDEX http_sessions_expiry_idx ON public.http_sessions USING btree (expires_on);

CREATE INDEX http_sessions_key_idx ON public.http_sessions USING btree (key);

