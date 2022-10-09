#!/bin/bash

abort() {
    sleep 5
    clear
    exit 0
}
trap 'abort' SIGHUP SIGINT SIGTERM

if [[ ${UID} -ne 0 ]]; then
    echo "ERR: script needs to be run as superuser."
    exit 1
fi

DIALOGRC='/opt/shift/config/dialog/.dialogrc'
export DIALOGRC

backtitle="BitBoxBase: Maintenance"
box_h=18
box_w=60

popup_h=14
popup_w=40

#
# check credentials
# -----------------------------------------------------------------------------

# exit if no flashdrive found
if ! /opt/shift/scripts/bbb-cmd.sh flashdrive mount &>/dev/null; then
    echo "MAINTENANCE: no flashdrive found"
    exit
fi
echo "MAINTENANCE: flashdrive mounted to check credentials"

# exit if no token found
if [[ ! -f /mnt/backup/.maintenance-token ]] || [[ ! -f /data/maintenance-token-hashes ]]; then
    echo "MAINTENANCE: flashdrive does not contain a valid maintenance token"
    exit
fi
private_token_hash="$(sha256sum /mnt/backup/.maintenance-token | cut -f 1 -d " ")"
echo "MAINTENANCE: found maintenance token (hash (${private_token_hash})"

# exit if token hash not authorized
if ! grep -q "${private_token_hash}" /data/maintenance-token-hashes; then
    echo "MAINTENANCE: maintenance token not authorized"
    exit
fi

echo "MAINTENANCE: valid token found, starting maintenance menu on tty2"

sleep 5
chvt 2

#
# Submenu: SSD presync
# -----------------------------------------------------------------------------
submenu_presync() {
    while true; do
        if ! menuitem=$(dialog --title "SSD presync" --backtitle "${backtitle}" --menu "\nPlease choose maintenance task" ${box_h} ${box_w} 10 \
                1 "CREATE snapshot on external storage" \
                2 "IMPORT snapshot to internal ssd" \
            3>&1 1>&2 2>&3)
            then break
        fi

        count=0
        unset options

        case $menuitem in
            1)  # create snapshot
                while read -r partition; do
                    count=$((count + 1));
                    onoff="off"
                    if [[ count -eq 1 ]]; then
                        onoff="on"
                    fi
                    options[$count]="${partition} ${onoff}"
                done <<< "$(lsblk -o NAME,SIZE,TYPE -arnp -e 1,7,31,179,252 | grep part | cut -f 1,2 -d " ")"

                options=(${options[@]})
                cmd=(dialog --title "Create snapshot" --backtitle "${backtitle}" --radiolist "Select target drive:" 22 76 16)
                target_drive=$("${cmd[@]}" "${options[@]}" 3>&1 1>&2 2>&3)

                if ! bbb-cmd.sh presync create "${target_drive}"; then
                    echo
                    read -rp "An error occurred. Press [Enter] to continue..."
                else
                    dialog  --title "Snapshot created" --msgbox "\nPresync snapshot created." ${popup_h} ${popup_w}
                fi
                ;;

            2)  # import snapshot
                # select source drive
                while read -r partition; do
                    count=$((count + 1));
                    onoff="off"
                    if [[ count -eq 1 ]]; then
                        onoff="on"
                    fi
                    options[$count]="${partition} ${onoff}"
                done <<< "$(lsblk -o NAME,SIZE,TYPE -arnp -e 1,7,31,179,252 | grep part | cut -f 1,2 -d " ")"

                options=(${options[@]})
                cmd=(dialog --title "Restore snapshot" --backtitle "${backtitle}" --radiolist "Select source drive:" 22 76 16)
                source_drive=$("${cmd[@]}" "${options[@]}" 3>&1 1>&2 2>&3)

                # select file
                count=0
                unset options

                while read -r partition; do
                    count=$((count + 1));
                    onoff="off"
                    if [[ count -eq 1 ]]; then
                        onoff="on"
                    fi
                    options[$count]="${partition} ${onoff}"
                done <<< "$(stat -c "%n %s" /mnt/ext/bbb-presync*)"

                options=(${options[@]})
                cmd=(dialog --title "Restore snapshot" --backtitle "${backtitle}" --radiolist "Select target drive:" 22 76 16)
                source_file=$("${cmd[@]}" "${options[@]}" 3>&1 1>&2 2>&3)

                if ! bbb-cmd.sh presync restore "${source_drive}" "${source_file}"; then
                    echo
                    read -rp "An error occurred. Press [Enter] to continue..."
                else
                    dialog  --title "Snapshot restored" --msgbox "\nPresync snapshot restored." ${popup_h} ${popup_w}
                fi
                ;;
            *)
                break
        esac
    done
}

