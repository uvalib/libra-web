-- Runs once, when the local postgres volume is first created.
-- Creates the easystore DB used by the backend's -deveasystoredb mode (in place of the easystore proxy).
-- Schema matches github.com/uvalib/easystore db/s3; files and metadata live in the local S3 bucket.
CREATE DATABASE easystore OWNER libraweb;
\connect easystore libraweb

CREATE TABLE objects (
   id         serial PRIMARY KEY,
   namespace  VARCHAR( 32 ) NOT NULL,
   oid        VARCHAR( 64 ) NOT NULL,
   vtag       VARCHAR( 32 ) UNIQUE NOT NULL,
   created_at timestamp DEFAULT NOW(),
   updated_at timestamp DEFAULT NOW()
);
CREATE UNIQUE INDEX objects_key_idx ON objects(namespace, oid);

CREATE TABLE fields (
   id         serial PRIMARY KEY,
   namespace  VARCHAR( 32 ) NOT NULL DEFAULT '',
   oid        VARCHAR( 64 ) NOT NULL DEFAULT '',
   name       VARCHAR( 32 ) NOT NULL DEFAULT '',
   value      TEXT NOT NULL DEFAULT '',
   created_at timestamp DEFAULT NOW(),
   updated_at timestamp DEFAULT NOW()
);
CREATE INDEX fields_key_idx ON fields(namespace, oid);
CREATE UNIQUE INDEX fields_distinct_idx ON fields(namespace, oid, name);

