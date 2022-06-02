# Adservice

**[Deprecated] This Java function consumes too much memory and CPU due to the springboot in fission/jvm-env.**  
Please use the go version function: [adservice-use-go](../adservice).

The Adservice provides advertisement based on context keys. If no context keys are provided then it returns random ads.

Use this command to zip fission Java function to **let fission-jvm-builder build**: 
```
zip adservice.zip -r src/ pom.xml
``` 
**Do not include other files or folders when zipping fission Java function!**

Or If you have JDK8 and Maven installed, you can **locally build** the JAR file using command:
```
mvn clean package
``` 

And use target/xxx-with-dependencies.jar to directly create fission Java function, this can prevent mistakes in fission-jvm-builder:
```
fission fn create --name xxx --deploy target/xxx-with-dependencies.jar --env jvm --entrypoint xxx.xxx
```

**NOTE: Do not use fastjson, use jackson!** Because "spring-boot-starter-web" package comes with Jackson's core class library("spring-boot-starter-web" package is used by default for writing fission Java function), and dependency conflicts will occur when importing fastjson(will throw an error: java.lang.NoClassDefFoundError: retrofit2/Converter$Factory).  
