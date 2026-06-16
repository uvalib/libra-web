BEGIN;

CREATE TABLE IF NOT EXISTS registrations (
   id serial PRIMARY KEY,
   registrar VARCHAR (20) not null,
   degree VARCHAR (80) not null,
   program VARCHAR (80) not null,
   submitted_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS registration_students (
   id serial PRIMARY KEY,
   registration_id integer NOT NULL REFERENCES registrations(id) ON DELETE CASCADE,
   work_id VARCHAR (30) not null,
   compute_id VARCHAR (20) not null,
   completed_at TIMESTAMPTZ
);


COMMIT;