DO $$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_database
      WHERE datname = 'yeetfile'
   ) THEN
      PERFORM dblink_exec('dbname=postgres', 'CREATE DATABASE yeetfile');
   END IF;
END $$;
