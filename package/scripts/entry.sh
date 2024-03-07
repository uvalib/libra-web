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

# S3
if [ "${ESMODE}" == "s3" ]; then
   ES_CONFIG="--esmode s3 --esdbhost ${DBHOST} --esdbport ${DBPORT} --esdb ${DBNAME} --esdbuser ${DBUSER} --esdbpass ${DBPASS} --esbucket ${ESBUCKET}"
   echo "Easystore S3 config: [${ES_CONFIG}]"
fi

# run application
cd bin; ./libra-web \
   -jwtkey       ${JWT_KEY} \
   -userws       ${USER_WS} \
   -oanamespace  ${OA_NAMESPACE} \
   -etdnamespace ${ETD_NAMESPACE} \
   -busname      ${BUS_NAME} \
   -eventsrc     ${EVENT_SRC_NAME} \
   ${ES_CONFIG}

#
# end of file
#
