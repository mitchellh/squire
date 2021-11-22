CREATE TABLE accounts (
  id         INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
COMMENT ON TABLE accounts IS 'All user accounts';

CREATE TABLE organizations (
  id         INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

  -- every organization has a single account that owns it.
  owner_id   INTEGER NOT NULL REFERENCES accounts (id),

  -- user default is true if this organization is a default organization
  -- created alongside a user. Default organizations CANNOT be deleted
  -- until the account is deleted.
  user_default BOOLEAN NOT NULL DEFAULT 'false'
);
COMMENT ON TABLE organizations IS $$
Organizations are made up of members who collectively have access to resources.
$$;
