CREATE OR REPLACE FUNCTION account_with_default_org(_id INTEGER)
RETURNS TABLE (
    account     accounts,
    default_org organizations
)
AS $$
BEGIN
  SELECT * FROM accounts WHERE id = _id INTO STRICT account;
  SELECT * FROM organizations WHERE owner_id = _id INTO STRICT default_org;
  RETURN NEXT;
END;
$$ LANGUAGE plpgsql;
