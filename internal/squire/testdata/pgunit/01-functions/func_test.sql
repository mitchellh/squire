CREATE OR REPLACE FUNCTION test_setup_account_with_default_org()
RETURNS VOID AS $$
DECLARE
  _id INTEGER;
BEGIN
  -- Create an initial user and default organization.
  INSERT INTO accounts DEFAULT VALUES RETURNING id INTO _id;
  INSERT INTO organizations (owner_id, user_default)
    VALUES (_id, 'true');
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION test_teardown_account_with_default_org()
RETURNS VOID AS $$
BEGIN
  -- We can just nuke accounts and all dependents.
  TRUNCATE accounts RESTART IDENTITY CASCADE;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION test_case_account_with_default_org()
RETURNS VOID AS $$
DECLARE
  _id INTEGER;
  rec RECORD;
BEGIN
  -- Get our user ID. By using STRICT we also verify there is only one.
  SELECT id FROM accounts INTO STRICT _id;

  -- Call our func and verify that we get the account and org.
  SELECT * FROM account_with_default_org(_id) INTO rec;
  PERFORM pgunit.test_assertNotNull('record should be set', rec);
  PERFORM pgunit.test_assertNotNull('record should have account', rec.account);
  PERFORM pgunit.test_assertNotNull('record should have org', rec.default_org);
END;
$$ LANGUAGE plpgsql;
