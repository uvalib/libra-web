#
#
#

# run application
cd bin; ./libra-web \
   -etdurl          ${ETD_URL} \
   -index           ${INDEX_URL} \
   -jwtkey          ${JWT_KEY} \
   -userws          ${USER_WS} \
   -namespace       ${ETD_NAMESPACE} \
   -busname         ${BUS_NAME} \
   -eventsrc        ${EVENT_SRC_NAME} \
   -auditqueryurl   ${AUDIT_QUERY_URL} \
   -depositauthurl  ${DEPOSIT_AUTH_URL} \
   -metricsqueryurl ${PAGE_METRICS_QUERY_URL} \
   -getorcidurl     ${ORCID_GET_DETAILS_URL} \
   -orcidurl        ${ORCID_CLIENT_URL} \
   -esproxy         ${ES_PROXY_URL}

#
# end of file
#
