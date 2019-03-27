#!/bin/bash


if [[ "$TRAVIS_OS_NAME" == "osx" ]]; then
    sudo systemsetup -f -setremotelogin on || exit $?
elif [[ "$TRAVIS_OS_NAME" == "linux" ]]; then
    sudo apt-get -qq install openssh-client openssh-server || exit $?
    sudo service ssh restart || exit $?
fi

mkdir -p ~/.ssh || exit $?

ssh-keygen -q -t rsa -b 4096 -N "" -f ~/.ssh/id_rsa || exit $?

cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys || exit $?

ssh-keyscan -t rsa localhost >> ~/.ssh/known_hosts || exit $?
