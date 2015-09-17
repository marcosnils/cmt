#!/bin/bash
set -e 
MACHINE_FROM=$1
shift
CONTAINER_ID_FROM=$1
shift
MACHINE_TO=$1
shift
CONTAINER_ID_TO=$1
shift
PRE_DUMP=$1


function dump() {
	ssh $MACHINE_TO "mkdir -p $CONTAINER_ID_TO/../"
	ssh $MACHINE_FROM  <<EOF 
cd $CONTAINER_ID_FROM
echo $(pwd)
mkdir -p images
sudo runc checkpoint --image-path ./images
sudo chown ubuntu -R ./
tar -cf lala.tar *
echo "Copying ..."
scp -r ./lala.tar $MACHINE_TO:$CONTAINER_ID_TO/lala.tar
rm lala.tar
EOF
echo "Starting migrated container"
ssh $MACHINE_TO "cd $CONTAINER_ID_TO; tar -xf lala.tar; nohup sudo runc restore --image-path ./images > /tmp/log&" 
}

function predump() {
	ssh $MACHINE_TO "mkdir -p $CONTAINER_ID_TO/../"
	ssh $MACHINE_FROM  <<EOF 
cd $CONTAINER_ID_FROM
echo $(pwd)
mkdir -p images
sudo runc checkpoint --pre-dump --image-path ./images/0
sudo chmod -R 777 ./
tar -cf predump.tar *
echo "Copying pre-dump ..."
scp -r ./predump.tar $MACHINE_TO:$CONTAINER_ID_TO/predump.tar
rm predump.tar
sudo runc checkpoint --image-path ./images/1 --prev-images-dir ./images/0
sudo chmod -R 777 ./
tar -cf dump.tar ./images/1/*
echo "Copying dump ..."
scp -r ./dump.tar $MACHINE_TO:$CONTAINER_ID_TO/dump.tar
rm dump.tar

EOF
echo "Starting migrated container"
ssh $MACHINE_TO "cd $CONTAINER_ID_TO; tar -xf predump.tar; tar -xf dump.tar; nohup sudo runc restore --image-path ./images/1 > /tmp/log&" 
}

if [[ -z "$PRE_DUMP" ]]; then
	dump
else
	predump
fi
