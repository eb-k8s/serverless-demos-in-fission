# Adservice

The Adservice provides advertisement based on context keys. If no context keys are provided then it returns random ads.

**NOTE: Do not use fastjson, use jackson!** Because "spring-boot-starter-web" package comes with Jackson's core class library("spring-boot-starter-web" package is used by default for writing fission Java function), and dependency conflicts will occur when importing fastjson(will throw an error: java.lang.NoClassDefFoundError: retrofit2/Converter$Factory).  

**NOTE: Do not include .md file when zipping fission Java function!**
