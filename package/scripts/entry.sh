#
#
#

#
# easy store configuration
#
ES_CONFIG=""

# sqlite
if [ "${ESMODE}" == "sqlite" ]; then
   ES_CONFIG="--esmode sqlite --esdbdir ${ES_DBDIR} --esdbfile sqlite.db"
   echo "Easystore sqlite config: [${ES_CONFIG}]"
fi

# Postgres
if [ "${ESMODE}" == "postgres" ]; then
   ES_CONFIG="--esmode postgres --esdbhost ${DBHOST} --esdbport ${DBPORT} --esdb ${DBNAME} --esdbuser ${DBUSER} --esdbpass ${DBPASS}"
   echo "Easystore Postgres config: [${ES_CONFIG}]"
fi

# run application
cd bin; ./libra-web \
   -jwtkey     ${JWT_KEY} \
   -userws     ${USER_WS} \
   ${ES_CONFIG}

#
# end of file
#
