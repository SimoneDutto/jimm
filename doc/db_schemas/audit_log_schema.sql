CREATE TABLE public.audit_log (
    id bigint NOT NULL,
    "time" timestamp with time zone,
    model text,
    conversation_id text,
    message_id integer,
    facade_name text,
    facade_method text,
    facade_version integer,
    object_id text,
    identity_tag text,
    is_response boolean,
    params json,
    errors json
);

ALTER TABLE public.audit_log OWNER TO jimm;

ALTER TABLE ONLY public.audit_log
    ADD CONSTRAINT audit_log_pkey PRIMARY KEY (id);

CREATE INDEX idx_audit_log_identity_tag ON public.audit_log USING btree (identity_tag);

CREATE INDEX idx_audit_log_method ON public.audit_log USING btree (facade_method);

CREATE INDEX idx_audit_log_model ON public.audit_log USING btree (model);

CREATE INDEX idx_audit_log_time ON public.audit_log USING btree ("time");

