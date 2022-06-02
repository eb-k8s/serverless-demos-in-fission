#!/bin/bash
allcpuoffunction=`kubectl top pods -n fission-function | grep gcpdemo | awk '{print $2}' | tail -n +2 | egrep -v "^$" | sed 's/[a-zA-Z]//g' | paste -sd+ | bc`
allcpuofdeployment=`kubectl top pods -n gcpdemo | awk '{print $2}' | tail -n +2 | egrep -v "^$" | sed 's/[a-zA-Z]//g' | paste -sd+ | bc`
allcpu=`echo "$allcpuoffunction+$allcpuofdeployment" | bc`
printf "CPU consumption(gcp-microservices-demo): %dm (function %dm + k8s-deployment %dm)\n" $allcpu $allcpuoffunction $allcpuofdeployment
allmemoryoffunction=`kubectl top pods -n fission-function | grep gcpdemo | awk '{print $3}' | tail -n +2 | egrep -v "^$" | sed 's/[a-zA-Z]//g' | paste -sd+ | bc`
allmemoryofdeployment=`kubectl top pods -n gcpdemo | awk '{print $3}' | tail -n +2 | egrep -v "^$" | sed 's/[a-zA-Z]//g' | paste -sd+ | bc`
allmemory=`echo "$allmemoryoffunction+$allmemoryofdeployment" | bc`
printf "Memory consumption(gcp-microservices-demo): %dMi (function %dMi + k8s-deployment %dMi)\n" $allmemory $allmemoryoffunction $allmemoryofdeployment