#
# Submenu: Factory reset
# -----------------------------------------------------------------------------
submenu_reset() {
    while true; do
        if ! menuitem=$(dialog --title "Factory reset" --backtitle "${backtitle}" --menu "\nPlease choose maintenance task" ${box_h} ${box_w} 10 \
                1 "AUTHENTICATION reset" \
                2 "CONFIGURATION reset..." \
                3 "DISK IMAGE reset..." \
            3>&1 1>&2 2>&3)
            then break
        fi

        case $menuitem in
            1)  action_confirmation="The AUTHENTICATION credentials will be reset.";;
            2)  action_confirmation="The CONFIGURATION will be reset to factory defaults.";;
            3)  action_confirmation="The DISK IMAGE on the USB flashdrive will be written to your BitBoxBase.\n\nIt needs to be signed and named 'update.base'.";;
            4)  action_confirmation="The SSD including all data (Bitcoin, Lightning, Electrum) will be wiped permanently.";;
        esac

        if dialog --title "${backtitle}" --yesno "\n${action_confirmation}\n\nContinue?" 12 50; then
            case $menuitem in
                1)  # auth
                    if bbb-cmd.sh reset auth --assume-yes; then
                        dialog  --title "Factory reset" --msgbox "\nOK: Authentication reset.\n\nUse BitBoxApp to set management password again." ${popup_h} ${popup_w}
                    else
                        read -rp "An error occurred. Press [Enter] to continue..."
                    fi
                    ;;

                2)  # config
                    if bbb-cmd.sh reset config --assume-yes; then
                        dialog  --title "Configuration reset" --msgbox "\nOK: Configuration reset to factory defaults." ${popup_h} ${popup_w}
                        while ! bbb-cmd.sh flashdrive check; do
                            dialog  --title "Configuration reset" --msgbox "\nInsert USB flashdrive to store new maintenance token." ${popup_h} ${popup_w}
                        done

                        if bbb-cmd.sh flashdrive mount && bbb-cmd.sh backup sysconfig; then
                            dialog  --title "Configuration reset" --msgbox "\nOK: New maintenance token created on USB flashdrive." ${popup_h} ${popup_w}
                        else
                            read -rp "An error occurred. Press [Enter] to continue..."
                        fi
                        bbb-cmd.sh flashdrive unmount || true

                    else
                        read -rp "An error occurred. Press [Enter] to continue..."
                    fi
                    ;;

                3)  # disk image
                    if bbb-cmd.sh flashdrive mount && bbb-cmd.sh mender-update install flashdrive; then
                        dialog  --title "Disk Image reset" --msgbox "\nOK: Updated from USB disk image, please reboot." ${popup_h} ${popup_w}
                    else
                        read -rp "An error occurred. Press [Enter] to continue..."
                    fi
                    bbb-cmd.sh flashdrive unmount || true
                    ;;
            esac
        fi

    done
}

#
# Main menu
# -----------------------------------------------------------------------------
while true; do

# Main menu
if ! menuitem=$(dialog --title "Main menu" --backtitle "${backtitle}" --menu "\nPlease choose maintenance task" ${box_h} ${box_w} 10 \
        1 "SSD presync data..." \
        2 "Finish factory setup" \
        3 "Factory reset..." \
        4 "Shutdown" \
    3>&1 1>&2 2>&3)
    then abort
fi

case $menuitem in
	1)  # presync
        submenu_presync
        ;;

	2)  # finish factory setup
        if dialog --title "${backtitle}" --yesno "\nThis will delete all user data on the SSD, preparing the device for shipping.\n\nIt also deletes the factory setup credentials.\n\nContinue?" 12 50; then

            bbb-systemctl.sh stop

            # first, cleanup ssd data
            rm -f /mnt/ssd/bitcoin/.bitcoin/*.log
            rm -f /mnt/ssd/bitcoin/.bitcoin/*.dat
            rm -f /mnt/ssd/bitcoin/.bitcoin/.cookie*
            rm -f /mnt/ssd/bitcoin/.bitcoin/.lock
            rm -rf /mnt/ssd/bitcoin/.bitcoin/onion_private_key
            rm -rf /mnt/ssd/bitcoin/.bitcoin/testnet3
            rm -rf /mnt/ssd/bitcoin/.lightning*
            rm -f /mnt/ssd/electrs/db/mainnet/LOG*
            rm -f /mnt/ssd/electrs/db/mainnet/LOCK
            rm -rf /mnt/ssd/lost+found
            rm -rf /mnt/ssd/system
            rm -rf /mnt/ssd/prometheus

            swapoff -a
            rm -f /mnt/ssd/swapfile

            systemctl start redis
            systemctl start bbbmiddleware

            # remove credentials
            if sed -i '/factory token/d' /data/maintenance-token-hashes; then
                dialog  --title "OK" --msgbox "\nFactory setup credentials deleted." ${popup_h} ${popup_w}
            else
                dialog  --title "ERR" --msgbox "\nError: could not delete factory setup credentials." ${popup_h} ${popup_w}
            fi
        fi
        ;;

	3)  # factory reset
        submenu_reset
        ;;

	4)  # shutdown
        if dialog --title "${backtitle}" --yesno "\n    Shut down BitBoxBase?" 8 40; then
            shutdown now
        fi
        ;;
esac

done

# -----------------------------------------------------------------------------

exit 0

clear