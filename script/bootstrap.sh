#!/usr/bin/env bash
#
# bootstrap.sh will check for and install any dependencies we have for building and using core-networking
#
# Usage: ./bootstrap.sh
#


set -e

# Import our helper functions
. script/utils.sh

prettyTitle "Installing iofogctl Dependencies"
echo

# What platform are we on?
OS=$(uname -s | tr A-Z a-z)


#
# All our Go related stuff
#

# Is go installed?
if ! checkForInstallation "go"; then
    echoNotify "\nYou do not have Go installed. Please install and re-run bootstrap."
    exit 1
fi

# Is dep installed?
if ! checkForInstallation "dep"; then
    echoInfo " Attempting to install 'go dep'"
    go get -u github.com/golang/dep/cmd/dep
fi

# Is go-junit-report installed?
if ! checkForInstallation "go-junit-report"; then
    echoInfo " Attempting to install 'go-junit-report'"
    go get -u github.com/jstemmer/go-junit-report
fi
