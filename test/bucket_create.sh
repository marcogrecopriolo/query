#!/bin/bash

# Copyright 2019-Present Couchbase, Inc.
#
# Use of this software is governed by the Business Source License included in
# the file licenses/Couchbase-BSL.txt.  As of the Change Date specified in that
# file, in accordance with the Business Source License, use of this software will
# be governed by the Apache License, Version 2.0, included in the file
# licenses/APL.txt.

Site=http://127.0.0.1:8091/pools/default/buckets
Auth=Administrator:password
bucket=(customer orders product purchase review shellTest)
q=${1:-250}

for i in "${bucket[@]}"
do
  if [ $i == 'orders' ]
   then
      curl --silent -X POST -u $Auth -d name=$i -d ramQuotaMB=$q -d bucketType=couchbase -d replicaNumber=0 $Site > /dev/null
  else
      curl --silent -X POST -u $Auth -d name=$i -d ramQuotaMB=$q -d bucketType=couchbase $Site > /dev/null
  fi
done

collections=('orders,_default,transactions' 'orders,_default,durability' 'orders,_default,flattenkeys')
for coll in "${collections[@]}"
do
    collpath=(${coll//,/ })
    curl --silent -X POST -u $Auth -d name=${collpath[2]} $Site/${collpath[0]}/scopes/${collpath[1]}/collections > /dev/null
done

cd filestore

mkdir -p data/dimestore/product
mkdir data/dimestore/customer
mkdir data/dimestore/orders
mkdir data/dimestore/review
mkdir data/dimestore/purchase
cd ../

UsersSite=http://localhost:8091/settings/rbac/users/local/
for i in "${bucket[@]}"
do
  Id=${i}owner
  Name=OwnerOf${i}
  Password=${i}pass
  curl --silent -X PUT $UsersSite$Id -d name=$Name -d roles=bucket_full_access[${i}] -d password=$Password -u $Auth > /dev/null
done

