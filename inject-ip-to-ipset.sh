#!/bin/bash

STOML="/home/sem/bin/stoml"
CONFIG="/home/sem/bin/configs/botnets-ipset.toml"

TOKEN=`$STOML $CONFIG Token`

SERVERS=`$STOML $CONFIG server_addr`

checkip() {
	local ip=$1
	local IFS=.; local -a a=($ip)
	[[ $ip =~ ^[0-9]+(\.[0-9]+){3}$ ]] || return 1
	local quad
	for quad in {0..3}; do
		[[ "${a[$quad]}" -gt 255 ]] && return 1
	done
	return 0
}

if ! checkip "$1"; then
	echo "IP format error: $1"
	exit 1
fi

for i in $SERVERS; do
	h=`echo $i|sed -e 's/:[[:digit:]]*$//'`
	#echo /usr/bin/curl -k -G https://$i/api/addip?ip=$1 -H "Token: $TOKEN"
	/usr/bin/curl -k -G https://$i/api/addip?ip=$1 -H "Token: $TOKEN"
done
