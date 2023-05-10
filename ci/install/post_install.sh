#!/bin/sh


# Generated by: gromit policy
# Generated on: Wed May 10 06:43:07 UTC 2023

# If "True" the install directory ownership will be changed to "tyk:tyk"
change_ownership="False"

# Step 1, decide if we should use systemd or init/upstart
use_systemctl="True"
systemd_version=0
if ! command -V systemctl >/dev/null 2>&1; then
    use_systemctl="False"
else
    systemd_version=$(systemctl --version | head -1 | sed 's/systemd //g')
fi

cleanup() {
    # After installing, remove files that were not needed on this platform / system
    if [ "${use_systemctl}" = "True" ]; then
        rm -f /lib/systemd/system/tyk-gateway.service
    else
        rm -f /etc/init.d/tyk-gateway
    fi
}

restoreServices() {
    if [ "${use_systemctl}" = "True" ]; then
        if [ ! -f /lib/systemd/system/tyk-gateway.service ]; then
            cp /opt/tyk-gateway/install/inits/systemd/system/tyk-gateway.service /lib/systemd/system/tyk-gateway.service
        fi
    else
        if [ ! -f /etc/init.d/tyk-gateway ]; then
            cp /opt/tyk-gateway/install/inits/sysv/init.d/tyk-gateway /etc/init.d/tyk-gateway
        fi
    fi
}

setupOwnership() {
    printf "\033[32m Post Install of the install directory ownership and permissions\033[0m\n"
    [ "${change_ownership}" = "True" ] && chown -R tyk:tyk /opt/tyk-gateway
    # Config file should never be world-readable
    chmod 660 /opt/tyk-gateway/tyk.conf
}

cleanInstall() {
    printf "\033[32m Post Install of an clean install\033[0m\n"
    # Step 3 (clean install), enable the service in the proper way for this platform
    if [ "${use_systemctl}" = "False" ]; then
        if command -V chkconfig >/dev/null 2>&1; then
            chkconfig --add tyk-gateway
            chkconfig tyk-gateway on
        fi
        if command -V update-rc.d >/dev/null 2>&1; then
            update-rc.d tyk-gateway defaults
        fi

        service tyk-gateway restart ||:
    else
        printf "\033[32m Reload the service unit from disk\033[0m\n"
        systemctl daemon-reload ||:
        printf "\033[32m Unmask the service\033[0m\n"
        systemctl unmask tyk-gateway ||:
        printf "\033[32m Set the preset flag for the service unit\033[0m\n"
        systemctl preset tyk-gateway ||:
        printf "\033[32m Set the enabled flag for the service unit\033[0m\n"
        systemctl enable tyk-gateway ||:
        systemctl restart tyk-gateway ||:
    fi
}

upgrade() {
    printf "\033[32m Post Install of an upgrade\033[0m\n"
    if [ "${use_systemctl}" = "False" ]; then
        service tyk-gateway restart
    else
        systemctl daemon-reload ||:
        systemctl restart tyk-gateway ||:
    fi
}

# Step 2, check if this is a clean install or an upgrade
action="$1"
if  [ "$1" = "configure" ] && [ -z "$2" ]; then
    # Alpine linux does not pass args, and deb passes $1=configure
    action="install"
elif [ "$1" = "configure" ] && [ -n "$2" ]; then
    # deb passes $1=configure $2=<current version>
    action="upgrade"
fi

case "$action" in
    "1" | "install")
        setupOwnership
        cleanInstall
        ;;
    "2" | "upgrade")
        printf "\033[32m Post Install of an upgrade\033[0m\n"
        setupOwnership
        restoreServices
        upgrade
        ;;
    *)
        # $1 == version being installed
        printf "\033[32m Alpine\033[0m"
        setupOwnership
        cleanInstall
        ;;
esac

# From https://www.debian.org/doc/debian-policy/ap-flowcharts.html and
# https://docs.fedoraproject.org/en-US/packaging-guidelines/Scriptlets/ it appears that cleanup is not
# needed to support systemd and sysv

#cleanup
