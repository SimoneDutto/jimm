#!/bin/bash

# This script is used to setup a Juju CLI to be authenticated with JIMM without going through login.
# This is particularly useful in headless environments like CI/CD.

set -eux


# c/p parsing semver from https://github.com/cloudflare/semver_bash/blob/master/semver.sh
function semverParseInto() {
    local RE='[^0-9]*\([0-9]*\)[.]\([0-9]*\)[.]\([0-9]*\)\([0-9A-Za-z-]*\)'
    #MAJOR
    eval $2=`echo $1 | sed -e "s#$RE#\1#"`
    #MINOR
    eval $3=`echo $1 | sed -e "s#$RE#\2#"`
    #MINOR
    eval $4=`echo $1 | sed -e "s#$RE#\3#"`
    #SPECIAL
    eval $5=`echo $1 | sed -e "s#$RE#\4#"`
}

function semverLT() {
    local MAJOR_A=0
    local MINOR_A=0
    local PATCH_A=0
    local SPECIAL_A=0

    local MAJOR_B=0
    local MINOR_B=0
    local PATCH_B=0
    local SPECIAL_B=0

    semverParseInto $1 MAJOR_A MINOR_A PATCH_A SPECIAL_A
    semverParseInto $2 MAJOR_B MINOR_B PATCH_B SPECIAL_B

    if [ $MAJOR_A -lt $MAJOR_B ]; then
        return 0
    fi

    if [[ $MAJOR_A -le $MAJOR_B  && $MINOR_A -lt $MINOR_B ]]; then
        return 0
    fi
    
    if [[ $MAJOR_A -le $MAJOR_B  && $MINOR_A -le $MINOR_B && $PATCH_A -lt $PATCH_B ]]; then
        return 0
    fi

    if [[ "_$SPECIAL_A"  == "_" ]] && [[ "_$SPECIAL_B"  == "_" ]] ; then
        return 1
    fi
    if [[ "_$SPECIAL_A"  == "_" ]] && [[ "_$SPECIAL_B"  != "_" ]] ; then
        return 1
    fi
    if [[ "_$SPECIAL_A"  != "_" ]] && [[ "_$SPECIAL_B"  == "_" ]] ; then
        return 0
    fi

    if [[ "_$SPECIAL_A" < "_$SPECIAL_B" ]]; then
        return 0
    fi

    return 1

}

JUJU_VERSION=$(juju version | head -1)
if semverLT "$JUJU_VERSION" "3.0.0"; then
    cat ~/.local/share/juju/accounts.yaml | yq '.controllers += {strenv(JIMM_CONTROLLER_NAME):{"type": "oauth2-device", "user": "jimm-test@canonical.com", "access-token": strenv(JWT)}}' | cat > temp-accounts.yaml && mv temp-accounts.yaml ~/.local/share/juju/accounts.yaml
    cat ~/.local/share/juju/controllers.yaml | yq '.controllers += {strenv(JIMM_CONTROLLER_NAME):{"oidc-login": true, "uuid": "3217dbc9-8ea9-4381-9e97-01eab0b3f6bb", "api-endpoints": ["jimm.localhost:443"]}}' | cat > temp-controllers.yaml && mv temp-controllers.yaml ~/.local/share/juju/controllers.yaml
else
   juju login jimm.localhost
fi


