\getenv name MICROJAAS_DB_NAME
\getenv password MICROJAAS_DB_PASSWORD

CREATE ROLE :"name" LOGIN PASSWORD :'password';
GRANT :"name" TO CURRENT_USER;
CREATE DATABASE :"name" OWNER :"name";
