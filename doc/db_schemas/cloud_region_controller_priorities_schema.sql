CREATE TABLE public.cloud_region_controller_priorities (
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    cloud_region_id integer NOT NULL,
    controller_id integer NOT NULL,
    priority integer NOT NULL
);

ALTER TABLE public.cloud_region_controller_priorities OWNER TO jimm;

ALTER TABLE ONLY public.cloud_region_controller_priorities
    ADD CONSTRAINT cloud_region_controller_priorities_pkey PRIMARY KEY (id);

CREATE INDEX idx_cloud_region_controller_priorities_deleted_at ON public.cloud_region_controller_priorities USING btree (deleted_at);

ALTER TABLE ONLY public.cloud_region_controller_priorities
    ADD CONSTRAINT cloud_region_controller_priorities_cloud_region_id_fkey FOREIGN KEY (cloud_region_id) REFERENCES public.cloud_regions(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.cloud_region_controller_priorities
    ADD CONSTRAINT cloud_region_controller_priorities_controller_id_fkey FOREIGN KEY (controller_id) REFERENCES public.controllers(id) ON DELETE CASCADE;

