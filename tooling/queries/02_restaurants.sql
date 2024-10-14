CREATE TABLE public.restaurants (
                                    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
                                    name character varying(255) NOT NULL,
                                    capacity jsonb NOT NULL,
                                    endorsements jsonb NOT NULL,
                                    location public.geography(Point,4326),
                                    opening_time time without time zone NOT NULL,
                                    closing_time time without time zone NOT NULL,
                                    PRIMARY KEY (id)
);
