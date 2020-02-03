diff --git a/lib/debootstrap.sh b/lib/debootstrap.sh
index 7a992f8..179c0d3 100644
--- a/lib/debootstrap.sh
+++ b/lib/debootstrap.sh
@@ -530,7 +530,7 @@ prepare_partitions()
 update_initramfs()
 {
 	local chroot_target=$1
-	update_initramfs_cmd="update-initramfs -uv -k ${VER}-${LINUXFAMILY}"
+	update_initramfs_cmd="update-initramfs -uv"
 	display_alert "Updating initramfs..." "$update_initramfs_cmd" ""
 	cp /usr/bin/$QEMU_BINARY $chroot_target/usr/bin/
 	mount_chroot "$chroot_target/"
