#!/bin/bash

set -e

MODE="$1"
ENV_FILE="$MODE.env"
SECRETS_FILE=".kamal/secrets.$MODE"
DEPLOY_FILE="config/deploy.$MODE.yml"
DEPLOY_FILE_TEMPLATE="config/deploy.template.yml"
SSH_FILE="config/ssh.$MODE.yml"
ENV_SECRETS_TMP_FILE="$(mktemp)"
SERVER_IPS_TMP_FILE="$(mktemp)"

ENV_SECRETS_STR_REPL="ENV_LIST_SECRETS"
SERVER_IPS_STR_REPL="SERVER_IP_ADDRESSES"

SERVER_IP_ENV_VAR_PREFIX="YEETFILE_SERVER_IP"

function replace_in_file() {
    TARGET="$1"
    REPLACEMENT="$2"
    FILE="$3"

    TMP_FILE=$(mktemp)
    echo "$REPLACEMENT" > "$TMP_FILE"

    sed -e "/$TARGET/{
        r $TMP_FILE
        d
    }" "$FILE" > tmpfile && mv tmpfile "$FILE"

    rm "$TMP_FILE"
}

if [ -z "$MODE" ]; then
    echo "Missing first arg, should be 'dev', 'prod', etc."
    exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
    echo "Env file $ENV_FILE does not exist"
    exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

mkdir -p .kamal
rm -f "$SECRETS_FILE"
rm -f "$ENV_SECRETS_TMP_FILE"
rm -f "$SERVER_IPS_TMP_FILE"
cp "$DEPLOY_FILE_TEMPLATE" "$DEPLOY_FILE"

HAS_SERVER_IPS=0

while IFS='=' read -r var_name var_value; do
    # Check if the line is a comment or blank
    if [[ -n "$var_name" ]] && [[ ! "$var_name" == "#"* ]]; then
        if [[ $var_name == "$SERVER_IP_ENV_VAR_PREFIX"* ]]; then
            # Var is a server IP, add to the IPs list
            echo "  - $var_value" >> "$SERVER_IPS_TMP_FILE"
            HAS_SERVER_IPS=1
        elif [[ $var_name == *"SSH"* ]]; then
            # Replace SSH values in the deploy file directly
            # (Kamal doesn't allow secrets in that section)
            var_value=$(echo "${!var_name}" | sed 's/\//\\\//g')
            sed -i "s/$var_name/$var_value/g" "$DEPLOY_FILE"
        elif [[ "$var_name" == *"KAMAL"* ]] || [[ "$var_name" == *"YEETFILE"* ]]; then
            # Add to env secrets list
            echo "$var_name=\$$var_name" >> "$SECRETS_FILE"
            if [[ ! $var_name == *"KAMAL"* ]]; then
                echo "    - $var_name" >> "$ENV_SECRETS_TMP_FILE"
            fi
        fi
    fi
done < "$ENV_FILE"

if [[ $HAS_SERVER_IPS -eq 0 ]]; then
    echo "Server IPs not found!"
    echo "These must be set in $ENV_FILE with" $SERVER_IP_ENV_VAR_PREFIX"_[N]=XX.XXX...."
    echo "i.e." $SERVER_IP_ENV_VAR_PREFIX"_1=10.140.11.26"
    exit 1
fi

# Update deploy file with contents
ENV_SECRETS=$(<"$ENV_SECRETS_TMP_FILE")
SERVER_IPS=$(<"$SERVER_IPS_TMP_FILE")

replace_in_file "$ENV_SECRETS_STR_REPL" "$ENV_SECRETS" "$DEPLOY_FILE"
replace_in_file "$SERVER_IPS_STR_REPL" "$SERVER_IPS" "$DEPLOY_FILE"

rm -f "$ENV_SECRETS_TMP_FILE"
rm -f "$SERVER_IPS_TMP_FILE"

if [[ -f $SSH_FILE ]]; then
    cat $SSH_FILE >> $DEPLOY_FILE
fi

read -r -p "Show secrets? (y/N): " input && [[ "$input" == "y" ]] && \
  kamal secrets print -d "$MODE"
read -r -p "Ready to deploy? (y/N): " input && [[ "$input" == "y" ]] && \
  kamal deploy -d "$MODE"
