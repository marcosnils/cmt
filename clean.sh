#!/bin/bash

sudo rm dump.tar.gz
ssh 172.31.60.136 <<EOF 
cd lala
sudo runc kill
sudo rm -fr /var/run/opencontainer/
sudo rm -fr images
sudo rm dump.tar.gz
sudo sh -c 'runc start &'
echo "DONE"
EOF
