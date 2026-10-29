#!/bin/sh
# The kirokit-desktop .deb's postinstall: load the app's AppArmor profile
# (/etc/apparmor.d/kirokit-desktop), so WebKit's sandbox may create the user
# namespaces Ubuntu 23.10+ otherwise refuses. On a system without AppArmor
# there is nothing to load and nothing to allow.
set -e

profile=/etc/apparmor.d/kirokit-desktop
if [ -x /usr/sbin/apparmor_parser ] && [ -d /sys/kernel/security/apparmor ] && [ -f "$profile" ]; then
  /usr/sbin/apparmor_parser --replace --write-cache --skip-read-cache "$profile" \
    || echo "kirokit-desktop: could not load $profile; the app may not start until it is loaded (sudo apparmor_parser -r $profile)" >&2
fi
exit 0
