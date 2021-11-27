CREATE TABLE accounts (
  id         SERIAL PRIMARY KEY,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
