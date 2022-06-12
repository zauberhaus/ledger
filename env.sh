RELEASE="${RELEASE:-localhost}"
NS="${NS:-immudb}"

ADDRESS="${ADDRESS:-$LB}"
PORT="${PORT:-3322}"

DB="${DB:-defaultdb}"

if [ "$RELEASE" = "localhost" ] ; then
    export CLIENT_OPTIONS_ADDRESS=localhost
    export CLIENT_OPTIONS_PORT=${PORT}
    export CLIENT_OPTIONS_USERNAME=immudb
    export CLIENT_OPTIONS_PASSWORD=immudb
    export CLIENT_OPTIONS_MTLS=false
    export CLIENT_OPTIONS_DATABASE=${DB}
else 
    LB=$(kubectl get service ${RELEASE}-immudb-primary -n $NS -o json | jq -r '.status.loadBalancer.ingress[0].ip')
    PASSWORD=$(kubectl get secret -n $NS ${RELEASE}-immudb -o json | jq -r '.data."ADMIN_PASSWORD"' | base64 -d)

    if [ ! -d ./certs ] ; then
        echo "Get client certificates ${RELEASE}-immudb-client-tls"

        mkdir -p ./certs/immudb

        SECRET=$(kubectl get secret -n $NS ${RELEASE}-immudb-client-tls -o json)

        echo $SECRET | jq -r ".data.\"tls.crt\"" | base64 -d > certs/immudb/tls.crt
        echo $SECRET | jq -r ".data.\"tls.key\"" | base64 -d > certs/immudb/tls.key
        echo $SECRET | jq -r ".data.\"ca.crt\"" | base64 -d > certs/immudb/ca.crt    
    fi

    export CLIENT_OPTIONS_ADDRESS=${ADDRESS}
    export CLIENT_OPTIONS_PORT=${PORT}
    export CLIENT_OPTIONS_USERNAME=immudb
    export CLIENT_OPTIONS_PASSWORD=${PASSWORD}
    export CLIENT_OPTIONS_MTLS=true
    export CLIENT_OPTIONS_DATABASE=${DB}
    export CLIENT_OPTIONS_MTLS_OPTIONS_CERTIFICATE=$(pwd)/certs/immudb/tls.crt
    export CLIENT_OPTIONS_MTLS_OPTIONS_CLIENT_CAS=$(pwd)/certs/immudb/ca.crt
    export CLIENT_OPTIONS_MTLS_OPTIONS_PKEY=$(pwd)/certs/immudb/tls.key
    export CLIENT_OPTIONS_MTLS_OPTIONS_SERVERNAME=${RELEASE}-immudb-primary
fi    