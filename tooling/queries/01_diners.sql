CREATE TABLE public.diners (
                               id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
                               name character varying(255) NOT NULL,
                               preferences jsonb NOT NULL,
                               location public.geography(Point,4326),
                               PRIMARY KEY (id)
);
