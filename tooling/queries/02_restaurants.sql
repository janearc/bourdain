CREATE TABLE IF NOT EXISTS restaurants (
                                           id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                           name VARCHAR(255) NOT NULL,
                                           capacity JSONB NOT NULL,
                                           endorsements JSONB NOT NULL,
                                           location GEOGRAPHY(POINT, 4326),
                                           opening_time TIME NOT NULL,
                                           closing_time TIME NOT NULL
);
