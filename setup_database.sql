-- Create database if it doesn't exist
-- Run this as a superuser (postgres user)

-- Create the database
CREATE DATABASE generalusermanagement;

-- Connect to the database
\c generalusermanagement;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- The application will create the tables automatically when it starts
-- This is just to ensure the database exists and UUID extension is available

-- Create a user for the application (optional)
-- CREATE USER app_user WITH PASSWORD 'your_password';
-- GRANT ALL PRIVILEGES ON DATABASE generalusermanagement TO app_user;