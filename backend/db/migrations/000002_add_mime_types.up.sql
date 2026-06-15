BEGIN;

CREATE TABLE IF NOT EXISTS  mime_types (
   mime_type VARCHAR (20) UNIQUE NOT NULL
);

INSERT INTO mime_types (mime_type) values 
   ('text/csv'),
   ('application/pdf'),
   ('image/*'),
   ('text/html'),
   ('application/xml'),
   ('text/plain'),
   ('video/mp4'),
   ('video/quicktime'),
   ('audio/mp3'),
   ('audio/mpeg');

COMMIT;