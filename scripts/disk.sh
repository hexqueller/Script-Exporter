#!/bin/bash

mapfile -t ARRAY_DISK < <(lsblk -dn -o NAME | grep '^sd[a-z]$')

for element in "${ARRAY_DISK[@]}"
do
  size=$(lsblk -dn -o SIZE -b "/dev/$element")
  # echo "disk_size_lsblk{disk=\"$element\"} $size"
done
