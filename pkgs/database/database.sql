-- # create a databse

-- ## connect to database
psql -U your_username
or
sudo -i -u postgres

-- ## to enter in to query mode
psql

-- ## create database
CREATE DATABASE rideshare;

-- ## select the database
\c rideshare;

-- ## create tables

-- Create the driver table
CREATE TABLE driver (
    id SERIAL PRIMARY KEY,
    mobile_number VARCHAR(15) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'non-binary')) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create the vehicle table
CREATE TABLE vehicle (
    id SERIAL PRIMARY KEY,
    driver_id INT REFERENCES driver(id) ON DELETE CASCADE,
    vehicle_number VARCHAR(20) UNIQUE NOT NULL,
    vehicle_info TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create the ride table
CREATE TABLE ride (
    id SERIAL PRIMARY KEY,
    driver_id INT REFERENCES driver(id) ON DELETE SET NULL,
    start_datetime TIMESTAMP NOT NULL,
    estimated_end_datetime TIMESTAMP NOT NULL,
    number_of_seats INT CHECK (number_of_seats > 0 AND number_of_seats <= 60) NOT NULL,
    genders VARCHAR(20[]) NOT NULL,  -- Array of genders for each seat
    start_location VARCHAR(255) NOT NULL,
    end_location VARCHAR(255) NOT NULL,
    fare DECIMAL(10, 2) NOT NULL,
    route_details TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create the passenger interest table
CREATE TABLE passenger_interest (
    id SERIAL PRIMARY KEY,
    ride_id INT REFERENCES ride(id) ON DELETE CASCADE,
    passenger_mobile_number VARCHAR(15) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (ride_id, passenger_mobile_number)  -- Prevent duplicate interest
);

-- pg trigger functions
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- driver trigger
CREATE TRIGGER update_driver_timestamp
BEFORE UPDATE ON driver
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();


-- vehicle trigger
CREATE TRIGGER update_vehicle_timestamp
BEFORE UPDATE ON vehicle
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

-- ride trigger
CREATE TRIGGER update_ride_timestamp
BEFORE UPDATE ON ride
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();


