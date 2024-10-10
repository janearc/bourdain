CREATE TABLE IF NOT EXISTS diners (
                                      id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                      name VARCHAR(255) NOT NULL,
                                      preferences JSONB NOT NULL,
                                      location GEOGRAPHY(POINT, 4326)
);