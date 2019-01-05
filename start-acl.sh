#!/bin/sh

export ACL_INTERNAL_TOKEN="$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 128 ; echo '')"

# start webserver
nohup /server/webserver &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start acl webserver: $status"
  exit $status
fi

## generate services
echo "#empty" > services.yaml
echo "{}" > services.json

# update the nginx conf and make sure nginx is not dead
while true; do
    sleep 3

    # generate the services.yaml
    consul-template -once -consul-addr consul-node:8500 \
        -template ./services.yaml.ctmpl:./services-tmp.yaml

    # only update the conf if there is a change
    cmp -s ./services-tmp.yaml ./services.yaml
    RESULT=$?
    if [ $RESULT -eq 0 ]; then
      continue
    fi

    # update file
    mv ./services-tmp.yaml ./services.yaml

    # convert to json
    cat services.yaml | ./yaml2json > services.json

    # send new content to web server
    curl \
        --request POST \
        --upload-file services.json \
        --url http://localhost:8888/consul/services/change?token=${ACL_INTERNAL_TOKEN}
done