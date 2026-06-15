# libra-web

 University of Virginia’s open access institutional repository

### Requirements
* Go 1.21+
* Node 21.6+

### Database Notes

The backend uses a Postgres DB for user settting, registraction tracking, etc. The schema is managed by 
https://github.com/golang-migrate/migrate and the scripts are in ./backend/db/migrations.

Install the migrate binary on your host system. For OSX, the easiest method is brew. Execute:

`brew install golang-migrate`.

Define your PSQL connection params in an environment variable, like this:

`export LIBRADB=postgres://libraweb:pass@localhost:5432/libraweb`

Steps to create and run a migration:
* `migrate create -ext sql -dir backend/db/migrations -seq add_table`
* `migrate -database ${LIBRADB} -path backend/db/migrations/ up`


### Sample server start script
Note: the jwt key can be found in the AWS Secretes Manager under: staging/jwt/libra

   echo 'lookup user-ws ip...'
   USER_IP=`../../terraform-infrastructure/scripts/resolve-private.ksh user-ws-staging.private.staging | awk -F: '{print $2}' | sed -e s/\ //g`
   echo $USER_IP 

   echo 'lookup orcid-access-ws ip...'
   ORCID_IP=`../../terraform-infrastructure/scripts/resolve-private.ksh orcid-access-ws-staging.private.staging | awk -F: '{print $2}' | sed -e s/\ //g`
   echo $ORCID_IP 

   echo 'lookup libra-index ip...'
   INDEX_IP=` ../../terraform-infrastructure/scripts/resolve-private.ksh libra-index-staging.private.staging | awk -F: '{print $2}' | sed -e s/\ //g`
   echo $INDEX_IP
   
   echo 'lookup easystore-proxy ip...'
   PROXY_IP=`../../terraform-infrastructure/scripts/resolve-private.ksh libra-easystore-staging.private.staging | awk -F: '{print $2}' | sed -e s/\ //g`
   echo $PROXY_IP 

   echo 'lookup deposit-auth ip...'
   DEPOSIT_IP=`../../terraform-infrastructure/scripts/resolve-private.ksh deposit-auth-ws-staging.private.staging | awk -F: '{print $2}' | sed -e s/\ //g`
   echo $DEPOSIT_IP 

   echo 'starting up service...'

   go run backend/*.go -port=8085 \
      -userws "http://$USER_IP:8080" \
      -index "http://$INDEX_IP:8080" \
      -depositauthurl "http://$DEPOSIT_IP:8080" \
      -jwtkey { KEY } \
      -namespace libraetd \
      -esproxy "http://$PROXY_IP:8080" \
      -busname uva-libra-bus-staging \
      -eventsrc libra-web \
      -auditqueryurl https://as8xyv369e.execute-api.us-east-1.amazonaws.com/staging \
      -metricsqueryurl https://908na81f04.execute-api.us-east-1.amazonaws.com/staging \
      -getorcidurl "http://$ORCID_IP:8080" \
      -orcidurl https://orciddev.lib.virginia.edu \
      -dbuser libraweb -dbpass pass -dbname libraweb \
      -devuser { computeID} \
      -devrole admin

