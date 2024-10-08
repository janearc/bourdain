CREATE TABLE IF NOT EXISTS reservations (
                                            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                            restaurant_id UUID REFERENCES restaurants(id),
                                            diner_id UUID REFERENCES diners(id),
                                            start_time TIMESTAMP NOT NULL,
                                            end_time TIMESTAMP NOT NULL,
                                            num_diners INTEGER NOT NULL
);