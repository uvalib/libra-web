BEGIN;

CREATE TABLE IF NOT EXISTS  licenses (
   id serial PRIMARY KEY,
   label VARCHAR (80) NOT NULL,
   url VARCHAR (80)
);

INSERT INTO licenses (label, url) values 
   ('Attribution 4.0 International (CC BY)', 'http://creativecommons.org/licenses/by/4.0/'),
   ('Attribution-NoDerivatives 4.0 International (CC BY-ND)', 'http://creativecommons.org/licenses/by-nd/4.0'),
   ('Attribution-ShareAlike 4.0 International (CC BY-SA)', 'http://creativecommons.org/licenses/by-sa/4.0/'),
   ('Attribution-NonCommercial 4.0 International (CC BY-NC)', 'http://creativecommons.org/licenses/by-nc/4.0/'),
   ('Attribution-NonCommercial-NoDerivatives 4.0 International (CC BY-NC-ND)', 'http://creativecommons.org/licenses/by-nc-nd/4.0/'),
   ('Attribution-NonCommercial-ShareAlike 4.0 International (CC BY-NC-SA)', 'http://creativecommons.org/licenses/by-nc-sa/4.0'),
   ('CC0 1.0 Universal', 'http://creativecommons.org/publicdomain/zero/1.0/'),   
   ('All rights reserved by the author (no additional license for public reuse)', ''),
   ('Attribution 2.0 Generic (CC BY)', 'https://creativecommons.org/licenses/by/2.0/');

COMMIT;