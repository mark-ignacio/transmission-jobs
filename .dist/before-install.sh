#!/bin/sh
# mostly from https://fedoraproject.org/wiki/Packaging:UsersAndGroups
getent group transmission-jobs >/dev/null || groupadd -r transmission-jobs
getent passwd transmission-jobs >/dev/null || \
    useradd -r -g transmission-jobs -s /sbin/nologin \
    -c "transmission-jobs run service" transmission-jobs
exit 0
