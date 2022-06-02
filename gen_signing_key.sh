#!/bin/sh

mkdir -p keys

openssl ecparam -name prime256v1 -genkey -noout -out keys/ec.key
openssl ec -in keys/ec.key -outform PEM -pubout -out keys/ec_pub.key
