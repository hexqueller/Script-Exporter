#!/bin/bash
echo 'node_exporter_host_info{type="VM", task="Maria", description="ClusterOfK8s", creater="Secret" } 1'
echo 'node_exporter_host_info{type="VM", task="Sanya", description="ClusterOfK8s", creater="John" } 1'
echo 'node_exporter_host_info{type="VM", task="Daria", description="ClusterOfK8s", creater="Boris" } 1'