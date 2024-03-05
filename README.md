# libra-web

 University of Virginia’s open access institutional repository

### Requirements
* Go 1.21+
* Node 21.6+

### Local setup
* Clone the easystore repo: https://github.com/uvalib/easystore.git
* From the easystore repo, navigate to ./db/sqlite and follow the steps in read.me
* Define an envirnment variable for the vue frontend: `export LIBRA_SRV=http://localhost:8085`
* Create a launch script to run the libra-web backend. Sample below.
* In another terminal window, launch the front end with: `npm run dev`


### Sample server start script
Note: the jwt key can be found in the AWS Secretes Manager under: staging/jwt/libra

   USER_IP=`[PATH_TO]/terraform-infrastructure/scripts/resolve-private.ksh user-ws-staging.private.staging | awk -F: '{print $2}' | sed -e s/\ //g`
   go run backend/*.go -port=8085 \
      -userws "http://$USER_IP:8080" \
      -jwtkey { randomized key used to generate jwt tokens} \
      -esmode sqlite \
      -devuser {your compute id}

